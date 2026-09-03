# Lab 06: Manual Dependency Injection và Composition Root

Thời lượng: 60-90 phút. Lab dùng standard library để tập trung vào object graph và lifecycle.

## Mục tiêu

- loại global mutable dependency;
- parse/validate config ở outer boundary;
- lắp Store → UseCase → Handler;
- làm resource ownership/cleanup explicit;
- smoke test graph qua HTTP.

## Kiến thức cần

- [Dependency Injection](../../08-dependency-injection/README.md)
- [Project Structure](../../09-project-structure/README.md)
- constructor injection, httptest.

## Diagram

~~~mermaid
flowchart LR
    CONFIG["Config"] --> BUILD["Build"]
    BUILD --> STORE["MemoryStore"]
    BUILD --> UC["GetBalance"]
    BUILD --> HANDLER["HTTP Handler"]
    HANDLER --> UC --> STORE
~~~

## Problem

Starter dùng package global store, handler tự gọi global và config đọc tại nơi sử dụng. Test có thể ảnh hưởng nhau và dependency bị ẩn.

## Yêu cầu

1. Config được validate trước khi build.
2. Mỗi dependency bắt buộc đi qua constructor.
3. Handler chỉ biết use-case interface.
4. Build trả App và cleanup function idempotent.
5. Không dùng package global mutable.
6. Smoke test gọi route và kiểm cleanup.

## Các bước

1. Chạy starter, tìm hidden dependency.
2. Extract Store port từ GetBalance consumer.
3. Inject Store vào use case.
4. Inject use case vào Handler.
5. Tạo Build ở composition package.
6. Validate config và test error path.
7. Dùng httptest gọi App.Handler.
8. Gọi cleanup hai lần để xác minh lifecycle an toàn.

## Expected behavior

GET /accounts/A trả balance seed. ID lạ trả 404. Config thiếu address bị reject trước khi graph chạy. Cleanup đóng resource đúng một lần theo semantics idempotent.

## Test

~~~bash
cd starter && go test ./...
cd ../solution && go test -race ./... && go vet ./...
~~~

## Questions

1. DI khác Dependency Inversion thế nào trong solution?
2. Vì sao composition package được biết concrete adapter?
3. Tại sao config toàn cục là hidden dependency?
4. Ai sở hữu cleanup?
5. Với graph này, Wire/Fx có hoàn vốn không?

## Challenge

- Thêm PostgreSQL implementation được chọn qua config.
- Thêm signal-based graceful shutdown.
- Viết startup health check fail-fast.
- Thêm một worker và xác định thứ tự stop.

## Solution explanation

Solution dùng manual DI để object graph đọc được từ trên xuống. Interface nằm phía consumer; Store vẫn là concrete khi được tạo. Build là composition root có thể biết cả adapter và delivery, còn use case không biết cách chúng được tạo.
