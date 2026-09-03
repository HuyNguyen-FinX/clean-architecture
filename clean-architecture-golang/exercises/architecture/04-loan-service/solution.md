# Solution: Loan Service

Loan có domain giàu rule, phù hợp với Clean Architecture hơn Todo CRUD.

Domain:

- `LoanApplication`.
- `RiskScore`.
- `LoanDecision`.
- State transition: submitted, under_review, approved, rejected, disbursing, disbursed.

External gateway:

```go
type CreditScoringGateway interface {
	Score(ctx context.Context, applicant ApplicantSnapshot) (RiskScore, error)
}
```

Approval rule có thể nằm trong domain service nếu là nghiệp vụ nội bộ. Nếu phụ thuộc provider policy, application gọi gateway rồi domain quyết định state transition dựa trên result đã map.
