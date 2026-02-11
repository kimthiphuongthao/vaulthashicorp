# Mô hình triển khai plugin Vault (so sánh với plugin Keycloak)

## 1. Tổng quan mô hình plugin Vault

- **Vault plugin** là một binary (file thực thi, ví dụ: `vault-plugin-legacy-cred`) chứ không phải file JAR như Keycloak.
- Plugin được build riêng (Go binary), không đóng gói vào Vault core.
- Khi triển khai:
  1. **Copy file binary plugin** vào thư mục `/vault/plugins` trên máy/chứa Vault server (hoặc mount volume nếu dùng Docker).
  2. **Khai báo plugin** với Vault qua API/CLI (`vault plugin register ...`).
  3. **Enable plugin** như một secrets engine/mount point (`vault secrets enable -path=legacy-cred vault-plugin-legacy-cred`).
  4. Vault sẽ chạy plugin này như một process con (subprocess), giao tiếp qua RPC (socket).
- Plugin có thể được update, reload mà không cần build lại Vault core.

## 2. So sánh với mô hình plugin Keycloak

| Tiêu chí                | Plugin Keycloak (JAR)         | Plugin Vault (Go binary)         |
|------------------------|-------------------------------|----------------------------------|
| Định dạng              | .jar (Java)                   | Binary (Go, ELF/EXE)             |
| Cách nạp               | Copy vào thư mục `standalone/deployments` hoặc `providers` | Copy vào `/vault/plugins`        |
| Cách khai báo          | Tự động nạp khi restart Keycloak | Đăng ký qua API/CLI, enable mount |
| Cách chạy              | Nạp vào JVM, cùng process     | Chạy riêng (subprocess), RPC     |
| Update plugin          | Thay file JAR, restart        | Thay file binary, reload plugin  |
| Tích hợp               | Java SPI                      | Vault plugin API (Go)            |

## 3. Sơ đồ mô hình plugin Vault

```
+-------------------+
|   Vault Server    |
|-------------------|
| - Core            |
| - Plugin Manager  |
+-------------------+
         |
         | (register/enable)
         v
+-----------------------------+
|  /vault/plugins (host/vol)  |
|  - vault-plugin-legacy-cred |
+-----------------------------+
         |
         | (run as subprocess, RPC)
         v
+-----------------------------+
|  Plugin Process (Go binary) |
+-----------------------------+
```

- Vault core quản lý lifecycle plugin, giao tiếp qua RPC.
- Plugin có thể được update, reload độc lập với Vault core.

## 4. Tài liệu tham khảo
- [Vault Plugin System](https://developer.hashicorp.com/vault/docs/plugins)
- [Vault Plugin Deployment](https://developer.hashicorp.com/vault/docs/plugins/deployment)

---

# Tóm tắt
- Plugin Vault là binary, chạy riêng, quản lý qua API/CLI, không cần build lại Vault core.
- Khác với Keycloak (JAR, nạp vào JVM), Vault plugin chạy như process con, giao tiếp qua RPC.
- Dễ update, dễ quản lý, phù hợp cho các hệ thống cần mở rộng động.
