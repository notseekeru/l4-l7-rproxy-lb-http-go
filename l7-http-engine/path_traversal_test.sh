#!/usr/bin/env bash
# Path Traversal Test Suite for http-engine-go
PASS=0
FAIL=0

ok()   { echo "  ✓ $1"; ((PASS++)); }
fail() { echo "  ✗ $1"; ((FAIL++)); }
check() { local label="$1" pattern="$2" input="$3"
  if echo "$input" | grep -q "$pattern"; then ok "$label"; else fail "$label"; fi
}
req()  { curl -s -o /dev/null -w "%{http_code}" "$@" 2>&1 || true; }
get()  { curl -s --raw "$@" 2>&1 || true; }

HOST="${1:-localhost:8080}"

# --- start server ---
cleanup() { kill "$SERVER_PID" 2>/dev/null || true; }
trap cleanup EXIT

./bin/http-go-engine &
SERVER_PID=$!
sleep 0.3

echo "=== 1. Direct dot-dot-slash ==="
R=$(req "http://$HOST/../../etc/passwd")
echo "  HTTP $R"
check "/../../etc/passwd -> 404" "404" "$R"

echo ""
echo "=== 2. URL-encoded dot-dot-slash ==="
R=$(req "http://$HOST/..%2f..%2fetc%2fpasswd")
echo "  HTTP $R"
check "URL-encoded ../ -> 404" "404" "$R"

echo ""
echo "=== 3. Path-appended traversal ==="
R=$(req "http://$HOST/styles.css/../../../etc/passwd")
echo "  HTTP $R"
check "path-appended ../ -> 404" "404" "$R"

echo ""
echo "=== 4. Null byte injection ==="
R=$(req "http://$HOST/index.html%00.txt")
echo "  HTTP $R"
check "null byte injection -> 404" "404" "$R"

echo ""
echo "=== 5. Absolute path ==="
R=$(req "http://$HOST/etc/passwd")
echo "  HTTP $R"
check "absolute path -> 404" "404" "$R"

echo ""
echo "=== 6. Double encoding ==="
R=$(req "http://$HOST/%252e%252e%252fetc%252fpasswd")
echo "  HTTP $R"
check "double encoding -> 404" "404" "$R"

echo ""
echo "=== 7. Backslash (Windows) traversal ==="
R=$(req "http://$HOST/..\\..\\..\\windows\\win.ini")
echo "  HTTP $R"
check "backslash traversal -> 404" "404" "$R"

echo ""
echo "=== 8. Baseline: normal / ==="
R=$(get "http://$HOST/")
echo "$R" | head -3
check "normal root request -> 200" "200 OK" "$R"

echo ""
echo "=========================================="
echo "  PASS: $PASS  FAIL: $FAIL"
echo "=========================================="

[ "$FAIL" -eq 0 ]
