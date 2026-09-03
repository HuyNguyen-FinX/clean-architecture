# gRPC Adapter: protobuf contract không phải domain model

gRPC là driving adapter giống HTTP. Generated protobuf code giải quyết wire compatibility, không nên sở hữu invariant hoặc trở thành model chung cho toàn core.

## Kết quả học tập

- map protobuf request/response sang application command/result;
- map error sang gRPC status;
- truyền deadline/metadata/auth đúng boundary;
- quản protobuf evolution;
- test server bằng bufconn/in-process transport;
- phân tích streaming, retry và idempotency.

## 1. Problem

~~~go
func (s *Server) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferReply, error) {
	account := &domain.Account{Balance: req.Amount}
	return s.repo.Save(ctx, account)
}
~~~

Generated type đi sâu vào domain làm field presence/default/version của protobuf shape business model. Thay protobuf hoặc thêm internal caller kéo generated package vào core.

## 2. Ba level

### Level 1

gRPC Server dịch message thành use-case intent, rồi dịch outcome thành status/message.

### Level 2

Adapter sở hữu generated interface, metadata, auth, deadline, status, interceptors và streaming lifecycle.

### Level 3

Protobuf là compatibility schema. Application command là semantic contract nội bộ. Tách hai model cho phép wire evolution độc lập và ngăn transport default tạo invalid domain state.

## 3. Dependency

~~~mermaid
flowchart LR
    CLIENT["gRPC Client"] --> SERVER["gRPC Adapter"]
    SERVER --> APP["Application"]
    APP --> DOMAIN["Domain"]
    SERVER --> PB["Generated protobuf"]
~~~

Application không import pb. gRPC adapter import application và pb.

## 4. Unary handler

~~~go
func (s *Server) Transfer(
	ctx context.Context,
	req *bankv1.TransferRequest,
) (*bankv1.TransferResponse, error) {
	cmd := application.TransferMoneyCommand{
		FromAccountID: req.GetFromAccountId(),
		ToAccountID:   req.GetToAccountId(),
		Amount:        req.GetAmountMinor(),
		Currency:      req.GetCurrency(),
	}
	result, err := s.transfer.Execute(ctx, cmd)
	if err != nil {
		return nil, statusFromError(err)
	}
	return &bankv1.TransferResponse{TransferId: result.ID}, nil
}
~~~

Đây là conceptual snippet vì repo không generate protobuf. Full handler phải implement generated server interface và compile trong module có protoc artifacts.

## 5. Error mapping

| Application/domain | gRPC code |
|---|---|
| invalid command | InvalidArgument |
| account missing | NotFound |
| insufficient balance/state | FailedPrecondition |
| idempotency conflict | AlreadyExists/Aborted theo contract |
| authorization | PermissionDenied |
| deadline | DeadlineExceeded |
| unknown | Internal |

Không trả raw error qua status.Error(codes.Internal, err.Error()). Message public phải sanitized; diagnostics vào log/trace.

## 6. Field presence và validation

Proto3 scalar default làm absent amount và amount=0 giống nhau nếu không dùng optional/wrapper. Nếu distinction quan trọng, schema phải biểu diễn presence.

Validate:

- wire format/required ở adapter/generated validator;
- cross-command/actor ở application;
- invariant ở domain.

## 7. Deadline và cancellation

ctx từ gRPC đã mang deadline/cancel. Truyền xuống use case/repository. Client deadline quá ngắn có thể cancel ngay trong transaction; commit outcome có thể ambiguous, nên idempotency vẫn cần.

Server có thể enforce max deadline hoặc default budget bằng interceptor, nhưng không kéo grpc metadata vào domain.

## 8. Metadata và auth

Interceptor verify token/mTLS principal, map sang Actor. Correlation/trace metadata đi qua context bằng typed key/telemetry propagation.

Không cho application đọc metadata map tùy ý; nó tạo semantic dependency lên header names.

## 9. Interceptors

Phù hợp:

- recovery;
- auth extraction;
- request/trace IDs;
- metrics/logging;
- transport rate limit;
- deadline policy.

Không phù hợp: overdraft, loan decision, database transaction business.

Interceptor order cần test như HTTP middleware.

## 10. Streaming

### Server streaming

History stream cần backpressure và context cancellation. Không giữ DB transaction mở suốt client chậm. Query page/chunk, đóng rows/resource đúng lúc.

### Client/bidirectional streaming

Mỗi message có identity/idempotency semantics. Stream disconnect không tự rollback remote effects. Chọn per-message ack/result rõ.

Application interface có thể dùng iterator/channel, nhưng channel có goroutine ownership/cancellation cost. Đừng để pb stream interface leak vào core.

## 11. Protobuf evolution

- không reuse field number;
- reserve removed numbers/names;
- add fields với tolerant defaults;
- enum cần UNSPECIFIED;
- không đổi semantic cũ dưới cùng field;
- version package khi breaking;
- test old/new clients nếu rolling.

Domain enum không nhất thiết mirror protobuf enum 1:1. Adapter map unknown value thành invalid/unspecified policy.

## 12. Retry và idempotency

gRPC client/service config có thể retry một số status. Server không nên giả định unary call chỉ đến một lần. Transfer cần idempotency key trong request/metadata và durable application handling.

Transparent retry có thể xảy ra trước server app nhận call; application-level retry vẫn cần contract riêng.

## 13. Testing

- mapper unit test;
- server test với fake use case;
- in-process/bufconn transport test cho serialization/interceptor/status;
- compatibility test descriptor/buf breaking;
- integration test auth/deadline;
- load test streaming/backpressure.

Fake generated request không chứng minh HTTP/2 deployment; end-to-end qua real listener/proxy bổ sung.

## 14. Production scenario

Client deadline 200ms, DB p99 180ms, proxy/interceptor thêm 40ms:

- nhiều calls bị cancel gần commit;
- client retry;
- duplicate transfer nếu không idempotent;
- metrics status DeadlineExceeded che DB saturation.

Timeout budget phải phân bổ end-to-end và trace từng span.

## 15. Debug

1. grpc status/code/details;
2. deadline remaining ở server ingress;
3. metadata propagation/auth;
4. mapped command;
5. interceptor chain;
6. application cause;
7. proxy HTTP/2/load balancing;
8. client retry policy/idempotency key.

## 16. Khi nào không dùng gRPC?

Public browser API, simple webhook hoặc ecosystem không hỗ trợ protobuf có thể hợp HTTP/JSON hơn. gRPC đem schema/codegen/streaming hiệu quả nhưng tăng tooling, proxy/debug compatibility. Internal không tự động đồng nghĩa gRPC.

## 17. Bài tập

1. Viết proto Transfer v1 và map command.
2. Map error taxonomy sang codes.
3. Thêm optional field và test presence.
4. Thiết kế server-stream history không giữ transaction.
5. Phân tích retry + idempotency.

## 18. Mastery questions

1. Vì sao pb message không phải Entity?
2. Deadline cancel có chứng minh rollback không?
3. Interceptor nên chứa rule nào?
4. Streaming làm resource ownership khó hơn ra sao?
5. Proto scalar absence cần xử lý thế nào?
6. Retry gRPC ảnh hưởng transfer gì?

## Further reading

- gRPC Go documentation.
- Protocol Buffers language guide và updating a message type.
- gRPC status codes, deadlines, retry và health checking docs.
- Buf breaking change documentation.

## Quality gate

- [x] Dependency/model mapping
- [x] Unary/error/deadline/auth/interceptor
- [x] Streaming, schema evolution, retry
- [x] Tests, production scenario, debugging
- [x] Trade-off, exercises, mastery
