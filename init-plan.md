# Init Plan — TL;DR Discord Service

## Informasi Proyek

| Item | Value |
|---|---|
| Module path | `github.com/redsylx/tldr-discord-service` |
| Go version | 1.26 |
| Framework | `net/http` standar |

## Langkah Inisialisasi

1. **`go mod init github.com/redsylx/tldr-discord-service`**
2. **Buat direktori project:**
   ```
   cmd/
   internal/
     handler/
     gcs/
     discord/
     model/
   ```
3. **Tulis `cmd/main.go`:**
   - Baca env var
   - Inisialisasi GCS client via ADC
   - Setup HTTP server (`POST /webhook`)
   - Graceful shutdown
4. **Tulis `go.mod` + `go.sum`** setelah `go mod tidy`
5. **Tulis `Dockerfile`** — multi-stage: `golang:1.26-alpine` → `gcr.io/distroless/static-debian12`
6. **Tulis `.gitignore`** — ignore binary, IDE, `.env`
7. **Init git commit** — `chore: initial project setup`

> **Catatan:** Struct, handler, GCS reader, Discord client — semuanya tetap ditulis di `cmd/main.go` sampai kebutuhan pemisahan modul muncul. Tidak ada file Go lain yang dibuat selain `cmd/main.go` pada tahap ini.