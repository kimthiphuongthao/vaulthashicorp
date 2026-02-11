# Tài liệu plugin Vault (legacy-cred)

## AI quick context
- [ai-context.md](ai-context.md)

Bộ tài liệu này chỉ tập trung vào **nhiệm vụ chính: code và vận hành thử plugin Vault** phục vụ tích hợp SSO (Keycloak/OpenIG) với legacy app.

## Mục lục

- [Tổng quan](plugin-overview.md)
- [API & Endpoints](api.md)
- [Hashing & Updater](hashing-and-updater.md)
- [Bảo mật & vận hành](security-and-ops.md)
- [Build & Test](build-and-test.md)
- [Test bằng curl](curl-test.md)
- [Demo thủ công SQL updater (Postgres/MySQL/MSSQL)](manual-demo-sql-updater-standard-db.md)

## TL;DR (luồng chính)

1) Admin cấu hình plugin: `legacy-cred/config`
2) Job/automation gọi rotate: `legacy-cred/rotate/<subject>`
3) OpenIG đọc plaintext để login legacy: `legacy-cred/creds/<subject>`

TTL plaintext mặc định: **3 giờ**.
