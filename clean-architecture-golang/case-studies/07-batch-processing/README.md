# Case Study 07: Batch Processing - Backpressure, Checkpoint Và Partial Failure

Batch job thường bắt đầu bằng một vòng `for` rồi lớn dần thành workflow nhiều giờ. Clean boundary giúp business operation test được; correctness của batch còn phụ thuộc ownership goroutine, checkpoint, retry và khả năng resume.

## Scenario

Mỗi đêm hệ thống tính phí cho 10 triệu Account từ snapshot cuối ngày. Mỗi Account độc lập về tính phí, nhưng output phải gắn với `business_date` và policy version. Job có deadline bốn giờ, chạy nhiều replicas và có thể bị SIGTERM khi deploy.

Yêu cầu:

- Không tính phí hai lần khi restart.
- Một record lỗi không làm mất progress của toàn job.
- Không load 10 triệu rows vào memory.
- Database chậm phải tạo backpressure, không tăng goroutine vô hạn.
- Operator xem được progress, lỗi và replay có kiểm soát.

## Tách Ba Loại Logic

Domain policy:

~~~go
func CalculateFee(snapshot AccountSnapshot, policy FeePolicy) (Fee, error)
~~~

Application operation `ChargeAccount` hoặc `CalculateAndPostFee` quản lý idempotency, load/save và transaction của một Account. Batch adapter quản lý scan, chunk, scheduling, concurrency, checkpoint và shutdown.

Nếu nhét worker pool vào domain service, invariant bị trộn với execution policy. Nếu nhét fee formula vào SQL vì tiện, policy version/test/audit khó kiểm soát. Có thể dùng set-based SQL khi rule đơn giản và performance cần thiết, nhưng đó là quyết định explicit với integration/property tests, không phải mặc định.

## Dependency

~~~mermaid
flowchart LR
    SCHED["Scheduler/CLI"] --> RUN["Batch runner adapter"]
    RUN --> SOURCE["ItemSource port"]
    RUN --> UC["ProcessAccount use case"]
    UC --> POLICY["Fee policy"]
    UC --> STORE["Posting UnitOfWork"]
    PG["PostgreSQL adapters"] -.implements.-> SOURCE
    PG -.implements.-> STORE
~~~

Batch runner là driving adapter/application orchestration tùy ownership project. Nó không được làm domain depend vào cron library. `main` wire worker count, deadlines và clients.

## Idempotency Key Và Work Identity

Business operation ID là `(job_type, business_date, account_id, policy_version)`, không phải index trong slice. Posting table có unique constraint trên key đó. Retry/restart cùng identity trả kết quả cũ; đổi policy version phải là operation mới có migration/reversal rule rõ.

Checkpoint chỉ nói scanner đã đi tới đâu; nó không chứng minh từng side effect đã commit. Correctness dựa vào per-item idempotency. Nếu commit checkpoint 10.000 nhưng item 9.999 chưa bền, resume có thể mất dữ liệu.

## Streaming Và Chunking

Source dùng keyset pagination:

~~~sql
SELECT account_id
FROM accounts
WHERE account_id > $1
ORDER BY account_id
LIMIT $2;
~~~

Offset pagination chậm dần và không ổn định khi dữ liệu thay đổi. Snapshot semantics phải explicit: transaction snapshot dài, materialized candidate table, hoặc filter theo cutoff/version. Không hứa "end-of-day" nếu scan thấy dữ liệu thay đổi giữa chừng.

Chunk size là trade-off round trip, memory, lock time và retry scope. Không bọc toàn triệu item trong một DB transaction; rollback/locks/WAL sẽ quá lớn. Thường mỗi item hoặc small chunk là transaction độc lập theo atomicity business.

## Worker Pool Có Ownership

~~~go
func Run(ctx context.Context, workers int, items <-chan Item, handle func(context.Context, Item) error) error {
	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i < workers; i++ {
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case item, ok := <-items:
					if !ok { return nil }
					if err := handle(ctx, item); err != nil { return err }
				}
			}
		})
	}
	return g.Wait()
}
~~~

