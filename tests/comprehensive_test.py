import urllib.request
import urllib.parse
import json
import time
import concurrent.futures
from bs4 import BeautifulSoup
import http.cookiejar

BASE_URL = "http://localhost:8080"
DSCHOOL_LOGIN_URL = "http://dschool-l1.gp-education.com/dschool_app_v2020/index.php?app=m&mobile_id=e4K2ixcwaVI:APA91bFqNdfS7jbA_WeiD9Zfo3YMjfgEtz_ELKerBjZAyTIFdztCFhAMM0U2oI4IneqAjuUYMexPFZPDn3nc0ZXjQs-tt5PxOPBGAMn9v1wZcWSd0T6VtKRL1K2cD8f_5HlXIFaA1qfP&gcm_regid=e4K2ixcwaVI:APA91bFqNdfS7jbA_WeiD9Zfo3YMjfgEtz_ELKerBjZAyTIFdztCFhAMM0U2oI4IneqAjuUYMexPFZPDn3nc0ZXjQs-tt5PxOPBGAMn9v1wZcWSd0T6VtKRL1K2cD8f_5HlXIFaA1qfP&type=m&school_id=1013270183&change_stat=1&latitude=0&longitude=0&v2022=8%3Bp8%3Bp8%3Bp33195susv%2C&user_id="

print("================================================================================")
print("     COMPREHENSIVE DEEP VERIFICATION & INTEGRITY TEST SUITE (dschool API)       ")
print("================================================================================")

test_stats = {"total": 0, "passed": 0, "failed": 0}
API_KEY = "dschool-secret-key-2026"

def api_open(url):
    req = urllib.request.Request(url, headers={"X-API-Key": API_KEY})
    return urllib.request.urlopen(req)

def record_test(name, passed, details=""):
    test_stats["total"] += 1
    if passed:
        test_stats["passed"] += 1
        print(f"[PASS] {name}")
    else:
        test_stats["failed"] += 1
        print(f"[FAIL] {name} -> {details}")

# ==============================================================================
# SECTION 1: DIRECT RAW WEB vs API ACCURACY (เทียบความตรงกัน 100% กับหน้าเว็บ dschool)
# ==============================================================================
print("\n--- SECTION 1: Direct Web vs API Exact Field-by-Field Match ---")

try:
    # 1. Login directly to dschool to get raw HTML for M.401 on today (2026-09-08)
    cj = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cj))
    opener.addheaders = [("User-Agent", "Mozilla/5.0")]
    opener.open(DSCHOOL_LOGIN_URL)
    
    # Set session for 401
    set_url = "https://dschool-l1.gp-education.com/dschool_app_v2020/set_sesstion.php?url=menu_105.php?menu=system[and]submenu_id=105&sname=edlevel&sdata=3&sname2=class_input&sdata2=401"
    res_web = opener.open(set_url)
    web_html = res_web.read().decode("utf-8", errors="ignore")
    
    # Parse raw web HTML
    soup = BeautifulSoup(web_html, "html.parser")
    web_students = []
    for tbl in soup.find_all("table"):
        if "ประตู" in tbl.text and "ชื่อ" in tbl.text:
            for rIdx, tr in enumerate(tbl.find_all("tr")):
                if rIdx == 0: continue
                tds = [c.get_text(strip=True) for c in tr.find_all("td")]
                if len(tds) >= 7 and tds[0].isdigit():
                    web_students.append({
                        "no": int(tds[0]),
                        "name": tds[2],
                        "gate": tds[3],
                        "status": tds[4],
                        "check_in_time": tds[5],
                        "check_out_time": tds[6]
                    })
    
    # 2. Call our Go API
    with api_open(f"{BASE_URL}/api/attendance/daily?classroom=401&date=2026-09-08") as resp:
        api_data = json.loads(resp.read().decode("utf-8"))["data"]
        api_students = api_data["students"]
        
    # Compare counts
    record_test("Raw Web vs API Student Count Match", len(web_students) == len(api_students), f"Web: {len(web_students)}, API: {len(api_students)}")
    
    # Compare each student
    all_match = True
    mismatches = []
    for ws, as_ in zip(web_students, api_students):
        if ws["no"] != as_["no"] or ws["name"] != as_["name"] or ws["status"] != as_["status"] or ws["gate"] != as_["gate"] or ws["check_in_time"] != as_["check_in_time"]:
            all_match = False
            mismatches.append(f"No {ws['no']}: Web={ws} vs API={as_}")
            
    record_test("Raw Web vs API Student Records Exact Match (Name, Gate, Status, Time)", all_match, f"Mismatches: {mismatches[:2]}")

