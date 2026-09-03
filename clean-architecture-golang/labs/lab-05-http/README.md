# Lab 05: Strict HTTP Adapter

Thời lượng: 60-100 phút.

## Mục tiêu

- limit/decode đúng một JSON object;
- tách request DTO khỏi command;
- truyền request context;
- map stable application errors;
- không leak unknown error;
- test contract bằng httptest.

## Kiến thức cần

- [Delivery Layer](../../06-delivery-layer/README.md)
- [HTTP REST API](../../12-http-rest-api/README.md)
- [Error Handling](../../17-error-handling/README.md)

## Diagram

~~~mermaid
flowchart LR
    HTTP["POST /transfers"] --> DTO["transferRequest"]
    DTO --> CMD["TransferCommand"]
    CMD --> UC["Transfer use case"]
    UC --> MAP["Error/result mapping"]
    MAP --> RESPONSE["Stable JSON response"]
~~~

## Problem

Starter bỏ decode error, dùng context.Background và trả err.Error. Viết input có hai JSON values hoặc fake use case trả secret error để quan sát.

## Yêu cầu

1. Max body 1 MiB.
2. Content-Type application/json.
3. Reject unknown fields và trailing JSON.
4. DTO map sang application command.
5. r.Context truyền nguyên vào use case.
6. ErrInsufficientBalance → 409 stable code.
7. Unknown → 500 không chứa cause.
8. Test happy/error/malformed/context.

## Các bước

1. Chạy starter và thêm characterization test cho leak.
2. Extract Transfer interface consumer-side.
3. Viết decodeJSON.
4. Viết responseFromError.
5. Set headers trước status.
6. Viết fake recording command/context.
7. Chạy race/vet.

## Expected behavior

Valid body trả 201; malformed/unknown/trailing/wrong media trả 400 hoặc 415 theo contract; insufficient trả 409; internal trả public internal_error.

## Test

~~~bash
cd starter && go test ./...
cd ../solution && go test -race ./... && go vet ./...
~~~

## Questions

1. Vì sao handler vẫn có logic dù “mỏng”?
2. Domain error có biết 409 không?
3. Client disconnect sau commit tạo failure gì?
4. DisallowUnknownFields có trade-off compatibility nào?
5. Idempotency-Key cần xử lý ngoài handler ở đâu?

## Challenge

- body-too-large response riêng;
- middleware request ID/recovery;
- idempotency header + request hash;
- graceful shutdown smoke test.

## Solution explanation

Solution dùng net/http thuần và fake use case. Test chỉ khóa transport mapping; domain/application tests riêng giữ business rule. Bản xuyên suốt ở [mini-banking handler](../../examples/mini-banking/internal/account/delivery/http/handler.go).
