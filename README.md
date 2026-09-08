# dschool Attendance Web API (Golang)

RESTful Web API พัฒนาด้วยภาษา Go สำหรับดึงและแปลงข้อมูลสถิติการ **ขาด ลา มาสาย เข้าเรียน** ของนักเรียนจากระบบ dschool (PHP Web Application) ให้กลายเป็น JSON Web API ที่มีโครงสร้างข้อมูลชัดเจน ใช้งานง่าย เชื่อมต่อกับ Frontend และมีระบบรักษาความปลอดภัยด้วย **API Key Authentication**

---

## สถาปัตยกรรมและฟีเจอร์หลัก

- **API Key Authentication**: ป้องกันทุก Endpoint ข้อมูลด้วย API Key รองรับ 3 รูปแบบ (Header `X-API-Key`, `Authorization: Bearer <token>`, หรือ Query `?api_key=`)
- **Authentication & Session Bridge**: เชื่อมต่อระบบ dschool อัตโนมัติด้วย `PHPSESSID` ใน `http.CookieJar` พร้อมระบบ Auto-relogin เมื่อเซสชันหมดอายุ
- **Session State Management**: จัดการสลับค่าตัวแปรใน `set_sesstion.php` (วันที่ `date0`, ห้องเรียน `class_input` & `edlevel`, เดือน `month0`, ภาคเรียน `term0`) โดยใช้ Mutex Lock ป้องกัน Race Condition
- **HTML DOM Parsing**: สกัดตารางข้อมูลด้วย `goquery` และแปลงค่าตัวเลข สถิติภาษาไทยให้เป็น Data Types ที่พร้อมใช้งาน
- **OpenAPI 3.0 & Swagger UI**: เอกสารมาตรฐานสากล พร้อม Interactive Swagger UI ในตัวที่ `/docs`
- **Standalone UI Documentation**: หน้าเว็บ UI API Document & Explorer แยกอิสระในโฟลเดอร์ `docs-ui/`

---

## การยืนยันตัวตน (API Key Authentication)

ทุกคำขอที่เรียกไปยัง Endpoint ข้อมูล (ยกเว้น `/api/health`, `/openapi.json`, `/docs`) จะต้องแนบ API Key อย่างใดอย่างหนึ่งใน 3 วิธีดังต่อไปนี้:

1. **ผ่าน Header `X-API-Key` (แนะนำ)**:
   ```bash
   curl -H "X-API-Key: dschool-secret-key-2026" "http://localhost:8080/api/attendance/daily?classroom=401"
   ```
2. **ผ่าน Header `Authorization: Bearer`**:
   ```bash
   curl -H "Authorization: Bearer dschool-secret-key-2026" "http://localhost:8080/api/attendance/daily?classroom=401"
   ```
3. **ผ่าน Query Parameter `?api_key=`**:
   ```bash
   curl "http://localhost:8080/api/attendance/daily?classroom=401&api_key=dschool-secret-key-2026"
   ```

> ค่าเริ่มต้นสำหรับ Development คือ `dschool-secret-key-2026`  
> สามารถเปลี่ยนค่าได้ง่ายๆ ผ่าน Environment Variable: `API_KEY="your-custom-production-key"`

หากไม่ได้ส่ง API Key หรือส่งค่าไม่ถูกต้อง เซิร์ฟเวอร์จะปฏิเสธด้วยรหัส **`401 Unauthorized`**:
```json
{
  "success": false,
  "error": "Unauthorized: Invalid or missing API Key. Please provide via 'X-API-Key' header, 'Authorization: Bearer <key>', or '?api_key=<key>'"
}
```

---

## รายการ API Endpoints

| Method | Endpoint | สิทธิ์เข้าถึง | คำอธิบาย |
|---|---|:---:|---|
| `GET` | `/api/health` | Public | ตรวจสอบสถานะการทำงานของเซิร์ฟเวอร์ |
| `GET` | `/docs` | Public | Interactive Swagger UI บนเบราว์เซอร์ |
| `GET` | `/openapi.json` | Public | OpenAPI 3.0 Specification (JSON) |
| `GET` | `/api/attendance/daily` | **Protected** | ข้อมูลการลงเวลาประจำวันของห้องเรียน (รายชื่อ, เวลา, ประตู, สถานะ) |
| `GET` | `/api/attendance/overview` | **Protected** | สรุปภาพรวมสถิติทั่วทั้งโรงเรียน และรายระดับชั้น ม.1 - ม.6 |
| `GET` | `/api/attendance/monthly` | **Protected** | สรุปสถิติประจำเดือนรายห้อง (มา, สาย, ไม่ลงเวลา, ลา, ขาด) |
| `GET` | `/api/attendance/semester` | **Protected** | สรุปสถิติสะสมประจำภาคเรียนรายห้อง (เทอม 1 หรือ 2) |
| `GET` | `/api/students/search` | **Protected** | ค้นหานักเรียนตามชื่อ นามสกุล หรือรหัสนักเรียน |
| `GET` | `/api/students/{sd_no}/attendance` | **Protected** | ประวัติการเข้าเรียนรายบุคคลย้อนหลังทั้งหมด |
| `GET` | `/api/attendance/flag-ceremony` | **Protected** | สถิติการเข้าร่วมกิจกรรมหน้าเสาธง |
| `GET` | `/api/admin/keys` | **Admin Only** | ดูรายการ API Key ทั้งหมด สถานะ และสิทธิ์ |
| `POST` | `/api/admin/keys` | **Admin Only** | สร้าง API Key ใหม่สำหรับไคลเอนต์หรือแอดมิน |
| `POST` | `/api/admin/keys/revoke` | **Admin Only** | เพิกถอน / ระงับ API Key ทันที (Revoke) |
| `POST` | `/api/admin/keys/activate` | **Admin Only** | เปิดใช้งาน API Key ที่เคยถูกเพิกถอนอีกครั้ง |