except Exception as e:
    record_test("Direct Web vs API Check", False, str(e))


# ==============================================================================
# SECTION 2: MATHEMATICAL CONSISTENCY CHECKS (ตรวจสอบความสอดคล้องของตัวเลข)
# ==============================================================================
print("\n--- SECTION 2: Mathematical Sum & Breakdown Consistency ---")

try:
    with api_open(f"{BASE_URL}/api/attendance/overview?date=2026-09-08") as resp:
        ov = json.loads(resp.read().decode("utf-8"))["data"]
        
    # 1. Sum of all grades total == summary total
    sum_grades_total = sum(g["total"] for g in ov["grade_levels"])
    record_test("Overview: Sum of grade levels total == Summary total (3,534)", sum_grades_total == ov["summary"]["total"], f"Sum: {sum_grades_total}, Summary: {ov['summary']['total']}")
    
    # 2. Sum of all grades present == summary present
    sum_grades_present = sum(g["present"] for g in ov["grade_levels"])
    record_test("Overview: Sum of grade levels present == Summary present", sum_grades_present == ov["summary"]["present"], f"Sum: {sum_grades_present}, Summary: {ov['summary']['present']}")
    
    # 3. Sum of all grades unrecorded == summary unrecorded
    sum_grades_unrec = sum(g["unrecorded"] for g in ov["grade_levels"])
    record_test("Overview: Sum of grade levels unrecorded == Summary unrecorded", sum_grades_unrec == ov["summary"]["unrecorded"], f"Sum: {sum_grades_unrec}, Summary: {ov['summary']['unrecorded']}")

    # 4. For each grade level: present + off_campus + unrecorded + sick + personal + absent == total
    math_valid_grades = True
    grade_errors = []
    for g in ov["grade_levels"]:
        sub_sum = g["present"] + g["off_campus"] + g["unrecorded"] + g["sick_leave"] + g["personal_leave"] + g["absent"]
        if sub_sum != g["total"]:
            math_valid_grades = False
            grade_errors.append(f"{g['grade']}: parts sum {sub_sum} != total {g['total']}")
    record_test("Overview: Each grade level sub-metrics sum == grade total", math_valid_grades, str(grade_errors))

    # 5. Classrooms under each grade sum up to grade total
    class_sum_valid = True
    for g in ov["grade_levels"]:
        c_list = ov["classrooms"].get(g["grade"], [])
        if c_list:
            c_sum = sum(c["total"] for c in c_list)
            if c_sum != g["total"]:
                class_sum_valid = False
                break
    record_test("Overview: Classrooms breakdown sum == Grade level total for all 6 grades", class_sum_valid)

except Exception as e:
    record_test("Mathematical Consistency in Overview", False, str(e))


# ==============================================================================
# SECTION 3: CROSS-ENDPOINT INTEGRITY (Daily vs Student 360 History Consistency)
# ==============================================================================
print("\n--- SECTION 3: Cross-Endpoint Student Integrity (Daily vs Student 360) ---")

