# Nhật ký thay đổi (Changelog)

Tất cả các thay đổi đáng chú ý của dự án này sẽ được ghi lại trong tệp này. Định dạng dựa trên [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.3.1] - 2026-04-08
### Added
- **Native SFTP Navigation:** Hỗ trợ duyệt tệp từ xa với xác thực SSH Agent và Identity Key.
- **Transfer Manager:** Hệ thống hàng đợi sao chép/di chuyển chạy nền (background queue).
- **Queue Controls:** Cho phép tạm dừng, tiếp tục và hủy các tác vụ chuyển tệp.
- **Image Preview:** Xem trước hình ảnh trực tiếp trong terminal bằng Half-block rendering.
- **File Sorting:** Hỗ trợ sắp xếp tệp theo tên, kích thước, thời gian.
- **Audit Logging:** Ghi lại lịch sử các giao dịch chuyển tệp qua SFTP.

## [0.2.0] - Trước đó
### Added
- Nền tảng dual-panel cơ bản.
- Hỗ trợ duyệt Archive (ZIP, TAR,...).
- Hệ thống Themes (TOML) và Live Picker.
- Bookmarks hệ thống tệp cục bộ.
