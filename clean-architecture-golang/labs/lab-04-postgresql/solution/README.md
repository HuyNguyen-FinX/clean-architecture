# Solution

Solution mong muốn:

```text
application -> domain
postgres adapter -> application/domain
```

Repository adapter chịu trách nhiệm query, scan row, map sang domain entity và map lỗi not found sang domain/application error phù hợp.
