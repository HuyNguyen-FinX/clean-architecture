# Project Structure: package tree phải kể đúng ownership

Không có folder tree nào tự động biến code thành Clean Architecture. Go compiler chỉ nhìn package/import; production team còn phải nhìn ownership, change coupling và cognitive load.

## Kết quả học tập

- chọn package by layer, feature hoặc hybrid theo quy mô;
- dùng internal/cmd/pkg đúng mục đích;
- đọc dependency direction từ import graph;
- kiểm soát shared package và cyclic dependency;
- tổ chức migration, transport, worker và test;
- tiến hóa structure mà không “big bang”.

## 1. Problem: folder đẹp, dependency vẫn sai

~~~text
internal/
  domain/
  usecase/
  repository/
  handler/
~~~

Tree có vẻ sạch nhưng domain import repository/postgres thì boundary vẫn bị phá. Ngược lại, service nhỏ với ba package handler/service/store có thể rất rõ nếu dependency một chiều và responsibility đúng.

Architecture nằm trong:

- package ownership;
- import direction;
- public API;
- nơi business decisions thay đổi;
- modules nào phải sửa cùng nhau.

## 2. Ba level

### Level 1: trực giác

Đặt code thay đổi cùng lý do gần nhau. Tách code thay đổi vì lý do khác nhau khi coupling bắt đầu gây đau.

### Level 2: Backend Engineer

Package boundary quyết định:

- tên nào được export;
- test có thể truy cập gì;
- import cycle có bị compiler chặn;
- compile scope và discoverability;
- team tìm feature nhanh hay phải nhảy nhiều folder.

### Level 3: Architecture

Structure là bản đồ ownership và socio-technical boundary. Bounded Context, team ownership và deployment unit thường quan trọng hơn tên layer. Một “shared domain” không có owner dễ trở thành coupling hub.

## 3. Ba cách tổ chức

### Package by layer

~~~text
internal/
  handlers/
  services/
  repositories/
  models/
~~~

Phù hợp khi:

- codebase nhỏ;
- domain đơn giản;
- team quen technical layers;
- feature thường thay đổi xuyên cùng pipeline.

Rủi ro khi lớn: thư mục services có hàng chục use case không chung domain; thay Account phải đi qua bốn thư mục xa nhau; ownership theo feature mờ.

### Package by feature

~~~text
internal/
  account/
  transfer/
  customer/
~~~

Mỗi feature có API hẹp, locality tốt. Nhưng nếu mọi file domain/application/SQL nằm chung package, detail có thể truy cập internals không kiểm soát.

### Feature + internal boundary

~~~text
internal/
  account/
    domain/
    application/
    delivery/http/
    infrastructure/postgres/
  transfer/
    domain/
    application/
    delivery/kafka/
~~~

Đây là lựa chọn mini-banking. Nó ưu tiên feature ownership rồi dùng package con khi boundary có giá trị.

## 4. Trade-off theo quy mô

| Bối cảnh | Điểm bắt đầu hợp lý | Rủi ro cần theo dõi |
|---|---|---|
| 10k LOC, team 3 | handler/service/store hoặc feature packages | ceremony nhiều hơn code |
| 50k LOC, team 6-10 | feature + domain/application/adapters | shared package tăng dần |
| 500k LOC, team 30 | bounded contexts/modules rõ, ownership tests | coupling giữa context, build/release |
| CRUD ngắn hạn | package phẳng | abstraction giả |
| Core banking | Aggregate/use case/adapters rõ | transaction xuyên context |

Số LOC không phải trigger duy nhất. Volatility, regulatory risk, throughput và team topology cũng ảnh hưởng.

## 5. Cấu trúc mini-banking

~~~text
examples/mini-banking/
  cmd/api/main.go
  internal/
    account/
      domain/
      application/
      delivery/http/
      infrastructure/
        memory/
        postgres/
    architecture/dependency_test.go
~~~

Source dependency:

