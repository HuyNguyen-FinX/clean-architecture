# Solution

Solution tốt chứng minh boundary bằng test:

- Domain test chỉ tạo object.
- Use case test dùng fake repository.
- HTTP test dùng `httptest`.
- Repository test dùng database thật hoặc testcontainer.
