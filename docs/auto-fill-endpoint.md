# Endpoint Auto-Fill Laporan (untuk tim UI/Flutter)

Endpoint ini dipakai di **alur Room Attendant lapor barang temuan**:

1. RA foto barang di kamar.
2. App kirim foto ke `POST /reports/auto-fill`.
3. Backend upload foto + analisis pakai Gemini Vision.
4. Backend balikin saran `title`, `description`, `category`, `confidence` (+ `photo_url`).
5. App pre-fill form, RA boleh edit manual sebelum submit ke `POST /reports`.

> Form pre-fill & flow submit ditangani tim UI — di luar scope dokumen ini
> (backend `POST /reports` sudah ada dan tidak berubah).

---

## 1. URL & Method

| Item      | Nilai                                        |
|-----------|----------------------------------------------|
| Method    | `POST`                                       |
| URL       | `{BASE_URL}/api/reports/auto-fill`           |
| Auth      | **Wajib** — `Authorization: Bearer <JWT>` (login dulu, ambil dari `POST /api/login`) |
| Body      | `multipart/form-data`                        |

Contoh base URL lokal: `http://localhost:8094/api/reports/auto-fill`

---

## 2. Format Request

`multipart/form-data` dengan **satu field file foto**.

| Field | Tipe       | Wajib | Keterangan |
|-------|-----------|-------|------------|
| `file` | file (image) | Ya | Foto barang. Bisa juga pakai key `image` atau `photo` (salah satu cukup). Format: `JPEG`, `PNG`, atau `WebP`. Maks 5 MB. |

Contoh `curl`:

```bash
curl -X POST "http://localhost:8094/api/reports/auto-fill" \
  -H "Authorization: Bearer <JWT>" \
  -F "file=@/path/to/foto-barang.jpg"
```

Contoh Flutter (`http` package / `dio`):

```dart
var request = http.MultipartRequest(
  'POST',
  Uri.parse('$baseUrl/api/reports/auto-fill'),
);
request.headers['Authorization'] = 'Bearer $token';
request.files.add(await http.MultipartFile.fromPath('file', fotoPath));
var response = await request.send();
```

---

## 3. Format Response Sukses (Analisis AI Berhasil)

**Status code:** `200 OK`

```json
{
  "status": "success",
  "message": "Analisis foto berhasil",
  "data": {
    "photo_url": "/uploads/9f3d2c1e-9b7a-4f2e-8b3c-5a1d6e2f8a0b.jpg",
    "title": "Dompet Kulit Hitam",
    "description": "Dompet kulit warna hitam, merek tidak terlihat, kondisi baik",
    "category": "Dompet",
    "confidence": 0.93
  }
}
```

### Deskripsi field `data`

| Field         | Tipe      | Keterangan |
|---------------|-----------|------------|
| `photo_url`   | string    | Path/URL foto yang sudah terupload. Prefix `{BASE_URL}` biar jadi URL lengkap. |
| `title`       | string    | Judul barang (bahasa Indonesia) — untuk pre-fill form. |
| `description` | string    | Deskripsi singkat (warna, kondisi, dsb.) — untuk pre-fill form. |
| `category`    | string    | Kategori barang. Diambil dari daftar tabel `categories` bila cocok. |
| `confidence`  | number    | Skor keyakinan AI, **0.0 – 1.0**. Bisa dipakai UI untuk kasih hint ke RA. |

> **Catatan confidence:** prompt Gemini memakai label `high|medium|low`. Backend
> memetakan ke number: `high → 1.0`, `medium → 0.5`, `low → 0.2`. Response API
> **selalu number**, jadi UI tidak perlu tahu soal label tsb.

---

## 4. Format Response Saat AI Gagal / Timeout (⚠️ penting)

**Status code tetap `200 OK`** — fitur ini **tidak boleh blocking** RA kalau AI down.

Foto tetap terunggah, jadi `photo_url` **selalu terisi**. Field lain kosong (string kosong / `0`), dan pesan menjelaskan kondisinya:

```json
{
  "status": "success",
  "message": "Foto berhasil diunggah, namun analisis AI gagal. Silakan isi manual.",
  "data": {
    "photo_url": "/uploads/9f3d2c1e-9b7a-4f2e-8b3c-5a1d6e2f8a0b.jpg",
    "title": "",
    "description": "",
    "category": "",
    "confidence": 0
  }
}
```

**Cara UI handle:** jika `data.title` (atau `data.category`) kosong/`""`, pertimbangkan untuk tidak otomatis mengisi form, tapi **tetap pakai `photo_url`** untuk preview gambar. Biarkan RA mengisi manual.

---

## 5. Tabel Status Code

| Status Code | Kondisi                                        | Response Bodi |
|-------------|------------------------------------------------|---------------|
| `200`       | Sukses (analisis berhasil **atau** AI gagal/timeout) | `data.photo_url` selalu terisi; field lain terisi penuh saat AI sukses, kosong saat AI gagal |
| `400`       | File tidak disertakan / format tidak didukung (`ErrInvalidFileType`) / nama file tidak aman (`ErrPathTraversal`) | `data` = `null` |
| `401`       | Token JWT tidak ada / tidak valid / kedaluwarsa | `data` = `null` |
| `413`       | File melebihi 5 MB                               | `data` = `null` |
| `500`       | Kegagalan teknis server (mis. gagal menyimpan file ke storage) | `data` = `null` |

> Catatan: AI down **bukan** error 500. Lihat bagian 4 — backend tetap `200` dengan field kosong.

Response error (non-200) selalu berbentuk:

```json
{
  "status": "error",
  "message": "<deskripsi error>",
  "data": null
}
```

---

## 6. Rekomendasi Alur di App (pre-fill form)

```text
RA foto barang
   │
   ▼
POST /reports/auto-fill (auth: Bearer token)
   │
   ▼
Response 200
   │
   ├─ data.title != ""  → pre-fill title, description, category, confidence ke form
   │
   └─ data.title == ""  → pakai data.photo_url untuk preview, form dibiarkan kosong
                              (RA isi manual)
   │
   ▼
RA review/edit → submit ke POST /reports (endpoint lama, tidak berubah)
```

- Tampilkan `confidence` (misal "90%") kalau < 0.7 sebagai saran "harap verifikasi".
- `photo_url` dari response bisa langsung dijadikan `photo_url` saat submit `POST /reports`.
- Foto sudah terupload di endpoint ini — app **tidak perlu** panggil `POST /upload` lagi secara terpisah.

---

## 7. Variabel Environment (untuk backend, bukan UI)

Dipakai di `.env` server, bukan di app:

```env
GEMINI_API_KEY=<api key dari Google AI Studio>
GEMINI_MODEL=gemini-3.6-flash          # optional, ada fallback default
GEMINI_AUTOFILL_PROMPT=                # optional, override prompt ({{categories}} = daftar kategori)
GEMINI_JSON_MODE=true                  # optional; set "false" untuk model tanpa structured output
GEMINI_TIMEOUT_SECONDS=60              # optional; timeout request ke Gemini (default 60)
GEMINI_THINKING_LEVEL=minimal          # optional; minimal/low/medium/high, atau "off"
```