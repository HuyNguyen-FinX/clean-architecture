# Lab 07: Test Portfolio Theo Boundary

Thời lượng: 90-150 phút.

## Mục tiêu

- domain test không mock;
- use-case test với fake/spy;
- Repository contract test;
- HTTP test bằng httptest;
- phân biệt failure mà mỗi test không chứng minh.

## Kiến thức cần

- [Testing](../../20-testing/README.md)
- Domain, Application, Repository và HTTP chapters.

## Diagram

~~~mermaid
flowchart TB
    DOMAIN["Domain unit"] --> APP["Use-case fake/spy"]
    APP --> CONTRACT["Repository contract"]
    CONTRACT --> HTTP["HTTP adapter test"]
    HTTP --> INTEGRATION["Real dependency integration"]
    INTEGRATION --> E2E["Few E2E"]
~~~

## Problem

Starter có public balance, concrete map service và chỉ happy-path smoke test. Hãy refactor/test theo question, không chase coverage.

## Yêu cầu

1. Domain matrix: happy, exact boundary, insufficient, invalid.
2. Use case: no save on reject, save both on success.
3. Fake clone ownership; spy records saves.
4. HTTP maps command/status/error.
5. Race detector pass.
6. Ghi test gap cho SQL/rollback/locking.

## Các bước

1. Characterize starter.
2. Extract Account behavior và test table.
3. Extract Repository port/use case.
4. Viết fake state + spy fields.
5. Viết HTTP adapter fake use case.
6. Viết test-gap note trước khi tuyên bố guarantee.

## Expected behavior

Transfer success đổi hai balance; insufficient không save; HTTP 409 cho rejection; malformed body 400. Test portfolio vẫn không chứng minh PostgreSQL rollback.

## Test

~~~bash
cd starter && go test ./...
cd ../solution && go test -race ./... && go vet ./...
~~~

## Questions

1. Fake, stub, spy và mock khác gì trong tests?
2. Race detector không bắt logical DB race nào?
3. Contract suite thêm giá trị gì?
4. Test nào đắt nhưng cần cho SQL?
5. 100% coverage có thể vẫn bỏ gì?

## Challenge

- Fuzz amount parser.
- Thêm PostgreSQL integration.
- Tạo deterministic concurrent barrier.
- Viết architecture import test.

## Solution explanation

Solution giữ một module nhỏ nhưng chia domain/application/http/memory. Mỗi test nêu đúng guarantee; test code không mock Entity và không assert telemetry sequence.