try:
    # Check student #1 in 401 ("กิตติพิชญ์ ปัชชามูล", sd_no: 18948)
    dates_to_verify = [
        ("2026-09-08", "08 ก.ย. 69", "มา"),
        ("2026-09-07", "07 ก.ย. 69", "สาย"),
        ("2026-09-06", "06 ก.ย. 69", "ไม่ลงเวลา"),
        ("2026-09-05", "05 ก.ย. 69", "ไม่ลงเวลา"),
        ("2026-09-04", "04 ก.ย. 69", "มา"),
    ]
    
    # 1. Fetch Student 360 profile history
    with api_open(f"{BASE_URL}/api/students/18948/attendance?term=1") as resp:
        profile = json.loads(resp.read().decode("utf-8"))["data"]
        history_map = {item["date"]: item["status"] for item in profile["history"]}
        
    cross_check_passed = True
    for date_param, thai_date, expected_status in dates_to_verify:
        # Check daily API
        with api_open(f"{BASE_URL}/api/attendance/daily?classroom=401&date={date_param}") as resp:
            daily_data = json.loads(resp.read().decode("utf-8"))["data"]
            student_daily = next((s for s in daily_data["students"] if s["name"] == "กิตติพิชญ์ ปัชชามูล"), None)
            
            daily_status = student_daily["status"] if student_daily else "NOT_FOUND"
            history_status = history_map.get(thai_date, "NOT_FOUND")
            
            is_match = (daily_status == expected_status == history_status)
            if not is_match:
                cross_check_passed = False
            record_test(f"Cross-Check Date {thai_date}: Daily={daily_status}, Profile={history_status}, Exp={expected_status}", is_match)
            
except Exception as e:
    record_test("Cross-Endpoint Student Integrity Check", False, str(e))


# ==============================================================================
# SECTION 4: CLASSROOM & GRADE LEVEL BROAD COVERAGE (ทดสอบหลากหลายระดับชั้น)
# ==============================================================================
print("\n--- SECTION 4: Broad Classroom & Grade Coverage ---")

sample_classes = [
    # M.1
    ("101", "ม.101", 30),
    ("105", "ม.105", 36),
    ("115", "ม.115", 38),
    # M.2
    ("201", "ม.201", 30),
    ("210", "ม.210", 41),
    # M.3
    ("301", "ม.301", 30),
    ("315", "ม.315", 38),
    # M.4
    ("401", "ม.401", 30),
    ("408", "ม.408", 41),
    ("417", "ม.417", 40),
    # M.5
    ("501", "ม.501", 30),
    ("517", "ม.517", 40),
    # M.6
    ("601", "ม.601", 30),
    ("617", "ม.617", 39),
]

for class_code, expected_label, expected_min_count in sample_classes:
    try:
        url = f"{BASE_URL}/api/attendance/daily?classroom={class_code}&date=2026-09-08"
        with api_open(url) as resp:
            data = json.loads(resp.read().decode("utf-8"))["data"]
            students_count = len(data["students"])
            total_stat = data["summary"][0]["total"] if data["summary"] else 0
            
            passed = (data["classroom"] == expected_label and students_count == total_stat and students_count >= expected_min_count - 5)
            record_test(f"Classroom {expected_label}: students={students_count}, summary_total={total_stat}", passed)
    except Exception as e:
        record_test(f"Classroom {class_code}", False, str(e))


# ==============================================================================
# SECTION 5: DATE FORMAT ROBUSTNESS (ทดสอบรูปแบบวันที่หลากหลาย)
# ==============================================================================
print("\n--- SECTION 5: Date Format & Temporal Variations ---")

date_variations = [
    ("YYYY-MM-DD Standard", "2026-09-08", "8 ก.ย. 2569"),
    ("DD/MM/YYYY Format", "08/09/2026", "8 ก.ย. 2569"),
    ("Buddhist YY/MM/DD", "69/09/08", "8 ก.ย. 2569"),
    ("Buddhist YYYY/MM/DD", "2569/09/08", "8 ก.ย. 2569"),
    ("Keyword 'today'", "today", "8 ก.ย. 2569"),
    ("Empty Date (Default to today)", "", "8 ก.ย. 2569"),
    ("Sunday (Weekend with Unrecorded status)", "2026-09-06", "6 ก.ย. 2569"),
    ("Saturday (Weekend with Unrecorded status)", "2026-09-05", "5 ก.ย. 2569"),
    ("Earlier School Day (2026-09-01)", "2026-09-01", "1 ก.ย. 2569"),
    ("Previous Month (2026-08-28)", "2026-08-28", "28 ส.ค. 2569"),
]

