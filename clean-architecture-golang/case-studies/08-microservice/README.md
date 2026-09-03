# Case Study 08: Từ Modular Monolith Đến Microservice

Clean Architecture trả lời cách dependency được tổ chức bên trong một deployable unit. Microservice architecture trả lời boundary triển khai, ownership dữ liệu và giao tiếp qua network. Hai vấn đề liên quan nhưng không thay thế nhau.

## Scenario

Một e-commerce monolith có Order, Inventory, Payment và Notification. 25 engineers chia thành bốn teams. Deploy chung bắt đầu gây coordination; Payment cần security/reliability riêng, Inventory có tải burst và schema được nhiều module đọc trực tiếp.

Câu hỏi không phải "microservice có scalable hơn không?" mà là boundary nào có cohesion, quyền sở hữu và lý do deploy độc lập đủ mạnh để bù operational cost.

## Baseline: Modular Monolith

~~~text
cmd/api
internal/order/{domain,application,postgres,http}
internal/inventory/{domain,application,postgres,http}
internal/payment/{domain,application,provider,http}
internal/notification/...
~~~

Mỗi module expose application API; module khác không import adapter hoặc đọc table riêng của nhau. Một process/database cluster vẫn có thể dùng transaction local. Architecture fitness tests cấm import sai.

Đây thường là điểm bắt đầu tốt: boundary logic được luyện trước khi thêm network. Nếu team không giữ được module boundary trong monolith, network không tự chữa coupling; nó chỉ chuyển coupling thành API/event và tăng failure modes.

## Chọn Service Boundary

Đánh giá theo:

- Ubiquitous language và invariant nào phải nhất quán cùng nhau.
- Team nào sở hữu roadmap/on-call/data.
- Release cadence và blast radius.
- Security/compliance isolation.
- Scaling profile và dependency riêng.
- Tần suất synchronous calls qua boundary.

Không tách theo bảng (`UserService`, `AddressService`) hoặc theo technical layer (`RepositoryService`). Boundary nên tạo business capability tương đối tự chủ.

## Data Ownership

Sau khi tách Payment, chỉ Payment ghi payment tables. Order không join trực tiếp database của Payment. Nó dùng API cho quyết định cần hiện tại hoặc local projection cho read. Shared database với "mỗi service một schema" có thể là migration stage, nhưng quyền truy cập phải enforce và có exit plan.

Không sao chép toàn Aggregate qua event. Publish integration contract tối thiểu, versioned, có semantic rõ. Consumer sở hữu projection và chấp nhận eventual consistency.

## Compile-Time Và Runtime

~~~mermaid
flowchart LR
    CLIENT["Client"] --> ORDER["Order service"]
    ORDER --> PAYAPI["Payment API"]
    ORDER --> ODB["Order DB"]
    PAYAPI --> PDB["Payment DB"]
    PAYAPI --> PROVIDER["Payment provider"]
    PAYAPI --> OUTBOX["Payment outbox"]
    OUTBOX --> BROKER["Kafka"]
    BROKER --> ORDER
~~~

Bên trong mỗi service vẫn có dependency rule. Qua network, OpenAPI/Protobuf/event schema trở thành dependency semantic. Không import source code không có nghĩa là decoupled: thay meaning field vẫn phá consumer.

## Synchronous Hay Asynchronous

Synchronous API phù hợp khi caller cần answer ngay và dependency có latency/reliability budget. Async event phù hợp notification, propagation và workflow chấp nhận pending. Không dùng event cho mọi query chỉ để "decoupled".

Một request Order gọi tuần tự Inventory, Payment, Shipping tạo latency cộng dồn và availability nhân. Cần deadline propagation, timeout per-hop, bounded retry, circuit/bulkhead và fallback có semantics. Retry POST chỉ khi idempotent.

## Distributed Workflow

Không dùng distributed transaction 2PC mặc định. Order Saga lưu state và outbox cho từng local transition; Payment/Inventory xử lý command idempotently, trả event. Compensation là business action như void authorization/release stock, có thể thất bại và cần manual resolution.

~~~text
Order Pending
  -> Inventory Reserved
  -> Payment Authorized
  -> Order Confirmed

Payment Declined
  -> Release Inventory
  -> Order Cancelled
~~~