---

## ระบบจัดการ API Key สำหรับ Admin (Admin Key Management & Revocation)

ระบบรองรับการบริหารจัดการ API Key แบบไดนามิกผ่าน Admin REST API และ UI Manager โดยข้อมูลถูกจัดเก็บแบบ Thread-safe ลงในไฟล์ `data/api_keys.json`

### สิทธิ์การเข้าถึงระดับ Admin (Role-Based Authorization)
- การเรียก Endpoint `/api/admin/*` จะต้องใช้ API Key ที่มีสิทธิ์ระดับ `admin` หรือ **Master Admin Key** (`dschool-admin-master-key-2026`)
- หาก Client Key ทั่วไปพยายามเรียก Admin API จะถูกปฏิเสธด้วย **`403 Forbidden`**
- **Master Admin Key** ได้รับการคุ้มครองพิเศษ ไม่สามารถถูกเพิกถอน (Revoke) ได้ เพื่อป้องกันสภาวะ Admin Lockout

---

### ตัวอย่างคำสั่ง Admin API

#### 1. ดูรายการ API Key ทั้งหมด (List Keys)
```bash
curl -X GET "http://localhost:8080/api/admin/keys" \
  -H "X-API-Key: dschool-admin-master-key-2026"
```
*ตัวอย่างผลลัพธ์:*
```json
{
  "success": true,
  "message": "API keys retrieved successfully",
  "data": [
    {
      "key": "dschool-secret-key-2026",
      "name": "Default Client Key",
      "is_active": true,
      "role": "client",
      "created_at": "2026-09-08T19:55:29+07:00"
    },
    {
      "key": "dsk_live_04abd0f4792ed29198ab5f23fa3a9428",
      "name": "LINE Official Bot Integration",
      "is_active": true,
      "role": "client",
      "created_at": "2026-09-08T19:55:39+07:00",
      "last_used_at": "2026-09-08T19:55:41+07:00"
    }
  ]
}
```

#### 2. สร้าง API Key ใหม่ (Create Key)
```bash
curl -X POST "http://localhost:8080/api/admin/keys" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dschool-admin-master-key-2026" \
  -d '{
    "name": "Mobile App Production",
    "role": "client"
  }'
```
*ผลลัพธ์จะส่งคืน API Key แบบสุ่มความปลอดภัยสูง เช่น `dsk_live_3f8a...`*

#### 3. ระงับ / เพิกถอน API Key (Revoke Key)
```bash
curl -X POST "http://localhost:8080/api/admin/keys/revoke" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dschool-admin-master-key-2026" \
  -d '{
    "key": "dsk_live_3f8a..."
  }'
```
> **ผลทันที**: คำขอหลังจากนี้ที่ใช้ Key ดังกล่าวจะถูกปฏิเสธด้วย `401 Unauthorized` ทันที

#### 4. เปิดใช้งาน API Key คืน (Activate Key)
```bash
curl -X POST "http://localhost:8080/api/admin/keys/activate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dschool-admin-master-key-2026" \
  -d '{
    "key": "dsk_live_3f8a..."
  }'
```

---

## วิธีติดตั้งและรันระบบ

### 1. รันเซิร์ฟเวอร์
```bash
# รันด้วยค่าเริ่มต้น
go run cmd/server/main.go

# กำหนดพอร์ต และ API Key ที่กำหนดเอง
PORT=8080 API_KEY="my-secret-token" go run cmd/server/main.go
```

### 2. คอมไพล์เป็น Binary Executable
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

### 3. รันชุดทดสอบ (Test Suites)
```bash
# Unit Tests
go test -v ./...

# Automated Integration & Auth Tests
python3 tests/auth_tests.py
python3 tests/test_suite.py
python3 tests/comprehensive_test.py
```

---

## หน้าเว็บ Standalone UI Documentation

สามารถเปิดดูและทดสอบ Interactive Explorer ได้ทันทีโดยไม่ต้องผ่านเซิร์ฟเวอร์หลัก:
```bash
open docs-ui/index.html
```
หรือรันผ่าน Local Server:
```bash
python3 -m http.server 3000 --directory docs-ui
```
