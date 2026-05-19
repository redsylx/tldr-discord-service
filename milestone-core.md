# Milestone: Core Implementation

## Ringkasan

Implementasi semua komponen inti: **model → GCS → Discord → handler → wiring → test**.

Prioritas: **success flow dulu, baru failed flow.** Setiap komponen di-unit-test.

---

## Langkah-Langkah

### 1. Model (`internal/model/types.go`)

Buat semua shared types:

- **`Config`** — semua env var
- **`WebhookParams`** — `File`, `Type`
- **`FailedPayload`** — `ProcessName`, `TraceInfo` (`EmailId`, `NewsTitle`, `NewsUrl`), `Message`
- **`ValidationError`** — untuk validasi payload failed
- **Discord embed types**: `Embed` (Title, Color, Fields), `Field` (Name, Value, Inline)

Function **`ValidateFailedPayload(payload) error`** — return error kalau field wajib kosong.

### 2. GCS Reader (`internal/gcs/reader.go`)

Interface-based agar bisa di-mock:

```go
type Reader interface {
    ReadFile(ctx context.Context, path string) ([]byte, error)
    ReadTextLines(ctx context.Context, path string) ([]string, error)
    ReadJSON(ctx context.Context, path string, dst any) error
}
```

Struct **`gcsReader`** — implementasi pakai `storage.Client`.

### 3. Discord Client (`internal/discord/client.go`)

Interface-based:

```go
type Client interface {
    SendTextMessage(ctx context.Context, content string) error
    SendEmbed(ctx context.Context, embed model.Embed) error
}
```

Struct **`discordClient`** — HTTP POST ke webhook URL. Delay antar kirim via `time.Sleep`.

### 4. Handler (`internal/handler/webhook.go`)

- **`NewHandler(reader, discordClient, cfg)`**
- **`ServeHTTP(w, r)`** — parse query params, validasi
- success flow: `ReadTextLines` → batch per `BATCH_LINE_COUNT` → `SendTextMessage` tiap batch dengan delay
- failed flow: `ReadJSON` → `ValidateFailedPayload` → buat Embed merah → `SendEmbed`

### 5. Wiring (`cmd/main.go`)

Update `cmd/main.go`:

- Import `cloud.google.com/go/storage`
- Inisialisasi `storage.Client` di `main()`
- Inisialisasi GCS reader, Discord client, handler
- Registrasi handler ke mux (ganti placeholder)

### 6. Unit Tests

| Package | Approach |
|---|---|
| `internal/gcs` | Mock `storage.Client` pake fake GCS server atau interface mock |
| `internal/discord` | Test server (httptest) untuk verifikasi payload ke webhook |
| `internal/handler` | Mock GCS reader + Discord client, test success & failed flow |
| `internal/model` | Test `ValidateFailedPayload` — valid & invalid cases |

### 7. Build & Verify

```bash
go build ./cmd/
go test ./... -v
```

Validasi struktur project:

```
cmd/main.go
internal/
  model/types.go
  gcs/reader.go (+ gcs/reader_test.go)
  discord/client.go (+ discord/client_test.go)
  handler/webhook.go (+ handler/webhook_test.go)
```

---

## Struktur File

### Files baru yang dibuat:

| File | Isi |
|---|---|
| `internal/model/types.go` | Config, WebhookParams, FailedPayload, Embed, Field, Validator |
| `internal/gcs/reader.go` | Interface + implementation GCS Reader |
| `internal/gcs/reader_test.go` | Unit test GCS reader |
| `internal/discord/client.go` | Interface + implementation Discord Client |
| `internal/discord/client_test.go` | Unit test Discord client |
| `internal/handler/webhook.go` | Handler struct + ServeHTTP + success/failed flow |
| `internal/handler/webhook_test.go` | Unit test handler |

### Files yang di-update:

| File | Perubahan |
|---|---|
| `cmd/main.go` | Wiring: init GCS, Discord, handler, ganti placeholder |
| `go.mod` | Tambah dependency `cloud.google.com/go/storage` |

---

## Catatan

- **Mocking strategy**: GCS reader & Discord client pakai interface, test handler dengan mock implementation.
- **Failed embed**: Color `0xE74C3C` (merah), Title `❌ Process Failed: {processName}`, fields: Email ID, News Title, News URL, Error Message.
- **Validasi failed payload**: `processName`, `traceInfo.emailId`, `traceInfo.newsTitle`, `message` wajib ada. `newsUrl` opsional.