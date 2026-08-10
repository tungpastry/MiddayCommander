# Hướng Dẫn Sử Dụng MiddayCommander

MiddayCommander (`mdc`) là trình quản lý tệp trong terminal theo kiểu hai bảng
(dual-panel). Bạn có thể duyệt hai vị trí cùng lúc, chọn nhiều tệp, copy/move
sang bảng đối diện, mở SFTP, duyệt archive, tìm file, đổi theme và chạy lệnh
shell ngay trong TUI.

Tài liệu này mô tả hành vi hiện tại của ứng dụng. Một số tính năng trong
roadmap có thể đã xuất hiện trong README, nhưng nếu chưa được ghi ở đây thì
chưa nên xem là workflow ổn định cho người dùng.

## 1. Khởi Chạy

```bash
mdc
```

Kiểm tra phiên bản:

```bash
mdc --version
# hoặc
mdc -v
```

Bật chế độ debug phím, hữu ích khi terminal/macOS gửi phím khác với mong đợi:

```bash
mdc --debug-keys
tail -n 80 ~/.config/mdc/key-debug.log
```

`--debug-keys` chỉ ghi log JSONL vào file, không hiện text trong giao diện TUI.

## 2. Tổng Quan Giao Diện

- Màn hình chính có hai panel trái/phải. Panel đang active có border/header nổi bật hơn.
- Mỗi panel hiện đường dẫn ở header, danh sách tệp ở giữa, footer panel ở dưới.
- Dòng menu cuối màn hình hiện nhóm phím `F1` đến `F10`.
- Khi giữ `Shift`, menu cuối màn hình đổi sang nhóm `Shift+F1` đến `Shift+F10` nếu terminal gửi được trạng thái Shift.
- Các hộp nổi như Help, Bookmarks, Profiles, Transfer, Theme Picker, Fuzzy Finder và dialog xác nhận sẽ ưu tiên nhận phím trước panel.
- Nhấn `Esc` hai lần liên tiếp trong khoảng ngắn để thoát app. `F10` hoặc `Ctrl+C` cũng thoát app.

## 3. Điều Hướng Cơ Bản

| Phím | Tác dụng |
| --- | --- |
| `Tab` | Đổi active panel trái/phải |
| `Ctrl+U` | Hoán đổi vị trí hai panel |
| `Up`, `k` | Di chuyển lên |
| `Down`, `j` | Di chuyển xuống |
| `PgUp` | Lên một trang |
| `PgDn` | Xuống một trang |
| `Home` | Về đầu danh sách |
| `End` | Về cuối danh sách |
| `Backspace` | Đi lên thư mục cha |
| `Ctrl+G` | Mở hộp Go To để nhập path/URI |
| `Ctrl+O` | Đổi chế độ sắp xếp |
| `F1` | Mở Help |
| `F10`, `Ctrl+C` | Thoát app |

Chế độ sắp xếp luân phiên theo `Name`, `Size`, `Time`, `Ext`. Thư mục được ưu
tiên hiện trước tệp; dòng `..` luôn nằm ở đầu nếu có thư mục cha.

### Enter Và Space

- `Enter` trên `..`: về thư mục cha.
- `Enter` trên thư mục: mở thư mục.
- `Enter` trên archive local: mở archive như một thư mục read-only.
- `Enter` trên tệp: thực hiện `behavior.enter_action`, mặc định là `edit`.
- `Space` trên tệp: thực hiện `behavior.space_action`, mặc định là `preview`.

`preview` với tệp local sẽ dùng `$PAGER`, mặc định `less`. Ảnh local có đuôi
`.jpg`, `.jpeg`, `.png`, `.gif` sẽ mở image preview nội bộ. `edit` dùng
`$EDITOR`, mặc định `vi`.

## 4. Tìm Nhanh Trong Panel

Có hai cách tìm trong panel hiện tại:

- Nhấn `Ctrl+S` để vào Quick Search, sau đó gõ chuỗi cần tìm.
- Hoặc gõ trực tiếp một ký tự hợp lệ (`a-z`, `A-Z`, `0-9`, `.`, `_`, `-`) khi panel đang active.

Quick Search tìm theo tiền tố tên file, không phân biệt hoa thường, bắt đầu từ
vị trí con trỏ hiện tại rồi quay vòng lên đầu danh sách.

Trong Quick Search:

