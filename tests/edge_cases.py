import urllib.request
import urllib.parse
import json

BASE = "http://localhost:8080"

edge_tests = [
    # 1. Invalid Classrooms
    ("Non-existent Classroom (9999)", f"{BASE}/api/attendance/daily?classroom=9999", 400),
    ("Letters as Classroom (abc)", f"{BASE}/api/attendance/daily?classroom=abc", 400),
    ("Negative Classroom (-1)", f"{BASE}/api/attendance/daily?classroom=-1", 400),
    ("SQL Injection in Classroom", f"{BASE}/api/attendance/daily?classroom=401%27%20OR%201=1", 400),
    
    # 2. Invalid Dates
    ("Garbage Date format", f"{BASE}/api/attendance/daily?classroom=401&date=invalid-date", 400),
    ("Future Valid Date format (2035-12-31)", f"{BASE}/api/attendance/daily?classroom=401&date=2035-12-31", 200),
    ("SQL Injection in Date", f"{BASE}/api/attendance/daily?classroom=401&date=2026-09-08%27--", 400),
    
    # 3. Student Search Edge Cases
    ("Search Non-existent student", f"{BASE}/api/students/search?q=XYZNonExistent99999", 200),
    ("Search Empty string (q=)", f"{BASE}/api/students/search?q=", 400),
    ("Search Exceeding max length (101 chars)", f"{BASE}/api/students/search?q=" + ("a"*105), 400),
    
    # 4. Student Profile Edge Cases
    ("Letters as sd_no", f"{BASE}/api/students/invalid_id/attendance?term=1", 400),
    ("SQL Injection in sd_no", f"{BASE}/api/students/%27%20OR%201=1/attendance?term=1", 400),
    ("Non-existent Numeric sd_no (99999999)", f"{BASE}/api/students/99999999/attendance?term=1", 200),
    
    # 5. Month, Year, Term Edge Cases
    ("Invalid Month 13", f"{BASE}/api/attendance/monthly?classroom=401&month=13&year=2569", 400),
    ("Invalid Month -1", f"{BASE}/api/attendance/monthly?classroom=401&month=-1&year=2569", 400),
    ("Invalid Term 5", f"{BASE}/api/attendance/semester?classroom=401&term=5", 400),
    ("Invalid Term abc", f"{BASE}/api/attendance/semester?classroom=401&term=abc", 400),
]

print(f"{'Status':<6} | {'HTTP':<8} | Edge Case")
print("-" * 65)

passed = 0
for name, url, expected_code in edge_tests:
    try:
        req = urllib.request.Request(url, headers={"X-API-Key": "dschool-secret-key-2026"})
        with urllib.request.urlopen(req) as resp:
            status = "PASS" if resp.status == expected_code else "FAIL"
            if status == "PASS": passed += 1
            print(f"{status:<6} | {resp.status} (exp {expected_code}) | {name}")
    except urllib.error.HTTPError as e:
        status = "PASS" if e.code == expected_code else "FAIL"
        if status == "PASS": passed += 1
        print(f"{status:<6} | {e.code} (exp {expected_code}) | {name}")
    except Exception as e:
        print(f"FAIL   | ERR      | {name}: {e}")

print("-" * 65)
print(f"Total: {len(edge_tests)} | Passed: {passed} | Failed: {len(edge_tests) - passed}")
