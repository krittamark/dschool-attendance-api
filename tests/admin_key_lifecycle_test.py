import urllib.request
import urllib.parse
import json

BASE = "http://localhost:8080"
ADMIN_KEY = "dschool-admin-master-key-2026"
CLIENT_KEY = "dschool-secret-key-2026"

print("==================================================================")
print("     ADMIN API KEY LIFECYCLE & REVOCATION INTEGRITY TEST         ")
print("==================================================================")

def run_test(name, fn):
    try:
        passed, msg = fn()
        status = "PASS" if passed else "FAIL"
        print(f"[{status}] {name} -> {msg}")
        return passed
    except Exception as e:
        print(f"[FAIL] {name} -> Error: {e}")
        return False

# 1. Test Access without Admin Key
def test_no_admin_key():
    req = urllib.request.Request(f"{BASE}/api/admin/keys", headers={"X-API-Key": CLIENT_KEY})
    try:
        urllib.request.urlopen(req)
        return False, "Expected 403 Forbidden, but got 200"
    except urllib.error.HTTPError as e:
        return (e.code == 403), f"Correctly rejected with HTTP {e.code}"

# 2. Test List Keys with Admin Key
def test_list_keys():
    req = urllib.request.Request(f"{BASE}/api/admin/keys", headers={"X-API-Key": ADMIN_KEY})
    with urllib.request.urlopen(req) as resp:
        data = json.loads(resp.read().decode("utf-8"))
        keys = data["data"]
        return (resp.status == 200 and len(keys) >= 2), f"Found {len(keys)} keys in keystore"

# 3. Test Create New Key
new_created_key = None
def test_create_key():
    global new_created_key
    payload = json.dumps({"name": "Test LINE Bot Integration", "role": "client"}).encode("utf-8")
    req = urllib.request.Request(
        f"{BASE}/api/admin/keys",
        data=payload,
        headers={"X-API-Key": ADMIN_KEY, "Content-Type": "application/json"},
        method="POST"
    )
    with urllib.request.urlopen(req) as resp:
        data = json.loads(resp.read().decode("utf-8"))
        new_created_key = data["data"]["key"]
        return (resp.status == 201 and new_created_key.startswith("dsk_live_")), f"Created key: {new_created_key}"

# 4. Test Use New Key on Protected Endpoint
def test_use_new_key():
    req = urllib.request.Request(f"{BASE}/api/attendance/daily?classroom=401", headers={"X-API-Key": new_created_key})
    with urllib.request.urlopen(req) as resp:
        return (resp.status == 200), "Successfully queried daily attendance with new key"

# 5. Test Revoke Key
def test_revoke_key():
    payload = json.dumps({"key": new_created_key}).encode("utf-8")
    req = urllib.request.Request(
        f"{BASE}/api/admin/keys/revoke",
        data=payload,
        headers={"X-API-Key": ADMIN_KEY, "Content-Type": "application/json"},
        method="POST"
    )
    with urllib.request.urlopen(req) as resp:
        data = json.loads(resp.read().decode("utf-8"))
        is_active = data["data"]["is_active"]
        return (resp.status == 200 and is_active is False), "Key marked is_active=false"

# 6. Test Using Revoked Key (Must fail with 401)
def test_use_revoked_key():
    req = urllib.request.Request(f"{BASE}/api/attendance/daily?classroom=401", headers={"X-API-Key": new_created_key})
    try:
        urllib.request.urlopen(req)
        return False, "Expected 401 Unauthorized, but got 200"
    except urllib.error.HTTPError as e:
        return (e.code == 401), f"Correctly rejected revoked key with HTTP {e.code}"

# 7. Test Re-activate Key
def test_activate_key():
    payload = json.dumps({"key": new_created_key}).encode("utf-8")
    req = urllib.request.Request(
        f"{BASE}/api/admin/keys/activate",
        data=payload,
        headers={"X-API-Key": ADMIN_KEY, "Content-Type": "application/json"},
        method="POST"
    )
    with urllib.request.urlopen(req) as resp:
        data = json.loads(resp.read().decode("utf-8"))
        is_active = data["data"]["is_active"]
        return (resp.status == 200 and is_active is True), "Key re-activated successfully"

# 8. Test Using Re-activated Key
def test_use_activated_key():
    req = urllib.request.Request(f"{BASE}/api/attendance/daily?classroom=401", headers={"X-API-Key": new_created_key})
    with urllib.request.urlopen(req) as resp:
        return (resp.status == 200), "Successfully queried daily attendance again after activation"

# 9. Test Prevent Revocation of Master Admin Key
def test_protect_master_key():
    payload = json.dumps({"key": ADMIN_KEY}).encode("utf-8")
    req = urllib.request.Request(
        f"{BASE}/api/admin/keys/revoke",
        data=payload,
        headers={"X-API-Key": ADMIN_KEY, "Content-Type": "application/json"},
        method="POST"
    )
    try:
        urllib.request.urlopen(req)
        return False, "Expected 403 Forbidden"
    except urllib.error.HTTPError as e:
        return (e.code == 403), f"Master Admin Key protected with HTTP {e.code}"

tests = [
    ("1. Reject Client Key on Admin API", test_no_admin_key),
    ("2. List Keys with Admin Key", test_list_keys),
    ("3. Create New API Key (dsk_live_...)", test_create_key),
    ("4. Use New Key on Data Endpoint", test_use_new_key),
    ("5. Revoke API Key via Admin API", test_revoke_key),
    ("6. Immediate 401 Block on Revoked Key", test_use_revoked_key),
    ("7. Re-activate Key via Admin API", test_activate_key),
    ("8. Re-activated Key Immediately Works", test_use_activated_key),
    ("9. Protect Master Admin Key from Revocation", test_protect_master_key),
]

passed_count = 0
for name, fn in tests:
    if run_test(name, fn):
        passed_count += 1

print("-" * 66)
print(f"Result: {passed_count}/{len(tests)} Passed ({(passed_count/len(tests))*100:.1f}%)")
print("==================================================================")
