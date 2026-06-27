#!/bin/bash
# Tests all CRUD methods against deployed Apps Script web app.
# Usage: ./test.sh [URL]

set -euo pipefail

_REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEPLOYMENT_ID=$(grep '^deploymentId:' "$_REPO_ROOT/config/sheets-deployment.yaml" | sed 's/^deploymentId:[[:space:]]*//')
BASE_URL="${1:-https://script.google.com/macros/s/${DEPLOYMENT_ID}/exec}"
TEST_COMPANY="__TEST_CORP_$(date +%s)__"
PASS=0
FAIL=0

green() { printf "\033[32m%s\033[0m\n" "$1"; }
red()   { printf "\033[31m%s\033[0m\n" "$1"; }
info()  { printf "\033[36m%s\033[0m\n" "$1"; }

# GET requests return JSON directly with -L
get() {
  local url="$1"
  curl -s -L "$url"
}

# POST requests get 302 redirected; curl -L converts to GET.
# Capture redirect Location, then GET that URL to get the JSON response.
post() {
  local body="$1"
  local headers
  headers=$(curl -s -D - -o /dev/null -X POST "$BASE_URL" \
    -H "Content-Type: application/json" \
    -d "$body")
  local location
  location=$(echo "$headers" | grep -i "^location:" | tail -1 | sed 's/^[Ll]ocation: //;s/\r$//')
  if [ -z "$location" ]; then
    echo '{"status":"error","message":"No redirect Location header"}'
    return
  fi
  curl -s -L "$location"
}

assert_eq() {
  local label="$1" expected="$2" actual="$3"
  if [ "$expected" = "$actual" ]; then
    green "PASS: $label"
    PASS=$((PASS + 1))
  else
    red "FAIL: $label — expected '$expected', got '$actual'"
    FAIL=$((FAIL + 1))
  fi
}

assert_contains() {
  local label="$1" haystack="$2" needle="$3"
  if echo "$haystack" | grep -qF -- "$needle"; then
    green "PASS: $label"
    PASS=$((PASS + 1))
  else
    red "FAIL: $label — '$needle' not found in response"
    FAIL=$((FAIL + 1))
  fi
}

cleanup() {
  info "Cleaning up test entry..."
  post "{\"action\":\"delete\",\"matchBy\":{\"companyName\":\"$TEST_COMPANY\"}}" > /dev/null 2>&1 || true
}

trap cleanup EXIT

info "=== Testing CRUD against $BASE_URL ==="
info "Test company: $TEST_COMPANY"
echo ""

# 1. READ — verify API responds and has data
info "[1/4] READ"
read_resp=$(get "${BASE_URL}?page=1&pageSize=5&order=desc")
assert_contains "read returns success" "$read_resp" '"status":"success"'
assert_contains "read has rows array" "$read_resp" '"rows":['
assert_contains "read has pagination" "$read_resp" '"totalPages"'
echo ""

# 2. CREATE — insert test entry
info "[2/4] CREATE"
create_body=$(cat <<EOF
{"action":"create","companyName":"$TEST_COMPANY","link":"https://test.example.com","dateApplied":"2026-06-27","industry":"Tech","phoneNumber":"5551234567","email":"test@test.com","status":"Applied Only","notes":"test entry"}
EOF
)
create_resp=$(post "$create_body")
assert_contains "create returns success" "$create_resp" '"status":"success"'
echo ""

# Verify create by reading it back
info "[2b/4] VERIFY CREATE"
verify_create=$(get "${BASE_URL}?page=1&pageSize=5&search=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$TEST_COMPANY'))")")
assert_contains "created entry visible in search" "$verify_create" "$TEST_COMPANY"
assert_contains "created entry has correct status" "$verify_create" 'Applied Only'
echo ""

# 3. PATCH — update status
info "[3/4] PATCH"
patch_body=$(cat <<EOF
{"action":"patch","matchBy":{"companyName":"$TEST_COMPANY"},"update":{"status":"Interview!","notes":"patched"}}
EOF
)
patch_resp=$(post "$patch_body")
assert_contains "patch returns success" "$patch_resp" '"status":"success"'
echo ""

# Verify patch
info "[3b/4] VERIFY PATCH"
verify_patch=$(get "${BASE_URL}?page=1&pageSize=5&search=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$TEST_COMPANY'))")")
assert_contains "patched status is Interview!" "$verify_patch" 'Interview!'
assert_contains "patched notes updated" "$verify_patch" 'patched'
echo ""

# 4. DELETE — remove test entry
info "[4/4] DELETE"
delete_body=$(cat <<EOF
{"action":"delete","matchBy":{"companyName":"$TEST_COMPANY"}}
EOF
)
delete_resp=$(post "$delete_body")
assert_contains "delete returns success" "$delete_resp" '"status":"success"'
echo ""

# Verify delete
info "[4b/4] VERIFY DELETE"
verify_delete=$(get "${BASE_URL}?page=1&pageSize=5&search=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$TEST_COMPANY'))")")
assert_contains "deleted entry gone from search" "$verify_delete" '"rows":[]'
echo ""

info "=== Results: $PASS passed, $FAIL failed ==="
exit $FAIL
