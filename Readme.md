# AppSecHub

## Luồng xử lý khi gọi API (ví dụ /api/v1/auth/login):

```txt
HTTP Request (Gin)
    ↓
Handler (UserHandler)
    ↓
DTO chuyển từ HTTP sang UseCase
    ↓
UseCase (UserUsecase)
    ↓
Gọi đến Repository Interface (UserRepository)
    ↓
Triển khai thật của Repository (Postgres)
    ↓
Domain Entity / ValueObject xử lý nghiệp vụ (validation, hash password...)
    ↓
Trả kết quả (token / lỗi)
    ↓
Handler trả HTTP Response
```

