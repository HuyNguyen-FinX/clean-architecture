# 21 Concurrency Golang

## Tại sao cần học?

Concurrency là năng lực mạnh của Go nhưng cũng dễ phá architecture nếu goroutine, channel và shared state bị dùng không kiểm soát.

## Câu hỏi chính

Concurrency thuộc domain hay application?

Phần lớn orchestration concurrency thuộc application hoặc infrastructure. Domain có thể có rule liên quan đến consistency, nhưng không nên tự spawn goroutine để gọi database hoặc Kafka.

## Nội dung trọng tâm

- Goroutine.
- Channel.
- Mutex.
- Worker pool.
- `errgroup`.
- Context cancellation.
- Backpressure.

## Anti-pattern

- Goroutine không gắn context.
- Channel xuyên qua layer như API tùy tiện.
- Shared mutable domain object giữa goroutine.
- Worker chứa business logic và infrastructure trộn lẫn.

## Mastery Check

- [ ] Tôi biết dùng context cancellation trong flow concurrent.
- [ ] Tôi biết tách batch orchestration khỏi domain rule.
- [ ] Tôi biết race condition có thể phá invariant.
