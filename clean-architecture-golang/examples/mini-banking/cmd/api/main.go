package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	httpadapter "github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/delivery/http"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
	kafkaadapter "github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/infrastructure/kafka"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/infrastructure/memory"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/infrastructure/postgres"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/platform/httpmiddleware"
	runtimeadapter "github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/platform/runtime"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("mini banking stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	runCtx, cancelRun := context.WithCancel(signalCtx)
	defer cancelRun()

	deps, err := buildDependencies(runCtx)
	if err != nil {
		return err
	}
	defer deps.close()

	clock := runtimeadapter.Clock{}
	transferMoney := application.NewTransferMoneyUseCase(
		deps.store,
		deps.transactor,
		runtimeadapter.IDs{},
		clock,
	)
	history := application.NewListTransfersUseCase(deps.store)
	handler := httpadapter.NewHandler(transferMoney, history)
	metrics := httpmiddleware.NewMetrics()
	root := http.NewServeMux()
	root.Handle("GET /metrics", metrics)
	root.Handle("/", httpmiddleware.ObserveRequestsWithMetrics(logger, metrics, handler.Routes()))
	server := &http.Server{
		Addr:              ":8080",
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	componentErrors := make(chan error, 3)
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		logger.Info("mini banking API listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			componentErrors <- fmt.Errorf("serve HTTP: %w", err)
		}
	}()

	closeKafka := startKafkaComponents(
		runCtx,
		&workers,
		componentErrors,
		logger,
		deps.outbox,
		transferMoney,
		clock,
	)

	var runErr error
	select {
	case runErr = <-componentErrors:
		logger.Error("component failed", "error", runErr)
	case <-signalCtx.Done():
		logger.Info("shutdown requested")
	}
	cancelRun()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("shutdown HTTP: %w", err))
	}
	workers.Wait()
	if err := closeKafka(); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("close Kafka: %w", err))
	}
	return runErr
}

type dependencies struct {
	store      application.TransferStore
	outbox     application.OutboxStore
	transactor application.Transactor
	close      func()
}

func buildDependencies(ctx context.Context) (dependencies, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		repo := memory.NewRepository(
			mustAccount("A-100", 1_000_000, 0, "VND"),
			mustAccount("B-200", 250_000, 0, "VND"),
		)
		return dependencies{
			store: repo, outbox: repo, transactor: repo, close: func() {},
		}, nil
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return dependencies{}, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return dependencies{}, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	if os.Getenv("AUTO_MIGRATE") == "1" {
		if err := postgres.Migrate(ctx, pool); err != nil {
			pool.Close()
			return dependencies{}, err
		}
	}
	repo := postgres.NewRepository(pool)
	return dependencies{
		store: repo, outbox: repo, transactor: postgres.NewTransactor(pool), close: pool.Close,
	}, nil
}

func startKafkaComponents(
	ctx context.Context,
	workers *sync.WaitGroup,
	componentErrors chan<- error,
	logger *slog.Logger,
	outbox application.OutboxStore,
	transfer *application.TransferMoneyUseCase,
	clock application.Clock,
) func() error {
	brokers := splitNonEmpty(os.Getenv("KAFKA_BROKERS"))
	if len(brokers) == 0 {
		logger.Info("Kafka disabled; outbox messages remain pending until a relay is configured")
		return func() error { return nil }
	}

	publisher := kafkaadapter.NewPublisher(brokers)
	relay := application.NewRelayOutboxUseCase(outbox, publisher, clock)
	workers.Add(1)
	go func() {
		defer workers.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			if _, err := relay.RunOnce(ctx, 100); err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("outbox relay iteration failed", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	var consumer *kafkaadapter.Consumer
	if os.Getenv("KAFKA_CONSUMER_ENABLED") == "1" {
		consumer = kafkaadapter.NewConsumer(
			brokers,
			"transfer-requests-v1",
			"mini-banking-transfer-v1",
			"transfer-requests-dlq-v1",
			transfer,
		)
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				componentErrors <- fmt.Errorf("run Kafka consumer: %w", err)
			}
		}()
	}

	return func() error {
		var err error
		if consumer != nil {
			err = consumer.Close()
		}
		return errors.Join(err, publisher.Close())
	}
}

func splitNonEmpty(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func mustAccount(id string, balanceAmount int64, overdraftAmount int64, currency string) *domain.Account {
	accountID, err := domain.NewAccountID(id)
	if err != nil {
		panic(err)
	}
	balance, err := domain.NewMoney(balanceAmount, currency)
	if err != nil {
		panic(err)
	}
	overdraftLimit, err := domain.NewMoney(overdraftAmount, currency)
	if err != nil {
		panic(err)
	}
	account, err := domain.NewAccount(accountID, balance, overdraftLimit)
	if err != nil {
		panic(err)
	}
	return account
}
