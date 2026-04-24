# Hướng dẫn cài đặt chi tiết

Midday Commander có thể được cài đặt thông qua các trình quản lý gói hoặc biên dịch trực tiếp từ mã nguồn.

## 1. Cài đặt nhanh (Gói dựng sẵn)

### macOS
Sử dụng Homebrew để cài đặt bản cập nhật mới nhất:
```bash
brew install kooler/apps/MiddayCommander
```

### Windows / Linux
Tải tệp nhị phân (`.exe` cho Windows hoặc binary cho Linux) từ trang [Releases](https://github.com/kooler/MiddayCommander/releases). Đặt tệp vào một thư mục trong biến môi trường `PATH` của bạn.

## 2. Biên dịch từ mã nguồn (Dành cho Developer)

Yêu cầu: **Go >= 1.21**

### Các bước thực hiện:
1. Clone repo:
   ```bash
   git clone https://github.com/kooler/MiddayCommander.git
   cd MiddayCommander
   ```
2. Biên dịch:
   ```bash
   make build
   ```
   Lệnh này sẽ tạo ra tệp thực thi `mdc` ở thư mục gốc.

3. Kiểm tra:
   ```bash
   ./mdc --version
   ```

## 3. Các mục tiêu Build khác (Makefile)

- `make test`: Chạy toàn bộ unit tests.
- `make vet`: Kiểm tra lỗi tĩnh mã nguồn.
- `make run`: Biên dịch và chạy ứng dụng ngay lập tức.
- `make clean`: Xóa các file build tạm thời.
