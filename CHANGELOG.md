# Nhật ký thay đổi (Changelog)

Tất cả các thay đổi đáng chú ý của dự án này sẽ được ghi lại trong tệp này. Định dạng dựa trên [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.4.0] - 2026-09-08
### Added
- **Local Navigation:** Thêm Quick View, ẩn/hiện dotfiles, đồng bộ vị trí hai panel, mở terminal tại thư mục hiện tại và trả active path về shell bằng `mdc -r`.
- **Advanced Selection:** Hỗ trợ chọn vùng, chọn tất cả, đảo lựa chọn và chọn/bỏ chọn theo shell glob.
- **Shell Workflow:** Bổ sung command/path completion, copy path qua clipboard và các thao tác điều hướng thân thiện với shell.
- **Key Input Diagnostics:** Thêm `mdc --debug-keys` để ghi chính xác input terminal phục vụ chẩn đoán phím chức năng.

### Changed
- **TUI Performance:** Giảm tải render, giới hạn FPS và rút ngắn thời gian chờ truy vấn màu nền terminal để khởi động nhanh hơn.
- **macOS Function Keys:** Xử lý ổn định `Shift+Fn+F6` cho Rename trong khi giữ `F6` cho Move và `Ctrl+K` cho Remote Profiles.
- **Project Identity:** Chuẩn hóa module, tài liệu và pipeline phát hành về fork `tungpastry/MiddayCommander`.

### Fixed
- Sửa các trường hợp selection bị mất hoặc chọn sai khi di chuyển con trỏ.
- Sửa xung đột sequence phím chức năng trên macOS Terminal.

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