Eventual consistency phải xuất hiện trong UX/API: `202 Pending`, poll/webhook, status reason. Che pending bằng success giả làm contract sai.

## Migration Theo Strangler

1. Đo coupling hiện tại: imports, table access, call graph, ownership.
2. Làm rõ boundary Payment trong monolith; thêm application facade.
3. Cấm module khác đọc/write payment tables; thay bằng facade.
4. Định nghĩa contract và characterization tests từ behavior hiện tại.
5. Tạo new service + backfill data có checksum.
6. Dùng outbox/change feed cho delta; tránh dual-write không atomic.
7. Shadow read/compare; sau đó canary traffic theo merchant/tenant.
8. Cut over writer, giữ rollback strategy không tạo split brain.
9. Gỡ đường cũ và quyền database sau stabilization.

Đừng big-bang rewrite. Mỗi bước phải deploy/revert độc lập và giữ observable equivalence.

## API Và Event Evolution

- Additive field thường an toàn nếu consumer bỏ qua unknown fields.
- Đổi meaning/type không an toàn dù tên endpoint giữ nguyên.
- Consumer-driven contract tests giúp phát hiện break nhưng không thay governance.
- Event tồn tại lâu cần schema version, compatibility rule và replay plan.
- Idempotency key/correlation/causation ID là phần contract, không phải header tùy hứng.

## Failure Matrix

| Failure | Thiết kế cần có |
|---|---|
| Payment DNS/timeout | deadline, bounded retry, pending/unknown state |
| Order commit, event publish fail | transactional outbox |
| Duplicate/reordered event | inbox + state transition/version |
| Consumer lag hàng giờ | lag/age SLO, projection freshness trong API |
| Schema không tương thích | compatibility CI + quarantine |
| Region/network partition | ownership/write policy rõ, không dual writer tùy tiện |
| Cascading overload | load shedding, pool limits, circuit/bulkhead |

## Observability Và Operations

Mỗi service có SLI riêng nhưng user journey cần trace/correlation xuyên boundary. Propagate W3C trace context, không buộc domain import OpenTelemetry. Metrics có RED cho API, lag/outbox age cho async, saturation cho pools. Log cùng `trace_id`, `operation_id`, contract version.

Microservice thêm deployment, discovery, secrets, TLS, schema registry, dashboards, runbooks, on-call và cost. "Mỗi service pass unit test" không chứng minh workflow end-to-end đúng.

## Testing Strategy

- Domain/use-case tests trong từng service.
- Adapter contract tests với provider và persistence thật.
- Provider/consumer contract tests cho API/event compatibility.
- Integration test outbox/inbox/idempotency.
- End-to-end ít nhưng tập trung money/order critical journey.
- Resilience test timeout, retry storm, duplicate, reorder, broker lag.
- Migration test backfill checksum, shadow compare và rollback.

## Trade-off Decision Record

Một service extraction được duyệt khi nêu rõ capability owner, dữ liệu sở hữu, consistency model, API/event contracts, SLO, failure handling, migration/rollback và on-call. Nếu lợi ích chỉ là "deploy riêng trong tương lai", modular monolith thường rẻ hơn.

Shared library chỉ chứa stable technical utilities/contracts rất thận trọng; chia sẻ domain model binary giữa services tạo lockstep deploy. Copy một DTO nhỏ đôi khi ít coupling hơn shared package.

## Clean Architecture Giúp Và Không Giúp

Giúp: cô lập provider/database, giữ use case testable, tạo application facade để extraction, làm dependency ownership rõ. Không giúp tự động: network partition, data consistency, event ordering, capacity planning, org ownership hay incident response.

## Câu Hỏi Mastery

1. Dấu hiệu nào cho thấy Payment đáng tách service thay vì chỉ module?
2. Vì sao database-per-service làm một query report khó hơn?
3. Source independence khác semantic independence ra sao?
4. Canary writer cần tránh split brain thế nào?
5. Nếu Order cần Payment status cho mọi request, boundary hiện tại có vấn đề gì?

## Bài Thực Hành

Viết ADR tách Payment gồm context, options, decision, consequences, migration và rollback. Vẽ sequence khi provider timeout và khi Kafka down; ở mỗi arrow ghi deadline, idempotency và owner của retry.
