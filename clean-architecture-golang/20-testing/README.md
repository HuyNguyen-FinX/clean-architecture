# 20 Testing

## Tại sao cần học?

Clean Architecture không tồn tại chỉ để test dễ hơn, nhưng boundary tốt làm test rẻ và chính xác hơn.

## Test Pyramid

- Domain test: nhanh, không mock.
- Use case test: fake repository/gateway.
- Repository integration test: PostgreSQL thật hoặc container.
- HTTP test: `httptest`.
- End-to-end test: ít hơn, đắt hơn, kiểm tra wiring.

## Mock, Stub, Fake, Spy

Không mock mọi thứ. Fake repository thường đọc dễ hơn mock call expectation khi test use case.

## Anti-pattern

- Test toàn mock nhưng không kiểm tra invariant.
- Unit test repository bằng mock database driver.
- E2E test thay thế mọi test nhỏ.
- Interface explosion chỉ vì muốn mock.

## Mastery Check

- [ ] Tôi biết domain test không cần mock.
- [ ] Tôi biết use case test không cần PostgreSQL.
- [ ] Tôi biết adapter cần integration test thật.
