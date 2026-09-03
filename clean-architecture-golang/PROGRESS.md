# Progress

Trạng thái trong file này dùng quality gate tại [CONTENT_AUDIT.md](./CONTENT_AUDIT.md), không dựa vào việc folder/file đã tồn tại.

## Ý Nghĩa Trạng Thái

- `[DEEP COMPLETE]`: đạt quality gate production-grade learning material; code và link liên quan đã được kiểm chứng.
- `[FOUNDATIONAL COMPLETE]`: đủ làm nền và đã có chiều sâu đáng kể, nhưng còn thiếu ít nhất một quality gate nâng cao.
- `[IN PROGRESS]`: đang được rewrite theo chiều sâu; chưa nên học như chapter hoàn chỉnh.
- `[OUTLINE ONLY]`: mới là summary, danh sách chủ đề hoặc hướng làm.
- `[NOT STARTED]`: chưa có material có ý nghĩa.

## Core Docs

- [FOUNDATIONAL COMPLETE] `README.md`
- [FOUNDATIONAL COMPLETE] `ROADMAP.md`
- [IN PROGRESS] `GLOSSARY.md`
- [FOUNDATIONAL COMPLETE] `CHEATSHEET.md`
- [DEEP COMPLETE] `CONTENT_AUDIT.md` cho baseline 2026-09-03
- [IN PROGRESS] Runnable example `examples/mini-banking`: V1-V3 chạy được; chưa có transaction production, persistence và reliability evolution.

## Chapters

- [FOUNDATIONAL COMPLETE] 00 Software Architecture Foundations
- [DEEP COMPLETE] 01 Clean Architecture Foundations
- [DEEP COMPLETE] 02 Dependency Rule
- [DEEP COMPLETE] 03 Domain Layer
- [OUTLINE ONLY] 04 Use Case / Application Layer
- [OUTLINE ONLY] 05 Repository Pattern
- [OUTLINE ONLY] 06 Delivery Layer
- [OUTLINE ONLY] 07 Infrastructure Layer
- [OUTLINE ONLY] 08 Dependency Injection
- [OUTLINE ONLY] 09 Project Structure
- [OUTLINE ONLY] 10 Database
- [OUTLINE ONLY] 11 Transaction Management
- [OUTLINE ONLY] 12 HTTP REST API
- [OUTLINE ONLY] 13 gRPC
- [OUTLINE ONLY] 14 Kafka Event Driven
- [OUTLINE ONLY] 15 Redis Cache
- [OUTLINE ONLY] 16 External Services
- [OUTLINE ONLY] 17 Error Handling
- [OUTLINE ONLY] 18 Validation
- [OUTLINE ONLY] 19 Logging / Observability
- [OUTLINE ONLY] 20 Testing
- [OUTLINE ONLY] 21 Concurrency Golang
- [OUTLINE ONLY] 22 Domain Driven Design
- [OUTLINE ONLY] 23 CQRS / Event Driven
- [OUTLINE ONLY] 24 Production Architecture
- [OUTLINE ONLY] 25 Refactoring
- [OUTLINE ONLY] 26 Anti-patterns
- [OUTLINE ONLY] 27 Case Studies
- [OUTLINE ONLY] 28 System Design
- [OUTLINE ONLY] 29 Interview Review

## Labs

- [DEEP COMPLETE] `lab-01-simple-domain`
- [OUTLINE ONLY] `lab-02-usecase`
- [OUTLINE ONLY] `lab-03-repository`
- [OUTLINE ONLY] `lab-04-postgresql`
- [OUTLINE ONLY] `lab-05-http`
- [OUTLINE ONLY] `lab-06-dependency-injection`
- [OUTLINE ONLY] `lab-07-testing`
- [OUTLINE ONLY] `lab-08-transaction`
- [OUTLINE ONLY] `lab-09-kafka`
- [OUTLINE ONLY] `lab-10-redis`
- [OUTLINE ONLY] `lab-11-refactoring`
- [OUTLINE ONLY] `lab-12-full-application`

## Exercises Và Case Studies

- [IN PROGRESS] Architecture exercises 01-07: problem/solution đã tách, solution còn nông.
- [OUTLINE ONLY] Code review exercise 01: chưa có realistic Go source để review.
- [OUTLINE ONLY] Case study briefs 01-08.

## Increment Tiếp Theo

1. Rewrite sâu `04-usecase-application-layer` và biến `lab-02-usecase` thành lab chạy được.
2. Rewrite sâu `05-repository-pattern` và hoàn thiện contract tests cho memory adapter.
3. Nâng `08-dependency-injection` và `09-project-structure` dựa trên object/import graph hiện có.
4. Thêm PostgreSQL ở `10-database`, rồi transaction thật ở `11-transaction-management`.

## Quy Tắc Cập Nhật

Mỗi lần đổi trạng thái phải ghi bằng chứng ở cuối file:

- Material nào được thêm hoặc sửa.
- Full example nào compile/test.
- Failure scenario nào đã được phân tích.
- Quality gate nào còn thiếu.

Không nâng trạng thái hàng loạt chỉ vì đã thêm cùng một template vào nhiều chapter.

## Lịch Sử Increment

### 2026-09-03 - Foundations, Dependency Rule Và Domain

- `CONTENT_AUDIT.md`: audit baseline toàn bộ 118 file/4.780 dòng và xếp priority.
- Chapter 01: 878 dòng, đạt quality gate foundations.
- Chapter 02: 894 dòng và architecture fitness test bằng `go/parser`.
- Chapter 03: 10 bài, tổng 2.831 dòng, dùng mini-banking làm model xuyên suốt.
- Mini-banking domain: constructor invariant, Currency/Money equality, overflow guard, frozen state và test matrix.
- `lab-01`: starter/solution có code thật và hướng dẫn 60-90 phút.
- Verification: mini-banking, lab starter và lab solution đều pass `go test -race ./...`; relative links của material mới tồn tại; `git diff --check` sạch.
