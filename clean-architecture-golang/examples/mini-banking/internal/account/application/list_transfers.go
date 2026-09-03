package application

import (
	"context"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

type TransferView struct {
	ID            string
	FromAccountID string
	ToAccountID   string
	Amount        int64
	Currency      string
	CreatedAt     time.Time
}

type ListTransfersUseCase struct {
	transfers TransferRepository
}

func NewListTransfersUseCase(transfers TransferRepository) *ListTransfersUseCase {
	if transfers == nil {
		panic("application: nil transfer repository")
	}
	return &ListTransfersUseCase{transfers: transfers}
}

func (uc *ListTransfersUseCase) Execute(
	ctx context.Context,
	rawAccountID string,
	limit int,
) ([]TransferView, error) {
	accountID, err := domain.NewAccountID(rawAccountID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	transfers, err := uc.transfers.ListTransfersByAccount(ctx, accountID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]TransferView, 0, len(transfers))
	for _, transfer := range transfers {
		result = append(result, TransferView{
			ID:            string(transfer.ID()),
			FromAccountID: string(transfer.From()),
			ToAccountID:   string(transfer.To()),
			Amount:        transfer.Amount().Amount(),
			Currency:      transfer.Amount().Currency().String(),
			CreatedAt:     transfer.CreatedAt(),
		})
	}
	return result, nil
}
