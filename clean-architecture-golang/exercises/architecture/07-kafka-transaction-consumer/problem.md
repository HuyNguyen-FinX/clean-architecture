# Bài 7: Kafka Consumer Xử Lý Transaction

## Requirement

Consumer đọc Kafka event `TransferRequested` và tạo banking transfer.

## Nhiệm vụ

1. Xác định consumer thuộc layer nào.
2. Map event payload sang command.
3. Thiết kế idempotency.
4. Thiết kế retry và DLQ.
5. Đảm bảo offset commit đúng thời điểm.

## Câu hỏi

- Kafka message có phải use case command không?
- Nếu message bị xử lý hai lần thì sao?
- Commit offset trước hay sau khi transaction DB commit?
