# Hướng dẫn sử dụng & Phím tắt

## 1. Điều hướng cơ bản (Navigation)

| Phím | Tác dụng |
| --- | --- |
| `Tab` | Chuyển đổi tiêu điểm giữa hai bảng (Left/Right) |
| `Mũi tên Up/Down` hoặc `j/k` | Di chuyển con trỏ |
| `Enter` | Mở thư mục hoặc Thực hiện hành động mặc định (Edit/View) |
| `Backspace` | Quay lại thư mục cha |
| `Home / End` | Nhảy nhanh về đầu hoặc cuối danh sách |
| `PgUp / PgDn` | Cuộn trang |

## 2. Thao tác tệp tin (File Operations)

| Phím | Tác dụng |
| --- | --- |
| `F3` | Xem tệp bằng `$PAGER` mặc định |
| `F4` | Chỉnh sửa tệp bằng `$EDITOR` mặc định |
| `F5` | Sao chép tệp/thư mục đã chọn sang bảng đối diện |
| `F6` | Di chuyển tệp/thư mục đã chọn sang bảng đối diện |
| `Shift + F6` | Đổi tên tệp/thư mục tại con trỏ |
| `F7` | Tạo thư mục mới (Mkdir) |
| `F8` | Xóa tệp/thư mục (có hộp thoại xác nhận) |
| `Insert` | Chọn/Bỏ chọn tệp tại con trỏ |
| `Shift + Up/Down` | Chọn tệp và di chuyển con trỏ |
| `Ctrl + A` hoặc `*` | Chọn tất cả các tệp trong thư mục hiện tại |
| `!` | Đảo ngược vùng chọn (Invert selection) |

Khi có tệp được chọn, thanh trạng thái (footer) sẽ hiển thị số lượng tệp đang chọn (ví dụ: `[ 3 selected ]`). Các thao tác Copy, Move, Delete sẽ áp dụng cho toàn bộ danh sách đã chọn này.

## 3. Tính năng nâng cao

- **Fuzzy Finder (`F9` hoặc `Ctrl+P`):** Tìm kiếm nhanh tệp tin trong thư mục hiện tại.
- **Bookmarks (`F2` hoặc `Ctrl+B`):** Quản lý các vị trí yêu thích.
- **Theme Picker (`Ctrl+T`):** Đổi màu giao diện tức thì.
- **Run Command (`Ctrl+R`):** Thực thi lệnh shell tại thư mục hiện hành.
- **Go To (`Ctrl+G`):** Nhảy nhanh đến một đường dẫn hoặc URI SFTP.

## 4. Cấu hình (Configuration)

File cấu hình chính nằm tại: `~/.config/mdc/config.toml` (Linux/macOS).
Bạn có thể tùy chỉnh phím tắt và hành động khi nhấn `Enter` tại đây. Tham khảo `config.example.toml` trong root dự án.