~~~mermaid
flowchart TD
    CMD["cmd/api"] --> HTTP["delivery/http"]
    CMD --> MEM["infrastructure/memory"]
    CMD --> PG["infrastructure/postgres"]
    HTTP --> APP["application"]
    MEM --> APP
    PG --> APP
    APP --> DOMAIN["domain"]
    MEM --> DOMAIN
    PG --> DOMAIN
~~~

Không có mũi tên domain/application đi ra adapter. Architecture test parse imports để giữ rule này tại [dependency_test.go](../examples/mini-banking/internal/architecture/dependency_test.go).

## 6. Package name và public surface

Go package là unit encapsulation. Private field/method có giá trị khi package nhỏ và cohesive; một package domain 200 file làm private gần như public nội bộ.

Guideline:

- tên package ngắn, theo capability: account, postgres, http;
- tránh utils/common/base vì không nói ownership;
- export tối thiểu;
- constructor trả type concrete khi caller không cần interface;
- interface đặt cạnh consumer;
- tránh stutter như account.AccountServiceManager.

## 7. internal, pkg và cmd

### internal

Go toolchain cấm module ngoài parent tree import package dưới internal. Đây là enforcement thật, không chỉ convention. Dùng cho application code không cam kết public compatibility.

### pkg

pkg không có semantics đặc biệt với compiler. Chỉ đưa code vào đó khi thật sự muốn public/reusable API. Copy template có pkg/utils thường làm implementation detail bị phụ thuộc từ bên ngoài.

### cmd

Mỗi thư mục con dưới cmd là một executable/composition root:

~~~text
cmd/api/
cmd/outbox-worker/
cmd/migrate/
~~~

cmd nên mỏng: parse process config, build graph, run lifecycle. Không đặt business rule ở main chỉ vì main được import mọi adapter.

## 8. Delivery và infrastructure không phải lúc nào cũng một folder

HTTP và Kafka consumer đều là driving adapters: chúng chuyển external input thành application command. PostgreSQL, Kafka publisher và payment client là driven adapters: application gọi qua port.

Tên delivery/infrastructure tiện cho học, nhưng production có thể dùng:

~~~text
adapters/
  incoming/http/
  incoming/kafka/
  outgoing/postgres/
  outgoing/payment/
~~~

Chọn vocabulary team hiểu và dependency tests enforce. Không tranh luận tên folder trong khi imports đi sai.

## 9. Migration và generated code

Migration thuộc ownership của adapter/database module:

~~~text
infrastructure/postgres/
  repository.go
  migrations/
    001_accounts.sql
~~~

OpenAPI/protobuf generated code nên được cô lập:

~~~text
internal/account/delivery/http/openapi/
internal/account/delivery/grpc/gen/
~~~

Application không nhận generated request type. Adapter map sang command để schema transport không trở thành semantic dependency của core.

## 10. Shared package: lực hút nguy hiểm

Shared hợp lý với primitive thực sự ổn định như Money nếu nhiều bounded context cùng một definition và owner rõ. Nhưng shared/models, shared/errors, utils dễ tích lũy:

- HTTP response type;
- database helper;
- domain constants;
- logging;
- random conversion.

Mỗi feature import shared làm package đó thành điểm thay đổi có blast radius lớn.

Trước khi đưa vào shared, hỏi:

1. Đây là duplication kiến thức hay chỉ duplication vài dòng?
2. Các consumer có cùng change reason?
3. Ai sở hữu API compatibility?
4. Copy nhỏ có rẻ hơn coupling không?

## 11. Cross-feature dependency

Account application không nên import Loan infrastructure. Nếu Transfer cần Account và Ledger:

- đặt orchestration ở module Transfer;
- phụ thuộc port/capability public của Account/Ledger;
- hoặc giao tiếp qua event nếu consistency cho phép;
- tránh đọc table của context khác như private API.

Trong monolith, direct package call vẫn có thể sạch. Microservice không phải cách duy nhất để tạo boundary.

## 12. Import cycle là tín hiệu

Go cấm cycle:

~~~text
account -> customer -> account
~~~

Không “sửa” bằng cách đẩy type vào shared vô điều kiện. Điều tra:

