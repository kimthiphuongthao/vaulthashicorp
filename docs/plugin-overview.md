# Tổng quan plugin `legacy-cred`

## Bài toán

- Legacy app **không sửa được mã nguồn** nhưng cần đưa vào SSO tập trung.
- Keycloak là IdP, OpenIG là enforcement point.
- OpenIG cần lấy `subject` (email/username) từ JWT, và cần một **plaintext password** tương ứng để đăng nhập legacy app.
- Đồng thời cần **rotate password** định kỳ/ theo lệnh: sinh plaintext mới → băm theo thuật toán legacy yêu cầu → cập nhật credential store của legacy.

## Giải pháp plugin Vault

Plugin này là một **Vault secrets engine plugin** cung cấp:

- Sinh plaintext ngẫu nhiên
- Băm theo thuật toán cấu hình (mặc định: `bcrypt`)
- Update legacy credential store theo 2 cách:
  - `webhook` (khuyến nghị để thử nghiệm đa DB)
  - `sql` (kết nối DB trực tiếp)
- Lưu plaintext trong storage của Vault theo `subject` với TTL (mặc định 3h)

## Thành phần chính trong repo

- Backend & routes: `internal/legacycred`
- Hashing: `internal/legacycred/hash`
- Updater: `internal/legacycred/updater`
- Binary entrypoint: `cmd/vault-plugin-legacy-cred`

## Lưu ý thiết kế

- Plugin không tự chạy scheduler background; rotation nên được trigger từ cron/job bên ngoài hoặc hệ thống orchestration.
- Lưu plaintext là rủi ro bảo mật — plugin này chỉ dành cho kịch bản “bridge” legacy; cần policy chặt + TTL ngắn.
