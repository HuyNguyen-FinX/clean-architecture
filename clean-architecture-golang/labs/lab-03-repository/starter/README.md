# Starter

Baseline chạy được nhưng Service phụ thuộc concrete MemoryStore và store trả pointer alias. Test khóa lại behavior lỗi để bạn nhìn thấy vấn đề trước khi refactor.

~~~bash
go test ./...
~~~

Sau khi sửa, test PointerAlias cần được đổi kỳ vọng: state trong store chỉ thay đổi sau Save.
