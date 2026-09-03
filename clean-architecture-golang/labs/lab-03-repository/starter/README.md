# Starter

Bắt đầu với use case đang phụ thuộc trực tiếp vào map hoặc slice trong memory. Nhiệm vụ là tách nhu cầu đó thành repository port.

Giữ interface nhỏ. Nếu use case chỉ cần `FindByID` và `Save`, đừng thêm `Delete`, `List`, `Count`.