for name, date_val, expected_thai_date in date_variations:
    try:
        url = f"{BASE_URL}/api/attendance/daily?classroom=401"
        if date_val:
            url += f"&date={date_val}"
        with api_open(url) as resp:
            data = json.loads(resp.read().decode("utf-8"))["data"]
            passed = (data["date"] == expected_thai_date and len(data["students"]) == 30)
            record_test(f"Date {name} ({date_val or 'empty'}) -> {data['date']}", passed)
    except Exception as e:
        record_test(f"Date {name}", False, str(e))


# ==============================================================================
# SECTION 6: SEARCH & PROFILE ACCURACY (ทดสอบค้นหาชื่อหลากหลายรูปแบบ)
# ==============================================================================
print("\n--- SECTION 6: Search & Profile Variations ---")

search_cases = [
    ("First Name Only (กิตติพิชญ์)", "กิตติพิชญ์", 2),
    ("Last Name Only (ปัชชามูล)", "ปัชชามูล", 1),
    ("Common Name (ธนกร)", "ธนกร", 1),
    ("Common Thai Prefix (ณัฐ)", "ณัฐ", 1),
    ("Single Char (ก)", "ก", 1),
]

for name, query, min_expected in search_cases:
    try:
        url = f"{BASE_URL}/api/students/search?" + urllib.parse.urlencode({"q": query})
        with api_open(url) as resp:
            data = json.loads(resp.read().decode("utf-8"))["data"]
            passed = (len(data) >= min_expected and all(d["sd_no"].isdigit() for d in data))
            record_test(f"Search {name} -> Found {len(data)} students", passed)
    except Exception as e:
        record_test(f"Search {name}", False, str(e))


# ==============================================================================
# SECTION 7: CONCURRENCY & SESSION ISOLATION (ทดสอบยิงพร้อมกันหลาย Thread)
# ==============================================================================
print("\n--- SECTION 7: Concurrency & Mutex Session Isolation ---")

def worker_fetch_class(classroom):
    url = f"{BASE_URL}/api/attendance/daily?classroom={classroom}&date=2026-09-08"
    req = urllib.request.Request(url, headers={"X-API-Key": "dschool-secret-key-2026"})
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read().decode("utf-8"))["data"]

concurrent_classes = ["101", "201", "301", "401", "501", "601", "102", "402"]
results_map = {}

with concurrent.futures.ThreadPoolExecutor(max_workers=8) as executor:
    future_to_class = {executor.submit(worker_fetch_class, c): c for c in concurrent_classes}
    for future in concurrent.futures.as_completed(future_to_class):
        c = future_to_class[future]
        try:
            data = future.result()
            results_map[c] = data
        except Exception as exc:
            results_map[c] = str(exc)

concurrency_passed = True
for c in concurrent_classes:
    res = results_map.get(c)
    if isinstance(res, dict) and res.get("classroom") == f"ม.{c}":
        pass
    else:
        concurrency_passed = False
        print(f"   Mismatch for class {c}: got {res}")

record_test(f"Concurrent 8 simultaneous requests: All returned their exact distinct classroom without session bleed", concurrency_passed)


# ==============================================================================
# SECTION 8: RIGOROUS NEGATIVE & EXTREME EDGE CASES (ทดสอบค่าแปลกๆ ค่าหลุดขอบ)
# ==============================================================================
print("\n--- SECTION 8: Comprehensive Negative & Extreme Edge Cases ---")

