import urllib.request
import urllib.parse
import json

BASE = "http://localhost:8080"
VALID_KEY = "dschool-secret-key-2026"

auth_cases = [
    ("Health Check without Key (Public)", f"{BASE}/api/health", {}, 200),
    ("Swagger UI without Key (Public)", f"{BASE}/docs", {}, 200),
    ("OpenAPI JSON without Key (Public)", f"{BASE}/openapi.json", {}, 200),
    
    ("Missing API Key on Protected Endpoint", f"{BASE}/api/attendance/daily?classroom=401", {}, 401),
    ("Invalid API Key Header", f"{BASE}/api/attendance/daily?classroom=401", {"X-API-Key": "wrong-secret-key"}, 401),
    ("Valid X-API-Key Header", f"{BASE}/api/attendance/daily?classroom=401", {"X-API-Key": VALID_KEY}, 200),
    ("Valid Authorization Bearer Header", f"{BASE}/api/attendance/daily?classroom=401", {"Authorization": f"Bearer {VALID_KEY}"}, 200),
    ("Valid Authorization ApiKey Header", f"{BASE}/api/attendance/daily?classroom=401", {"Authorization": f"ApiKey {VALID_KEY}"}, 200),
    ("Valid Query Parameter api_key", f"{BASE}/api/attendance/daily?classroom=401&api_key={VALID_KEY}", {}, 200),
    ("Invalid Query Parameter api_key", f"{BASE}/api/attendance/daily?classroom=401&api_key=invalid_key", {}, 401),
]

print(f"{'Status':<6} | {'HTTP':<8} | Test Case")
print("-" * 65)

passed = 0
for name, url, headers, exp_code in auth_cases:
    try:
        req = urllib.request.Request(url, headers=headers)
        with urllib.request.urlopen(req) as resp:
            status = "PASS" if resp.status == exp_code else "FAIL"
            if status == "PASS": passed += 1
            print(f"{status:<6} | {resp.status} (exp {exp_code}) | {name}")
    except urllib.error.HTTPError as e:
        status = "PASS" if e.code == exp_code else "FAIL"
        if status == "PASS": passed += 1
        print(f"{status:<6} | {e.code} (exp {exp_code}) | {name}")
    except Exception as e:
        print(f"FAIL   | ERR      | {name}: {e}")

print("-" * 65)
print(f"Total: {len(auth_cases)} | Passed: {passed} | Failed: {len(auth_cases) - passed}")
