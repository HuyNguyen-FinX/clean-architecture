# Concurrency trong Go: ownership, cancellation và backpressure

Goroutine rẻ nhưng không miễn phí; channel không tự tạo architecture; mutex không bảo vệ nhiều process. Concurrency cần boundary ownership, cancellation và bounded work.

## Kết quả học tập

- quyết định khi nào parallelism có lợi;
- quản goroutine/channel ownership;
- dùng context/errgroup/worker pool;
- thiết kế backpressure và shutdown;
- phân biệt memory race, logical race và DB concurrency;
- test không dựa sleep.

## 1. Ba level

### Level 1

Goroutine chạy đồng thời. Nếu nhiều goroutine chia sẻ state, phải phối hợp.

### Level 2

Engineer quản WaitGroup/channel/mutex/context, error propagation, limits và leak.

### Level 3

Concurrency là application/infrastructure orchestration. Domain giữ invariant; consistency across requests/processes cần database/distributed protocol, không chỉ mutex.

## 2. Đừng spawn rồi quên

Sai:

~~~go
go publisher.Publish(context.Background(), event)
return nil
~~~

Request đã báo success nhưng publish có thể fail; process shutdown mất goroutine; không retry/observe.

Đúng tùy guarantee:

- synchronous wait;
- durable outbox;
- owned worker queue có shutdown;
- best-effort được công bố rõ.

## 3. Structured concurrency với errgroup

~~~go
group, ctx := errgroup.WithContext(ctx)
group.Go(func() error { return loadProfile(ctx) })
group.Go(func() error { return loadLimits(ctx) })
if err := group.Wait(); err != nil {
	return err
}
~~~

Tasks cùng lifecycle; một error cancel siblings. Chỉ parallelize independent I/O. Nếu hai DB query phải cùng transaction/lock order, concurrent query trên một pgx.Tx có thể không supported/hữu ích.

## 4. Channel ownership

Rule:

- producer sở hữu việc close channel;
- receiver không close channel không tạo;
- gửi phải tôn trọng cancellation;
- buffer size là capacity policy;
- nil/closed channel semantics phải hiểu.

~~~go
select {
case jobs <- job:
	return nil
case <-ctx.Done():
	return ctx.Err()
}
~~~

Không expose channel xuyên layers nếu lifecycle/close ownership không rõ. Iterator/callback có thể đơn giản hơn.

## 5. Worker pool bounded

~~~go
func RunWorkers(ctx context.Context, workers int, jobs <-chan Job, handle func(context.Context, Job) error) error {
	group, ctx := errgroup.WithContext(ctx)
	for i := 0; i < workers; i++ {
		group.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case job, ok := <-jobs:
					if !ok {
						return nil
					}
					if err := handle(ctx, job); err != nil {
						return err
					}
				}
			}
		})
	}
	return group.Wait()
}
~~~

Snippet conceptual cần policy: một job fail có dừng pool không, retry/DLQ, drain hay cancel, panic recovery.

## 6. Backpressure

Unbounded queue chuyển overload thành memory/latency collapse. Bounded queue khi đầy:

- block producer trong deadline;
- reject/load shed;
- drop best-effort;
- spill durable queue;
- scale consumers.

Policy phụ thuộc business guarantee. Không drop transfer/outbox silently.

## 7. Mutex và ownership

~~~go
type Repository struct {
	mu sync.RWMutex
	accounts map[AccountID]*Account
}
~~~

Clone object dưới lock để caller không mutate shared pointer sau unlock. Lock scope ngắn, không gọi network while held.

Mutex process-local. Nhiều replicas cùng update Postgres cần row lock/version/atomic SQL.

## 8. Memory race vs logical race

Data race: unsynchronized memory access, race detector có thể bắt.

Logical race:

~~~text
check balance
other request changes balance
save based on stale read
~~~

Mỗi memory access có mutex vẫn có thể sai nếu check+act không atomic. Database transaction/version giải quyết scope durable.

## 9. Atomic package

sync/atomic phù hợp counters/flags đơn giản với memory ordering hiểu rõ. Không dùng nhiều atomics để dựng state machine phức tạp thay mutex/channel.

Metrics counters thường SDK quản; business balance không dùng atomic int64 xuyên process.

