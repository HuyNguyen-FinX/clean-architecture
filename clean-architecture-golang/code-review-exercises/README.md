# Code Review Exercises

Đọc problem và source trước, tự ghi findings theo severity rồi mới mở solution.

Review order:

1. correctness/data loss/security;
2. concurrency/transaction/idempotency;
3. error/context/lifecycle;
4. dependency/ownership;
5. testing/operability;
6. naming/style.

Một review senior không bắt đầu bằng “nên chia folder”. Nó chỉ ra failure cụ thể, path tái hiện và refactor nhỏ nhất an toàn.

## Bài hiện có

- [01 God Handler tưởng như sạch](./01-god-handler/problem.md): source compile được, có interfaces/constructor nhưng transaction, dependency và async guarantee sai tinh tế.
