# Bảo mật & vận hành

## 1) Nguyên tắc

- Plaintext là rủi ro: chỉ lưu để bridge legacy.
- Tách quyền theo vai trò:
  - OpenIG chỉ được `read` `legacy-cred/creds/*`
  - Rotation job chỉ được `update` `legacy-cred/rotate/*`
  - Admin mới được `update/read` `legacy-cred/config`
- TTL plaintext nên ngắn; hiện default 3h theo yêu cầu.

## 2) Khuyến nghị vận hành

- Dùng Vault audit device để audit truy cập `creds/*`
- Ưu tiên dùng Vault response wrapping khi OpenIG đọc plaintext:
  - Gửi header `X-Vault-Wrap-TTL` từ OpenIG (nếu tích hợp được)
- Cân nhắc `one_time_read=true` nếu flow đảm bảo OpenIG chỉ cần dùng 1 lần

## 3) Rủi ro cần chấp nhận trong POC

- Endpoint `rotate` hiện trả thẳng `plaintext` trong response (phục vụ test). Khi chuyển sang môi trường nghiêm ngặt, nên:
  - chỉ trả plaintext qua wrapped response
  - hoặc tách endpoint chỉ dành riêng cho OpenIG

## 4) Rotation trigger

- Plugin không tự lên lịch.
- Rotation nên trigger từ:
  - cronjob
  - CI pipeline
  - Vault Agent template + exec
  - hoặc hệ thống orchestration khác