## 10. Context

Context first parameter, không lưu lâu trong struct. Derived goroutine phải kết thúc khi parent cancel. Không dùng context.Value như generic dependency bag.

Domain computation thuần không cần context. Repository/external calls cần.

## 11. Goroutine leak

Ví dụ:

~~~go
result := make(chan Result)
go func() { result <- slowCall() }()
select {
case <-ctx.Done():
	return ctx.Err()
case value := <-result:
	return value
}
~~~

Nếu caller cancel, goroutine block gửi mãi trên unbuffered channel. Dùng buffer 1 hoặc select send với ctx, và slowCall cũng phải cancel được.

## 12. Fan-out/fan-in

Parallel query 1000 accounts không spawn 1000 uncontrolled goroutines. Semaphore/worker limit theo pool/provider quota:

~~~go
group.SetLimit(16)
~~~

Concurrency > DB pool chỉ tạo wait. Measure.

## 13. Error aggregation

Fail-fast hay collect all? Batch validation có thể collect; transaction workflow thường fail-fast. errors.Join cho multiple independent failures, nhưng cancellation errors có thể che primary cause, cần policy.

## 14. Ordering

Goroutine completion không theo start order. Nếu output cần order, carry index/sequence và sort/assemble. Kafka order theo partition không được khôi phục bằng goroutine tùy ý; parallel per key cần partitioned workers.

## 15. Shutdown

Worker lifecycle:

1. stop accept/fetch;
2. close producer-owned jobs hoặc cancel;
3. drain/finish theo timeout;
4. commit only completed;
5. close clients;
6. report unfinished.

WaitGroup Add phải xảy ra trước goroutine; không copy mutex.

## 16. Transaction concurrency

Mini-banking lock Account theo sorted ID. Không chạy concurrent queries trên cùng transaction để “nhanh hơn” khi driver serializes/prohibits. Transaction càng dài, lock/pool contention càng lớn.

## 17. Production scenario

Kafka lag tăng, team nâng workers 10 → 1000:

- DB pool 50, 950 goroutines chờ;
- transaction timeouts/retries;
- provider quota 429;
- memory queue tăng;
- rebalance drain chậm.

Tune end-to-end bottleneck, bounded concurrency và backpressure. Worker count không phải throughput knob độc lập.

## 18. Testing

- go test -race;
- barrier channels ép interleaving;
- fake clock cho retry;
- goleak hoặc ownership assertions;
- timeout test tránh hang;
- invariant/final state;
- stress nhiều seeds;
- real DB/broker cho distributed semantics.

Không dùng time.Sleep để “đợi goroutine chắc chạy”.

## 19. Debug

1. goroutine profile;
2. block/mutex profile;
3. queue depth/age;
4. pool wait;
5. context cancellation;
6. channel owner/close path;
7. race output;
8. thread dump during shutdown.

## 20. Khi nào không dùng concurrency?

CPU/I/O nhỏ, operation dependent, transaction cùng connection, workload thấp: sequential code dễ đúng/debug hơn. Parallelism có overhead và failure combination. Measure before/after.

## 21. Bài tập

1. Sửa goroutine leak.
2. Implement bounded worker pool có drain.
3. Tái hiện logical lost update dù race detector sạch.
4. Partition jobs theo AccountID giữ order.
5. Thiết kế overload policy.

## 22. Mastery questions

1. Mutex vì sao không bảo vệ nhiều replicas?
2. Buffer size là architecture policy gì?
3. Race detector bỏ logical race nào?
4. Ai close channel?
5. Goroutine fire-and-forget làm mất guarantee?
6. errgroup cancellation có trade-off gì?
7. Worker count liên hệ DB pool?
8. Khi nào sequential tốt hơn?

## Further reading

- Go memory model.
- Go race detector, context, sync và x/sync/errgroup docs.
- Go blog Pipelines and cancellation.
- Concurrency Is Not Parallelism talk.

## Quality gate

- [x] Goroutine/channel/mutex/context ownership
- [x] errgroup/worker/backpressure
- [x] races/DB concurrency/order
- [x] shutdown/leaks/errors/tests
- [x] production/debug/trade-off
- [x] exercises/mastery
