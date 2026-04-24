# Hướng dẫn đóng góp (Contributing Guide)

Chào mừng bạn đến với dự án Midday Commander! Chúng tôi rất cảm kích vì bạn đã quan tâm đến việc cải thiện ứng dụng này.

## Quy trình phát triển

1. **Fork** repository này về tài khoản của bạn.
2. **Clone** bản fork về máy cá nhân.
3. Tạo một **Branch** mới cho tính năng hoặc lỗi bạn muốn xử lý:
   - `feature/ten-tinh-nang` cho tính năng mới.
   - `bugfix/mo-ta-loi` cho sửa lỗi.
4. Thực hiện thay đổi mã nguồn.
5. Đảm bảo mã nguồn tuân thủ các tiêu chuẩn:
   - Chạy `make test` để kiểm tra logic.
   - Chạy `make vet` để kiểm tra phân tích tĩnh.
   - Chạy `make build` để đảm bảo dự án vẫn biên dịch thành công.
6. **Commit** thay đổi của bạn (khuyến khích dùng [Conventional Commits](https://www.conventionalcommits.org/)).
7. **Push** lên branch của bạn và tạo một **Pull Request** (PR) về nhánh chính của dự án gốc.

## Tiêu chuẩn viết mã (Coding Standards)

- Dự án sử dụng ngôn ngữ Go với framework Bubble Tea (TUI).
- Vui lòng viết code theo phong cách idiomatic Go (`gofmt`).
- Mọi tính năng mới cần có unit test đi kèm trong thư mục tương ứng.

Cảm ơn bạn đã đóng góp!
