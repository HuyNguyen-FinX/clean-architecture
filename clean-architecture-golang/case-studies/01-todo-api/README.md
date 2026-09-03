# Case Study 01: Todo API

Todo API dùng để phân tích khi nào Clean Architecture đầy đủ là over-engineering.

Trọng tâm:

- CRUD đơn giản vs behavior thật như `MarkCompleted`.
- Khi nào tách DTO/Entity là cần thiết.
- Khi nào chỉ cần package đơn giản.
- Cách tăng abstraction khi rule tăng.

Kết luận chính: với domain nhỏ, architecture nên tối giản nhưng vẫn giữ handler không quá dày và dependency dễ đổi.
