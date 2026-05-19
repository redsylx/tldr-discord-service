ini adalah micro project
di mana akan berjalan dengan input

file : lokasi file -> untuk dicari ke google cloud storage
type : success | failed -> penanda sebuah proses success atau gagal

alur kerjanya adalah
jika success
sistem akan mencari file (expected to be markdown atau txt) dari google cloud storage
kemudian membuat batching / bahkan jika memungkinkan stream only N baris
yang kemudian dikirim ke discord webhook as simple message
tiap pengiriman akan diberi jeda sekitar 2 detik agar tidak melampaui rate limit discord

jika failed
sistem akan mencari file (expected to be json) dari gcs
kemudian mengirim embeded message ke discord webhook
data embed yang dibutuhkan berupa
- processName   : nama proses yang gagal
- traceInfo     : objek (emailId, newsTitle, newsUrl)
- message       : error message

---

trigger untuk logic tersebut ada dua
pertama dari queue (gcp service)
kedua dari rest api (untuk keperluan testing)

rencana deploy ke googel cloud run!
pastikan opsi queue dapat memanggil cloud run yang dalam keadaan zero instance
proses ini harus single instance, karena urutan pengiriman sangat penting!