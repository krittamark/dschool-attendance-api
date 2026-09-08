import urllib.request
import urllib.parse
import json
import time

BASE_URL = "http://localhost:8080"

test_cases = [
    {
        "id": "TC-01",
        "name": "Health Check & Service Status",
        "url": f"{BASE_URL}/api/health",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d.get("data", {}).get("status") == "healthy"
    },
    {
        "id": "TC-02",
        "name": "Daily Classroom Attendance (Senior High M.401)",
        "url": f"{BASE_URL}/api/attendance/daily?classroom=401",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d["data"]["classroom"] == "ม.401" and len(d["data"]["students"]) == 30
    },
    {
        "id": "TC-03",
        "name": "Daily Classroom Attendance (Junior High M.101)",
        "url": f"{BASE_URL}/api/attendance/daily?classroom=101",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d["data"]["classroom"] == "ม.101" and len(d["data"]["students"]) == 30
    },
    {
        "id": "TC-04",
        "name": "Historical Daily Attendance with Date Filter (2026-09-07)",
        "url": f"{BASE_URL}/api/attendance/daily?classroom=401&date=2026-09-07",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d["data"]["date"] == "7 ก.ย. 2569" and any(s["status"] == "สาย" for s in d["data"]["students"])
    },
    {
        "id": "TC-05",
        "name": "School Overview Attendance & Grade Breakdown",
        "url": f"{BASE_URL}/api/attendance/overview",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d["data"]["summary"]["total"] == 3534 and len(d["data"]["grade_levels"]) == 6
    },
    {
        "id": "TC-06",
        "name": "Monthly Classroom Attendance (M.401 Month 9 Year 2569)",
        "url": f"{BASE_URL}/api/attendance/monthly?classroom=401&month=9&year=2569",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and len(d["data"]["students"]) > 0 and d["data"]["students"][0]["present"] >= 0
    },
    {
        "id": "TC-07",
        "name": "Semester Classroom Attendance (M.401 Term 1)",
        "url": f"{BASE_URL}/api/attendance/semester?classroom=401&term=1",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d["data"]["term"] == 1 and len(d["data"]["students"]) == 30
    },
    {
        "id": "TC-08",
        "name": "Student Search by Name (กิตติพิชญ์)",
        "url": f"{BASE_URL}/api/students/search?" + urllib.parse.urlencode({"q": "กิตติพิชญ์"}),
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and len(d["data"]) == 2 and d["data"][1]["sd_no"] == "18948"
    },
    {
        "id": "TC-09",
        "name": "Individual Student Attendance Profile (sd_no 18948 Term 1)",
        "url": f"{BASE_URL}/api/students/18948/attendance?term=1",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d["data"]["sd_no"] == "18948" and d["data"]["summary"]["present"] == 57 and len(d["data"]["history"]) > 50
    },
    {
        "id": "TC-10",
        "name": "Morning Flag Ceremony Assembly Attendance",
        "url": f"{BASE_URL}/api/attendance/flag-ceremony",
        "expected_status": 200,
        "assert_fn": lambda d: d.get("success") is True and d["data"]["summary"]["total"] == 3534 and d["data"]["summary"]["attended"] == 462
    },
    {
        "id": "TC-11",
        "name": "Validation Error Handling (Search without query parameter)",
        "url": f"{BASE_URL}/api/students/search",
        "expected_status": 400,
        "assert_fn": lambda d: d.get("success") is False and "required" in d.get("error", "")
    },
    {
        "id": "TC-12",
        "name": "Cross-Origin Resource Sharing (CORS Headers)",
        "url": f"{BASE_URL}/api/health",
        "expected_status": 200,
        "assert_fn": lambda d: True
    }
]

results = []
print(f"{'ID':<7} | {'Status':<6} | {'Time (ms)':<9} | {'HTTP':<4} | Test Case")
print("-" * 75)

for tc in test_cases:
    start = time.perf_counter()
    status_code = None
    passed = False
    details = ""
    try:
        req = urllib.request.Request(tc["url"], headers={"Origin": "http://localhost:3000", "X-API-Key": "dschool-secret-key-2026"})
        with urllib.request.urlopen(req) as resp:
            elapsed_ms = (time.perf_counter() - start) * 1000
            status_code = resp.status
            body = resp.read().decode("utf-8")
            data = json.loads(body)
            
            if tc["id"] == "TC-12":
                cors_origin = resp.headers.get("Access-Control-Allow-Origin")
                passed = (cors_origin == "*")
            else:
                passed = (status_code == tc["expected_status"] and tc["assert_fn"](data))
    except urllib.error.HTTPError as e:
        elapsed_ms = (time.perf_counter() - start) * 1000
        status_code = e.code
        body = e.read().decode("utf-8")
        try:
            data = json.loads(body)
            passed = (status_code == tc["expected_status"] and tc["assert_fn"](data))
        except Exception:
            passed = False
    except Exception as e:
        elapsed_ms = (time.perf_counter() - start) * 1000
        passed = False
        details = str(e)

    res_str = "PASS" if passed else "FAIL"
    print(f"{tc['id']:<7} | {res_str:<6} | {elapsed_ms:>8.1f}  | {status_code:<4} | {tc['name']}")
    results.append({
        "id": tc["id"],
        "name": tc["name"],
        "passed": passed,
        "elapsed_ms": elapsed_ms,
        "status_code": status_code,
        "url": tc["url"]
    })

passed_count = sum(1 for r in results if r["passed"])
print("-" * 75)
print(f"Total: {len(results)} | Passed: {passed_count} | Failed: {len(results) - passed_count}")

# Dump detailed json for reporting
with open("test_results.json", "w", encoding="utf-8") as f:
    json.dump(results, f, ensure_ascii=False, indent=2)
