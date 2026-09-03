# Case Study 07: Batch Processing

Batch Processing giúp học concurrency, cancellation và boundary trong job dài.

Trọng tâm:

- Worker pool.
- Context cancellation.
- Chunking.
- Retry từng item.
- Partial failure.
- Observability cho progress.

Kết luận chính: orchestration batch thuộc application/job adapter; domain vẫn chỉ giữ rule nghiệp vụ.