| Phím | Tác dụng |
| --- | --- |
| Ký tự hợp lệ | Thêm vào chuỗi tìm |
| `Backspace` | Xóa một ký tự |
| `Esc` | Thoát tìm kiếm |
| Phím khác | Thoát tìm kiếm và xử lý phím đó như bình thường |

## 5. Chọn Tệp Và Thư Mục

| Phím | Tác dụng |
| --- | --- |
| `Insert`, `Ctrl+Space` | Chọn/bỏ chọn mục tại con trỏ |
| `Shift+Up` | Mở rộng vùng chọn lên trên |
| `Shift+Down` | Mở rộng vùng chọn xuống dưới |
| `Ctrl+A`, `*` | Chọn tất cả mục trong thư mục hiện tại |
| `!` | Đảo ngược vùng chọn |

Khi có mục được chọn, footer panel sẽ hiện số lượng mục đang chọn. Các lệnh
`Copy`, `Move`, `Delete` sẽ áp dụng cho toàn bộ vùng chọn. Nếu không có mục nào
được chọn, lệnh sẽ áp dụng cho mục đang đặt con trỏ.

Lưu ý: default trong code chỉ gán `toggle_select` cho `Insert` và
`Ctrl+Space`. File `config.example.toml` có thêm `x` như một ví dụ tùy biến;
nếu muốn dùng `x`, hãy thêm nó vào config của bạn.

## 6. Thao Tác Tệp

| Phím | Tác dụng |
| --- | --- |
| `F3` | Preview/xem tệp |
| `F4` | Edit/sửa tệp |
| `F5` | Copy sang thư mục của panel đối diện |
| `F6` | Move sang thư mục của panel đối diện |
| `Shift+F6` | Rename mục tại con trỏ |
| `F7` | Tạo thư mục mới |
| `F8` | Xóa mục đang chọn hoặc mục tại con trỏ |

### Dialog Xác Nhận Và Nhập Liệu

- Dialog xác nhận: `Y`, `Enter` để đồng ý; `N`, `Esc` để hủy.
- Dialog nhập liệu: gõ text, dùng `Left`, `Right`, `Home`, `End`,
  `Backspace`, `Delete`; `Enter` để xác nhận; `Esc` để hủy.
- Dialog lỗi: `Enter`, `Esc` hoặc `q` để đóng.

### Copy Và Move

- Copy/Move local-local chạy trực tiếp sau khi bạn xác nhận.
- Nếu nguồn hoặc đích có SFTP, app mở hộp `Transfer Options` trước khi đưa tác vụ vào hàng đợi transfer.
- Đích của Copy/Move luôn là thư mục đang mở trong panel đối diện.
- Nếu đang chọn nhiều mục, app copy/move tất cả mục đã chọn.

### Rename

`Shift+F6` mở dialog `Rename` với tên hiện tại làm giá trị mặc định. Rename
không áp dụng cho dòng `..`. Trên Mac, nếu `Shift+F6` không kích hoạt rename,
thử `Shift+Fn+F6` hoặc bật chế độ function keys trong thiết lập bàn phím/macOS.

## 7. Transfer Manager

Khi Copy/Move có SFTP, hộp `Transfer Options` hiện trước khi queue tác vụ.

Trong `Transfer Options`:

| Phím | Tác dụng |
| --- | --- |
| `Tab`, `Down` | Chuyển field tiếp theo |
| `Shift+Tab`, `Up` | Chuyển field trước |
| `Left`, `h` | Giảm/chọn giá trị trước |
| `Right`, `l` | Tăng/chọn giá trị tiếp theo |
| `0`-`5` | Đặt số retry khi đang ở field Retries |
| `Enter` | Đưa tác vụ vào queue |
| `Esc`, `q` | Hủy |

Tùy chọn conflict:

- `Overwrite existing destination`: xóa đích cũ và ghi file mới.
- `Skip conflicting files`: bỏ qua file trùng tên.
- `Keep both by renaming new copy`: tạo tên dạng `name (copy N).ext`.

Tùy chọn verify:

- `Size check`: xác minh kích thước sau copy.
- `SHA-256 checksum`: xác minh bằng checksum, chậm hơn nhưng chặt hơn.
- `No post-copy verification`: không xác minh sau copy.

Retries từ `0` đến `5`. Mỗi retry dùng backoff ngắn trước lần thử tiếp theo.

Trong overlay `Transfers`:

