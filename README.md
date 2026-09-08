# Midday Commander

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-blue.svg)
![Status](https://img.shields.io/badge/Status-Active-success.svg)
[![CI](https://github.com/tungpastry/MiddayCommander/actions/workflows/ci.yml/badge.svg)](https://github.com/tungpastry/MiddayCommander/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/tungpastry/MiddayCommander)](https://github.com/tungpastry/MiddayCommander/releases/latest)

**Midday Commander (`mdc`)** là một trình quản lý tệp trên terminal hiện đại với giao diện hai bảng (dual-panel), được viết bằng Go và lấy cảm hứng từ Midnight Commander cổ điển.

Dự án giữ nguyên luồng công việc điều khiển bằng bàn phím quen thuộc, đồng thời bổ sung các tính năng hiện đại như duyệt tệp nén, tìm kiếm mờ, hệ thống theme phong phú và khả năng duyệt tệp từ xa qua SFTP an toàn.

![Midday Commander](images/sc_general.png)

---

## ✨ Tính năng cốt lõi (USP)

- **Dual-panel Efficiency:** Duyệt tệp đồng thời giữa Local, Remote (SFTP) và Archive.
- **Native Remote Power:** Kết nối SFTP an toàn, hỗ trợ SSH Agent, quản lý profiles và xác minh host key nghiêm ngặt.
- **Advanced Transfer Queue:** Quản lý hàng đợi sao chép/di chuyển chạy nền với khả năng tạm dừng/tiếp tục và kiểm soát lỗi.
- **Rich Aesthetic:** Hệ thống Live Theme Picker (TOML) cho phép xem trước giao diện ngay lập tức.
- **Community Ready:** Kiến trúc mở rộng dễ dàng, cấu hình phím tắt linh hoạt.
- **Local Workflow:** Quick View, ẩn/hiện dotfiles, đồng bộ vị trí hai panel và copy path qua clipboard.
- **Shell Integration:** Completion cho command/path, mở terminal tại thư mục hiện tại và `mdc -r` để shell nhận lại active path.
- **Advanced Selection:** Chọn vùng, chọn tất cả, đảo lựa chọn và chọn/bỏ chọn theo shell glob.

---

## 🚀 Khởi động nhanh

### 1. Cài đặt
Nếu bạn dùng macOS và đã cài đặt Homebrew:
```bash
brew install tungpastry/apps/middaycommander
```
Hoặc tải bản dựng sẵn từ trang [Releases](https://github.com/tungpastry/MiddayCommander/releases).

Bạn cũng có thể cài trực tiếp bằng Go:
```bash
go install github.com/tungpastry/MiddayCommander@latest
```

### 2. Chạy ứng dụng
```bash
mdc
```

---

## 📚 Tài liệu chi tiết

Vui lòng tham khảo thư mục `docs/` để biết thêm thông tin:

- [**Cài đặt & Build từ Source**](docs/installation.md)
- [**Hướng dẫn sử dụng & Phím tắt**](docs/usage.md)
- [**Kết nối Remote (SFTP) & Profiles**](docs/remote-access.md)
- [**Kiến trúc hệ thống (Dành cho Dev)**](docs/architecture/system-overview.md)

---

## 🗺️ Lộ trình (Roadmap)

- **Phase 1 (Hoàn thành):** Core Local Filesystem & UI Framework.
- **Phase 2 (Hiện tại):** Native SFTP, Profiles & Transfer Manager Queue.
- **Phase 3 (Kế hoạch):** Secret stores gốc (Platform-native) & Xác thực mật khẩu SSH.
- **Phase 4 (Kế hoạch):** Archive view/edit flow & Cải thiện hiệu năng render ảnh.

---

## 🤝 Đóng góp & Cộng đồng

Chúng tôi rất hoan nghênh các Pull Request! Hãy xem [CONTRIBUTING.md](CONTRIBUTING.md) để bắt đầu.
Dự án tuân thủ [Quy tắc ứng xử](CODE_OF_CONDUCT.md) để đảm bảo môi trường cộng đồng văn minh.

---

## 📄 Giấy phép

Phân phối dưới giấy phép **MIT**. Xem tệp [LICENSE](LICENSE) để biết thêm chi tiết.

## Nguồn dự án và bảo trì

Fork này được duy trì và phát hành bởi **Tung Nguyen Thanh**
([@tungpastry](https://github.com/tungpastry)). Midday Commander bắt nguồn từ
[kooler/MiddayCommander](https://github.com/kooler/MiddayCommander); lịch sử Git
và giấy phép MIT được giữ nguyên để ghi nhận dự án upstream.
