# Panduan Setup & Testing

## 1. Prasyarat

- **GCS bucket** — untuk menyimpan file test
- **Discord webhook** — untuk menerima pesan
- **ADC (Application Default Credentials)** — autentikasi ke GCS

---

## 2. Setup GCS

```bash
# 1. Buat bucket (ganti <bucket-name>)
gcloud storage buckets create gs://<bucket-name> --location=asia-southeast1

# 2. Upload file success (contoh: hasil.txt)
cat > /tmp/hasil.txt << 'EOF'
✅ TL;DR News berhasil diproses untuk 3 subscribers.

📬 Email terkirim ke:
- user1@example.com
- user2@example.com
- user3@example.com

⏱ Waktu proses: 2.3 detik
EOF

gcloud storage cp /tmp/hasil.txt gs://<bucket-name>/hasil.txt

# 3. Upload file failed
cat > /tmp/failed.json << 'EOF'
{
  "processName": "send-newsletter",
  "traceInfo": {
    "emailId": "user@example.com",
    "newsTitle": "Berita Hari Ini",
    "newsUrl": "https://example.com/news/123"
  },
  "message": "smtp connection timeout after 30s"
}
EOF

gcloud storage cp /tmp/failed.json gs://<bucket-name>/failed.json
```

---

## 3. Discord Webhook

1. Buka Discord server → Settings → Integrations → Webhooks
2. Buat webhook baru, copy URL
3. Simpan sebagai env var `DISCORD_WEBHOOK_URL`

---

## 4. Run Lokal

```bash
export GCS_BUCKET_NAME=<bucket-name>
export DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
export PORT=8080

# Login ADC (kalau belum)
gcloud auth application-default login

# Jalankan
go run ./cmd/
```

---

## 5. Test dengan curl

### Success flow
```bash
curl -X POST "http://localhost:8080/webhook?file=hasil.txt&type=success"
```

### Failed flow
```bash
curl -X POST "http://localhost:8080/webhook?file=failed.json&type=failed"
```

### Error cases
```bash
# Missing file param → 400
curl -X POST "http://localhost:8080/webhook?type=success"

# Invalid type → 400
curl -X POST "http://localhost:8080/webhook?file=test.txt&type=invalid"

# GET instead of POST → 405
curl "http://localhost:8080/webhook?file=test.txt&type=success"

# File not found → 500 (GCS error)
curl -X POST "http://localhost:8080/webhook?file=nonexistent.txt&type=success"
```

---

## 6. (Opsional) Build & Deploy ke Cloud Run

```bash
export PROJECT_ID=<gcp-project-id>

docker build -t gcr.io/$PROJECT_ID/tldr-discord-service .
docker push gcr.io/$PROJECT_ID/tldr-discord-service

gcloud run deploy tldr-discord-service \
  --image gcr.io/$PROJECT_ID/tldr-discord-service \
  --region asia-southeast1 \
  --max-instances 1 \
  --no-cpu-throttling \
  --set-env-vars "GCS_BUCKET_NAME=<bucket>,DISCORD_WEBHOOK_URL=<webhook>"

# Test via Cloud Run URL
curl -X POST "https://<service-url>/webhook?file=hasil.txt&type=success"
```

---

## Catatan

- `BATCH_LINE_COUNT` default 5 baris per batch. Bisa diubah via env var.
- `DISCORD_DELAY_MS` default 2000ms antar kirim. Bisa diubah via env var.
- Delay hanya berlaku di success flow (multiple batches). Failed flow hanya kirim 1 embed, tidak kena delay.