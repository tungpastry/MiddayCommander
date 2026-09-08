# Hướng dẫn cài đặt chi tiết

Midday Commander có thể được cài đặt thông qua các trình quản lý gói hoặc biên dịch trực tiếp từ mã nguồn.

## 1. Cài đặt nhanh (Gói dựng sẵn)

### macOS
Sử dụng Homebrew để cài đặt bản cập nhật mới nhất:
```bash
brew install tungpastry/apps/middaycommander
```

### Windows / Linux
Tải archive phù hợp từ trang [Releases](https://github.com/tungpastry/MiddayCommander/releases), giải nén và đặt `mdc` (`mdc.exe` trên Windows) vào một thư mục trong biến môi trường `PATH`.

## 2. Biên dịch từ mã nguồn (Dành cho Developer)

Yêu cầu: **Go >= 1.26**

### Các bước thực hiện:
1. Clone repo:
   ```bash
   git clone https://github.com/tungpastry/MiddayCommander.git
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

Ngoài ra có thể cài trực tiếp từ Go module chính thức:

```bash
go install github.com/tungpastry/MiddayCommander@latest
```

## 3. Các mục tiêu Build khác (Makefile)

- `make test`: Chạy toàn bộ unit tests.
- `make vet`: Kiểm tra lỗi tĩnh mã nguồn.
- `make run`: Biên dịch và chạy ứng dụng ngay lập tức.
- `make clean`: Xóa các file build tạm thời.