| Phím | Tác dụng |
| --- | --- |
| `p` | Pause/Resume queue. Tác vụ đang chạy có thể kết thúc, queue mới sẽ chờ |
| `c` | Hủy tác vụ đang chạy |
| `k` | Hủy/clear các tác vụ đang đợi |
| `a` | Mở Audit Log |
| `Esc`, `q` | Ẩn overlay transfer |

Transfer events được ghi vào `~/.config/mdc/audit.log`.

## 8. SFTP Remote

MiddayCommander hỗ trợ SFTP với host key verification nghiêm ngặt. Hiện tại app
hỗ trợ auth qua `ssh-agent` hoặc private key không mã hóa passphrase trong flow
TUI; không có prompt nhập password trực tiếp.

Mở remote bằng một trong các cách sau:

- `Ctrl+K` hoặc `Shift+F2`: mở Remote Profiles.
- Trong Remote Profiles nhấn `n`: manual connect.
- `Ctrl+G`: nhập URI dạng `sftp://user@host:port/path`.

### Profiles

Profiles nằm tại:

```text
~/.config/mdc/profiles.toml
```

Ví dụ:

```toml
[[profiles]]
name = "lab-agent"
host = "192.168.1.30"
port = 22
user = "nexus"
path = "/home/nexus"
auth = "agent"
known_hosts_file = "~/.ssh/known_hosts"

[[profiles]]
name = "deploy-key"
host = "files.example.com"
port = 2222
user = "deploy"
path = "/srv/releases"
auth = "key"
identity_file = "~/.ssh/id_ed25519"
known_hosts_file = "~/.ssh/known_hosts"
```

Quy tắc validate:

- `name`, `host`, `user` bắt buộc phải có.
- `port` mặc định là `22`, hợp lệ trong khoảng `1..65535`.
- `path` mặc định là `/`.
- `auth` mặc định là `agent`; chỉ hỗ trợ `agent` và `key`.
- Nếu `auth = "key"` thì `identity_file` bắt buộc phải có.
- `known_hosts_file` nếu để trống sẽ dùng `~/.ssh/known_hosts`.

Trong Remote Profiles:

| Phím | Tác dụng |
| --- | --- |
| `Up`, `k` | Đi lên |
| `Down`, `j` | Đi xuống |
| `f` | Filter profiles |
| `n` | Manual connect |
| `0`-`9` | Chọn nhanh profile theo số hiện trên danh sách |
| `Enter` | Connect profile đang chọn |
| `Esc` | Đóng |

### Manual Connect

Manual Connect có các field: `Host`, `Port`, `User`, `Path`, `Auth`,
`Identity`, `Known Hosts`.

| Phím | Tác dụng |
| --- | --- |
| `Tab`, `Down` | Field tiếp theo |
| `Shift+Tab`, `Up` | Field trước |
| `Left`, `Right` | Di chuyển con trỏ; với field `Auth` thì đổi `agent`/`key` |
| `a`, `k` | Khi đang ở field `Auth`, chọn nhanh `agent` hoặc `key` |
| `Enter` | Connect |
| `Esc` | Hủy |

### View/Edit Remote File

- Preview remote file: app download bản tạm vào cache rồi mở bằng `$PAGER`.
- Edit remote file: app download bản tạm vào cache rồi mở bằng `$EDITOR`.
- Sau khi editor đóng, nếu file thay đổi, app hỏi có upload ngược lên remote hay không.
- Nếu upload lỗi, bản tạm local được giữ lại và đường dẫn sẽ hiện trong dialog lỗi.
- Trong panel SFTP, `F7`, `Shift+F6`, `F8` vẫn dùng cho mkdir, rename, delete nếu tài khoản remote có quyền.
- Copy/Move có SFTP sẽ đi qua `Transfer Options` và `Transfers` thay vì chạy theo đường local-local trực tiếp.

Remote workfiles nằm trong cache:

```text
${XDG_CACHE_HOME}/mdc/remote
# hoặc thư mục cache mặc định của hệ điều hành + /mdc/remote
```

## 9. Archive

Khi đang ở local filesystem, `Enter` trên một archive được nhận diện bởi
`archiver` sẽ mở archive như một thư mục.

Trạng thái hiện tại của archive filesystem:

