# Solution

Solution tốt có:

```text
kafka message -> adapter DTO -> use case command -> application/domain
```

Producer nên implement port như `TransferEventPublisher`, không để use case import Kafka client.