Đây là skeleton minh họa fail-fast. Production runner thường gửi per-item failure sang result collector thay vì cancel toàn batch. Quan trọng là channel bounded, producer respect context, goroutine có owner và `Wait`.

## Backpressure Và Adaptive Concurrency

Queue nội bộ bounded làm scanner chậm khi workers chậm. Worker count phải liên hệ DB pool và dependency quotas: 100 workers với pool 20 chỉ tạo wait; 100 external calls có thể vượt rate limit. Có thể giảm concurrency khi latency/error tăng, nhưng control loop cần giới hạn để tránh oscillation.

Không retry ngay trong hàng trăm workers với cùng interval. Backoff có jitter, retry budget và global rate limiter bảo vệ dependency.

## Partial Failure Policy

Phân loại:

- Invalid business data: ghi rejected result, tiếp tục; cần remediation queue.
- Transient DB/provider: retry bounded; sau đó mark retryable failure.
- Systemic error như schema mismatch: dừng job sớm để không tạo hàng triệu lỗi.
- Context deadline/shutdown: ngừng nhận item, cho in-flight hoàn thành trong grace period.

Job final status nên là `Succeeded`, `SucceededWithErrors`, `Failed` hoặc `Cancelled`, kèm counts. "Process exit 0" không đủ nếu có 20.000 record lỗi.

## Multi-Replica Claiming

Candidate table có trạng thái pending/processing/done và lease owner/expiry. Workers claim batch bằng transaction, thường với `FOR UPDATE SKIP LOCKED`. Lease hết hạn cho phép reclaim sau crash. Completion vẫn phải compare owner/version để worker cũ không ghi đè worker mới.

Lease không thay per-item idempotency: pause GC/network có thể làm worker cũ tỉnh lại sau khi lease bị reclaim.

## Failure Matrix

| Failure | Điều phải giữ |
|---|---|
| Process chết sau posting trước result mark | unique operation làm replay no-op |
| Process chết sau claim | lease expiry cho reclaim |
| Source scan timeout | cursor cuối đã xác nhận, query lại |
| Một item poison | quarantine, không retry vô hạn |
| DB outage toàn vùng | pause/fail job, tránh retry storm |
| SIGTERM | stop source, drain bounded, persist result/checkpoint |
| Policy deploy giữa job | pin policy version từ lúc tạo job |

## Testing Strategy

- Pure domain/property tests cho fee boundaries và sum.
- Use-case tests idempotent replay, posting rollback và policy version.
- Runner tests với fake clock/source: bounded concurrency, cancellation, không leak goroutine.
- Integration tests claim lease, `SKIP LOCKED`, unique operation và crash recovery.
- Load test memory, DB pool wait, throughput và skew theo Account size.
- Resume test kill process ngẫu nhiên nhiều lần rồi so output với oracle sequential.

## Observability

Metrics: total/discovered/claimed/succeeded/rejected/retryable, throughput, remaining estimate, oldest lease, queue utilization, dependency latency. `account_id` không là metric label. Progress được persist để dashboard không mất khi process restart.

Log mỗi lỗi cần job ID, operation ID, classification và attempt; không log success từng item ở 10 triệu records. Trace sampling có chủ đích cho error/slow item.

## Trade-off Và Alternatives

Set-based SQL có thể nhanh và atomic hơn worker pool cho phép tính thuần dữ liệu, nhưng coupling policy-schema cao và rollback lớn. Stream processor phù hợp continuous events hơn nightly snapshot. Managed workflow engine đáng giá với workflow dài/nhiều dependency; với một scan đơn giản nó có thể quá nặng.

## Câu Hỏi Mastery

1. Checkpoint khác idempotency như thế nào?
2. Vì sao lease không loại bỏ duplicate processing?
3. Worker count nên bị giới hạn bởi tài nguyên nào?
4. Khi nào một item lỗi phải dừng toàn job?
5. Làm sao chứng minh job restart không double charge?

## Bài Thực Hành

Implement runner với channel capacity 2, ba workers và fake handler đo concurrency. Viết test cancel giữa batch, restart từ candidate table và assert mỗi business operation chỉ có một posting.