- Hỗ trợ list và read nội dung.
- Hỗ trợ di chuyển trong archive bằng `Enter` và `Backspace`.
- Có thể copy file/folder từ archive ra một đích có khả năng ghi.
- Không hỗ trợ ghi vào archive.
- Không hỗ trợ mkdir, rename, delete trong archive.
- Không hỗ trợ move từ archive vì move cần xóa nguồn sau khi copy.
- Preview/Edit trực tiếp tệp nằm bên trong archive hiện chưa được map vào pager/editor.

## 10. Bookmarks

Mở Bookmarks bằng `F2` hoặc `Ctrl+B`. Bookmark lưu local tại:

```text
~/.config/mdc/bookmarks.json
```

Danh sách bookmark được sắp xếp theo tần suất dùng và lần dùng gần đây.

| Phím | Tác dụng |
| --- | --- |
| `Up`, `k` | Đi lên |
| `Down`, `j` | Đi xuống |
| `Enter` | Mở bookmark đang chọn |
| `a` | Thêm vị trí panel hiện tại vào bookmark |
| `f` | Filter bookmarks |
| `d`, `Delete` | Xóa bookmark đang chọn |
| `0`-`9` | Chọn nhanh bookmark theo số |
| `Esc`, `Ctrl+B` | Đóng |

Khi thêm bookmark, app hỏi `Name`. Có thể để trống nếu chỉ muốn lưu path/URI.

## 11. Fuzzy Finder

Mở bằng `F9` hoặc `Ctrl+P`. Fuzzy Finder quét đệ quy từ thư mục đang mở trong
panel active. Hỗ trợ local và SFTP, không hỗ trợ archive. Giới hạn quét hiện tại
là 50.000 entries, kết quả hiển thị tối đa 1.000 matches.

| Phím | Tác dụng |
| --- | --- |
| Ký tự bất kỳ | Thêm vào query |
| `Backspace` | Xóa một ký tự |
| `Up`, `Ctrl+P` | Lên kết quả trước |
| `Down`, `Ctrl+N` | Xuống kết quả tiếp |
| `PgUp`, `PgDn` | Cuộn trang |
| `Enter` | Chọn kết quả |
| `Esc` | Đóng |

Nếu chọn thư mục, panel sẽ mở thư mục đó. Nếu chọn tệp, panel mở thư mục cha và
đặt con trỏ vào tệp đó.

## 12. Theme Picker

Mở bằng `Ctrl+T`.

- `Up`/`Down` hoặc `k`/`j`: di chuyển và preview theme ngay lập tức.
- `Enter`: áp dụng theme.
- `Esc`: hủy và quay lại theme trước khi mở picker.

Theme local nằm tại:

```text
~/.config/mdc/themes/<theme-name>.toml
```

Khi chọn theme remote, app lưu file TOML về thư mục themes local và cập nhật
`theme = "<theme-name>"` trong `~/.config/mdc/config.toml`.

## 13. Run Command

Mở bằng `Ctrl+R`. Tính năng này chỉ chạy trong local directory; nếu panel đang ở
SFTP/archive, app sẽ báo lỗi.

Command runner dùng:

```bash
sh -c "<command>"
```

và đặt working directory là thư mục local đang mở trong panel active.

| Phím | Tác dụng |
| --- | --- |
| Ký tự bất kỳ | Nhập lệnh |
| `Left`, `Right`, `Home`, `End` | Di chuyển con trỏ |
| `Backspace`, `Delete` | Sửa lệnh |
| `Enter` | Chạy lệnh |
| `Up`, `Down`, `PgUp`, `PgDn` | Cuộn output sau khi chạy |
| `Esc` | Đóng overlay |

Stdout và stderr được gom chung vào vùng output.

## 14. Help Và Audit Log

Help:

- Mở bằng `F1`.
- Đóng bằng `Esc`, `q`, `F1` hoặc `Enter`.
- Cuộn bằng `Up`/`Down` hoặc `k`/`j`.

Audit Log:

- Mở từ overlay transfer bằng phím `a`.
- Đọc các event gần đây từ `~/.config/mdc/audit.log`.
- Cuộn bằng `Up`, `Down`, `PgUp`, `PgDn`, `Home`, `End`.
- `r` để refresh.
- `Esc` hoặc `q` để đóng.

## 15. Cấu Hình

Config chính nằm tại:

```text
~/.config/mdc/config.toml
```

Nếu file không tồn tại, app dùng default trong code. Bạn có thể copy
`config.example.toml` để bắt đầu:

```bash
mkdir -p ~/.config/mdc
cp config.example.toml ~/.config/mdc/config.toml
```

