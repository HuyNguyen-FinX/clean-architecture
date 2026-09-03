# Solution tham khảo: Loan Service

## Strategic model

Contexts có thể là Loan Origination, Risk Decision và Disbursement. Ban đầu deploy modular monolith nhưng ownership/API rõ. Không share một Customer struct toàn hệ thống; dùng ApplicantSnapshot/IDs.

## Aggregate và language

LoanApplication transitions:

~~~text
Draft → Submitted → UnderReview → Approved/Rejected
Approved → Disbursing → Disbursed/DisbursementFailed
~~~

RiskScore là Value Object nếu range/source/time có semantics. LoanDecision chứa reasons/audit policy version.

## Ports

~~~go
type CreditScoringGateway interface {
	Score(context.Context, ApplicantSnapshot) (RiskAssessment, error)
}

type DisbursementGateway interface {
	Disburse(context.Context, DisbursementRequest) (DisbursementResult, error)
}
~~~

Application loads aggregate, calls gateway ngoài DB transaction, then domain transition in local transaction. Unknown disbursement → PendingInquiry, không mark failed mù.

## Approval policy

Nếu internal business rules, Domain Service pure nhận assessment/config version. Nếu provider owns decision, ACL maps decision and domain validates allowed transition.

## Audit

Store decision inputs summary, policy/model version, actor/time/reasons, transitions/outbox. PII access/retention/encryption. Technical logs không thay audit.

## Consistency/workflow

Long-running process manager with timeouts, idempotent commands/events và manual review. Compensation for failed disbursement may be cancellation, but if funds ambiguous requires inquiry/reconcile.

## Tests

- transition matrix/policy boundaries;
- gateway fixtures/timeouts;
- process manager retries/duplicates;
- audit/outbox transaction;
- authorization;
- E2E manual/unknown paths.

## Alternative

If provider returns final approve/reject and app only stores result, rich local approval Domain Service may be false complexity. Keep state machine/ACL/audit.
