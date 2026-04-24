# Hướng dẫn Kết nối từ xa (SFTP)

Midday Commander hỗ trợ kết nối và thao tác tệp trên máy chủ từ xa thông qua giao thức SFTP một cách mượt mà như tệp cục bộ.

## 1. Cách kết nối

Có 3 cách để mở một vị trí từ xa:
1. **Dùng Profiles (`Shift + F2` hoặc `Ctrl + K`):** Chọn từ danh sách máy chủ đã cấu hình sẵn.
2. **Manual Connect:** Nhấn `n` trong bảng Profile để nhập thông tin Host, User, Port thủ công.
3. **Go To URIs (`Ctrl + G`):** Nhập trực tiếp địa chỉ dạng `sftp://user@host:port/path`.

## 2. Quản lý Profiles

Các máy chủ thường xuyên truy cập nên được lưu vào tệp:
`~/.config/mdc/profiles.toml`

**Ví dụ cấu hình:**
```toml
[[profiles]]
name = "Production Server"
host = "1.2.3.4"
user = "admin"
auth = "agent"  # Sử dụng ssh-agent
known_hosts_file = "~/.ssh/known_hosts"

[[profiles]]
name = "Backup Lab"
host = "lab.local"
user = "nexus"
auth = "key"    # Sử dụng Private Key
identity_file = "~/.ssh/id_ed25519"
```

## 3. Quản lý Chuyển tệp (Transfer Manager)

Khi bạn copy hoặc move tệp giữa bảng Local và bảng Remote:
1. Một hộp thoại **Transfer Options** sẽ hiện ra để bạn chọn:
   - Xử lý xung đột (Overwrite/Skip).
   - Chế độ xác minh (None/Size/SHA256).
   - Số lần thử lại (Retry).
2. Tác vụ được đưa vào hàng đợi chạy nền.
3. Bạn có thể nhấn `p` để tạm dừng, `c` để hủy tác vụ đang chạy trong bảng điều khiển Transfer.

## 4. Bảo mật

- Dự án xác minh nghiêm ngặt `host keys`. Nếu host key thay đổi hoặc không có trong `known_hosts`, kết nối sẽ bị từ chối để đảm bảo an toàn.
- Hiện tại chưa hỗ trợ nhập mật khẩu trực tiếp (Password Auth). Khuyến khích sử dụng SSH Agent.