Ví dụ tối thiểu:

```toml
theme = "catppuccin-mocha"

[behavior]
enter_action = "edit"
space_action = "preview"

[keys]
rename = "shift+f6"
toggle_select = ["insert", "ctrl+space", "x"]
```

Keybinding có thể là một string hoặc list string. Các key `shift+f1` đến
`shift+f8` được normalize nội bộ thành `f13` đến `f20` theo convention của
Bubble Tea.

Đường dẫn dữ liệu:

| File/Thư mục | Tác dụng |
| --- | --- |
| `~/.config/mdc/config.toml` | Config chính |
| `~/.config/mdc/profiles.toml` | Remote profiles |
| `~/.config/mdc/bookmarks.json` | Bookmarks |
| `~/.config/mdc/themes/` | Theme TOML local |
| `~/.config/mdc/audit.log` | Log transfer/audit |
| `~/.config/mdc/key-debug.log` | Log debug key input khi dùng `--debug-keys` |
| Cache `mdc/remote` | Bản tạm cho remote view/edit |

Nếu set `XDG_CONFIG_HOME`, config sẽ nằm trong `$XDG_CONFIG_HOME/mdc`. Nếu set
`XDG_CACHE_HOME`, cache sẽ nằm trong `$XDG_CACHE_HOME/mdc`.

## 16. Troubleshooting

### Shift+F6 Trên Mac Không Mở Rename

Một số bàn phím/terminal trên Mac không gửi function key khi không giữ `Fn`.
Thử theo thứ tự:

1. Dùng `Shift+Fn+F6`.
2. Kiểm tra macOS Keyboard settings để F1/F2/... hoạt động như standard function keys.
3. Kiểm tra Terminal Settings -> Profiles -> Keyboard để xem `Shift+F6` đang gửi escape sequence nào.
4. Chạy:

   ```bash
   mdc --debug-keys
   tail -n 80 ~/.config/mdc/key-debug.log
   ```

Trong log, dòng quan trọng là `received`, `effective`, `matched_action`:

- `received.string = "f6"` và `effective.string = "f18"`: app đã map Shift+F6 fallback và sẽ rename.
- `received.string = "f18"`: terminal gửi native Shift+F6, app sẽ rename.
- Không có dòng F-key nào sau khi bấm phím: macOS/terminal không gửi phím vào app; hãy dùng `Fn` hoặc đổi thiết lập bàn phím.

### `$EDITOR` Hoặc `$PAGER` Mở Sai Chương Trình

Đặt biến môi trường trước khi chạy `mdc`:

```bash
export EDITOR=nvim
export PAGER=less
mdc
```

### SFTP Báo Lỗi Host Key

MiddayCommander dùng strict host key verification. Đảm bảo host nằm trong
`~/.ssh/known_hosts` hoặc khai báo đúng `known_hosts_file` trong profile/URI.

Ví dụ thêm host key bằng OpenSSH:

```bash
ssh-keyscan -H 192.168.1.30 >> ~/.ssh/known_hosts
```

### SFTP Báo Lỗi Auth

- Với `auth = "agent"`: đảm bảo `SSH_AUTH_SOCK` đang có và key đã được add vào agent.
- Với `auth = "key"`: đảm bảo `identity_file` trỏ đúng private key.
- Password prompt trực tiếp chưa được hỗ trợ trong TUI hiện tại.

## 17. Cheat Sheet

| Nhóm | Phím |
| --- | --- |
| Help / Quit | `F1`, `F10`, `Ctrl+C`, double `Esc` |
| Panel | `Tab`, `Ctrl+U` |
| Navigate | `Up/Down`, `j/k`, `PgUp/PgDn`, `Home/End`, `Backspace` |
| Open | `Enter`, `Space` |
| View/Edit | `F3`, `F4` |
| Copy/Move | `F5`, `F6` |
| Rename/Mkdir/Delete | `Shift+F6`, `F7`, `F8` |
| Select | `Insert`, `Ctrl+Space`, `Shift+Up/Down`, `Ctrl+A`, `*`, `!` |
| Search | `Ctrl+S`, gõ ký tự trực tiếp, `F9`, `Ctrl+P` |
| Remote | `Ctrl+K`, `Shift+F2`, `Ctrl+G` |
| Tools | `F2`, `Ctrl+B`, `Ctrl+T`, `Ctrl+R`, `Ctrl+O` |
