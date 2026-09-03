# Architecture Exercises

Mỗi bài tập có `problem.md` và `solution.md`. Hãy đọc problem trước, tự thiết kế boundary, rồi mới mở solution.

Mục tiêu không phải tìm đáp án duy nhất. Mục tiêu là luyện cách phân tích:

- Business rule nằm đâu?
- Dependency nào đúng hướng?
- Interface nào có giá trị?
- Layer nào đang dư?
- Trade-off nào chấp nhận được?

## Cách làm

Với mỗi problem, hãy nộp:

1. assumptions và non-functional requirements;
2. domain language/invariants;
3. compile-time dependency diagram;
4. runtime happy/failure flow;
5. ports với Go signatures;
6. transaction/idempotency/concurrency decisions;
7. test portfolio;
8. “khi nào không dùng” hoặc simpler alternative.

Tự chấm 0-2 điểm mỗi mục. Chỉ mở solution sau khi đạt ít nhất 10/16. Solution là một design hợp lệ theo assumptions, không phải đáp án duy nhất; hãy ghi điểm bạn phản đối và alternative.
