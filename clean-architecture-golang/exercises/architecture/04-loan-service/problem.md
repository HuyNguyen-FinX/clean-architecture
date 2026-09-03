# Bài 4: Thiết Kế Loan Service

## Requirement

Loan Service:

- Submit loan application.
- Check eligibility.
- Score risk.
- Approve/reject loan.
- Disburse loan.

## Nhiệm vụ

1. Xác định Bounded Context.
2. Thiết kế domain model.
3. Thiết kế external gateway tới credit scoring/core banking.
4. Phân tích rule nào thuộc domain, rule nào thuộc provider.
5. Thiết kế audit trail.

## Câu hỏi

- Risk score có phải Value Object không?
- Approval rule nằm trong domain hay application?
- Disbursement failure xử lý thế nào?
