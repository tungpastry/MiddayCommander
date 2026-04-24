# Kiến trúc hệ thống Midday Commander

Tài liệu này dành cho các lập trình viên muốn tìm hiểu sâu về cách Midday Commander hoạt động.

## 1. Framework & UI (Bubble Tea)

Ứng dụng được xây dựng trên mô hình **Model-Update-View (MUV)**:
- **Model:** Lưu trữ trạng thái ứng dụng (danh sách file, vị trí con trỏ, trạng thái kết nối).
- **Update:** Xử lý các tin nhắn (`tea.Msg`) như sự kiện bàn phím, kết quả từ hệ thống tệp hoặc cập nhật tiến trình từ Transfer Manager.
- **View:** Render giao diện dựa trên trạng thái hiện tại bằng thư viện `lipgloss`.

## 2. Lớp trừu tượng hệ thống tệp (Filesystem Abstraction)

Trung tâm của ứng dụng là `internal/fs/router.go`. Mọi thao tác đều thông qua interface `FileSystem`:
- **URI-based:** Mỗi vị trí được định danh bằng một URI (vd: `file:///home`, `sftp://user@host`, `archive://path/to/zip`).
- **Router:** Tự động điều hướng yêu cầu đến đúng driver (`local`, `sftp`, hoặc `archive`).

## 3. Transfer Manager (Background Worker)

Các tác vụ tệp tin nặng (Remote Copy/Move) không chạy trên luồng giao diện chính:
- **Job Queue:** Các tác vụ được đóng gói thành các `Job` và đưa vào hàng đợi.
- **Worker Pool:** Một luồng chạy nền sẽ xử lý lần lượt các Job.
- **Messaging:** Worker gửi các tin nhắn tiến trình (`transferProgressMsg`) về vòng lặp Update của Bubble Tea để cập nhật UI.

## 4. Xử lý Bàn phím (Kitty Protocol)

Để hỗ trợ các tổ hợp phím phức tạp (như phát hiện phím Shift độc lập), Midday Commander tích hợp giao thức bàn phím Kitty và các kỹ thuật polling phím bổ trợ đặc thù cho từng nền tảng trong `internal/platform`.