- concept thuộc module nào?
- một chiều nào phản ánh business ownership?
- cần một port nhỏ ở consumer không?
- hai package thực chất là một cohesive unit chưa nên tách?

Merge package đôi khi đúng hơn thêm interface.

## 13. Runtime dependency khác package dependency

Postgres Repository được gọi từ application ở runtime, nhưng adapter import application port ở compile time. Folder tree chỉ giúp nhìn compile-time relation; object graph ở main cho runtime relation. Cần đọc cả hai.

## 14. Testing layout

- Domain unit tests cạnh package.
- Application tests dùng external package khi muốn chỉ nhìn public contract, internal package khi cần white-box hữu ích.
- Adapter integration tests cạnh adapter, gated bằng environment/container.
- Architecture fitness tests ở package riêng.
- End-to-end tests có thể ở tests/e2e nếu chạy process thật.

Không tạo tests/unit, tests/integration rồi mirror toàn bộ source tree nếu việc tìm test trở nên khó.

## 15. Evolution không big bang

Từ layered monolith:

~~~text
handlers/ services/ repositories/
~~~

Refactor một vertical slice:

1. chọn use case thay đổi nhiều;
2. khóa behavior bằng characterization test;
3. extract domain behavior;
4. đặt application port;
5. chuyển adapter;
6. composition root nối cả cũ/mới;
7. thêm import fitness test;
8. lặp slice tiếp theo.

Không di chuyển 500 file chỉ để đạt tree đích. Movement không đồng nghĩa decoupling.

## 16. Failure scenario

Team 30 người cùng sửa global services/repositories:

- merge conflict cao;
- interface chung phình;
- ownership review mờ;
- deploy một feature có thể chạm module khác;
- “common” thay đổi gây test toàn repo.

Đo change coupling bằng git history: file/module nào thường đổi cùng nhau? Structure nên hỗ trợ cách hệ thống thực sự tiến hóa, không chỉ sơ đồ lý tưởng.

## 17. Debug dependency

Các công cụ:

~~~bash
go list -deps ./...
go list -json ./...
go mod graph
go test ./internal/architecture
~~~

Khi thấy import sai, tìm semantic dependency trước. Chỉ chuyển type vào package trung tâm có thể làm arrow hợp lệ trên giấy nhưng ownership tệ hơn.

## 18. Khi nào không nên chia sâu?

Todo API 4 endpoint không cần account/domain/model/entity, application/usecase/service, adapter/repository/dao. Mỗi package một file làm navigation và mapping cost lớn hơn isolation.

Bắt đầu phẳng, tách khi behavior, team ownership hoặc volatile details tạo pressure thật. Clean Architecture cho phép proportionality.

## 19. Bài tập

1. Vẽ import graph của mini-banking và runtime graph riêng.
2. Thêm module transaction history, chọn package ownership.
3. Viết fitness test chặn domain import net/http và pgx.
4. Refactor một package shared giả định thành owner cụ thể.
5. So sánh tree cho team 3 và team 30, giải thích khác biệt.

## 20. Mastery questions

1. Vì sao folder theo layer không chứng minh dependency rule?
2. internal enforce điều gì, pkg enforce điều gì?
3. Khi nào merge hai package tốt hơn tạo port?
4. Shared package làm tăng blast radius thế nào?
5. Generated protobuf type đi vào application gây coupling gì?
6. Package by feature có rủi ro nào nếu package quá lớn?
7. Vì sao main mỏng dù được phép biết details?
8. Bạn dùng signal nào để quyết định tách package?

## Further reading

- [Organizing a Go module](https://go.dev/doc/modules/layout).
- Go specification về package/import cycle.
- Team Topologies, phần cognitive load và stream-aligned team.
- DDD literature về Bounded Context và Context Map.

## Quality gate

- [x] So sánh ba structure và nhiều quy mô
- [x] Package/import/runtime ownership
- [x] internal/pkg/cmd/shared/generated code
- [x] Correct/wrong trade-off, không template dogma
- [x] Evolution, failure, debugging, tests
- [x] Exercises, mastery và executable architecture test