negative_cases = [
    # Classroom boundaries
    ("Invalid Classroom Prefix (701)", f"{BASE_URL}/api/attendance/daily?classroom=701", 400),
    ("Invalid Classroom (000)", f"{BASE_URL}/api/attendance/daily?classroom=000", 400),
    ("Invalid Classroom (999)", f"{BASE_URL}/api/attendance/daily?classroom=999", 400),
    ("Classroom Symbols (&, $, @)", f"{BASE_URL}/api/attendance/daily?classroom=401%26%24%40", 400),
    ("Classroom Script Injection", f"{BASE_URL}/api/attendance/daily?classroom=%3Cscript%3E", 400),
    ("Classroom Path Traversal (../../)", f"{BASE_URL}/api/attendance/daily?classroom=..%2F..%2F", 400),
    
    # Date boundaries
    ("Impossible Date Feb 30", f"{BASE_URL}/api/attendance/daily?classroom=401&date=2026-02-30", 400),
    ("Impossible Month 13", f"{BASE_URL}/api/attendance/daily?classroom=401&date=2026-13-01", 400),
    ("Impossible Day 32", f"{BASE_URL}/api/attendance/daily?classroom=401&date=2026-01-32", 400),
    ("Date Text 'yesterday'", f"{BASE_URL}/api/attendance/daily?classroom=401&date=yesterday", 400),
    ("Date SQLi (' OR '1'='1)", f"{BASE_URL}/api/attendance/daily?classroom=401&date=2026-09-08%27%20OR%20%271%27%3D%271", 400),
    
    # Student ID boundaries
    ("Student ID Letters", f"{BASE_URL}/api/students/abc123/attendance", 400),
    ("Student ID Negative (-1)", f"{BASE_URL}/api/students/-1/attendance", 400),
    ("Student ID Path Traversal", f"{BASE_URL}/api/students/..%2F..%2Fattendance", 404),
    ("Student ID Injection", f"{BASE_URL}/api/students/%27%20OR%201%3D1/attendance", 400),
    
    # Month & Year & Term boundaries
    ("Month 0", f"{BASE_URL}/api/attendance/monthly?classroom=401&month=0&year=2569", 400),
    ("Month 13", f"{BASE_URL}/api/attendance/monthly?classroom=401&month=13&year=2569", 400),
    ("Month text 'september'", f"{BASE_URL}/api/attendance/monthly?classroom=401&month=september&year=2569", 400),
    ("Year Out of Range (1990)", f"{BASE_URL}/api/attendance/monthly?classroom=401&month=9&year=1990", 400),
    ("Term 0", f"{BASE_URL}/api/attendance/semester?classroom=401&term=0", 400),
    ("Term 3", f"{BASE_URL}/api/attendance/semester?classroom=401&term=3", 400),
    ("Term Text 'first'", f"{BASE_URL}/api/attendance/semester?classroom=401&term=first", 400),
    
    # Search boundaries
    ("Search Query Empty (q=)", f"{BASE_URL}/api/students/search?q=", 400),
    ("Search Query 105 chars", f"{BASE_URL}/api/students/search?q=" + ("x"*105), 400),
    ("Search Non-existent (Random GUID)", f"{BASE_URL}/api/students/search?q=a8b7c6d5-e4f3-2109-8765-43210fedcba9", 200),
]

for name, url, expected_status in negative_cases:
    try:
        req = urllib.request.Request(url, headers={"X-API-Key": "dschool-secret-key-2026"})
        with urllib.request.urlopen(req) as resp:
            code = resp.status
            data = json.loads(resp.read().decode("utf-8"))
            passed = (code == expected_status)
            record_test(f"Negative: {name} (HTTP {code}, Exp {expected_status})", passed)
    except urllib.error.HTTPError as e:
        passed = (e.code == expected_status)
        record_test(f"Negative: {name} (HTTP {e.code}, Exp {expected_status})", passed)
    except Exception as e:
        record_test(f"Negative: {name}", False, str(e))


# ==============================================================================
# SUMMARY REPORT
# ==============================================================================
print("\n================================================================================")
print(f"TEST EXECUTION FINISHED: Total {test_stats['total']} tests | Passed: {test_stats['passed']} | Failed: {test_stats['failed']}")
print(f"Overall Pass Rate: {(test_stats['passed'] / test_stats['total']) * 100:.1f}%")
print("================================================================================")
