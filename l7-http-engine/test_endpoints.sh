#!/usr/bin/env bash
# Usage: ./test_endpoints.sh [host:port]
PASS=0
FAIL=0

ok()   { echo "  ✓ $1"; ((PASS++)); }
fail() { echo "  ✗ $1"; ((FAIL++)); }
req()  { timeout 2 nc "$@" 2>&1 || true; }
check() { local label="$1" pattern="$2" input="$3"
  if echo "$input" | grep -q "$pattern"; then ok "$label"; else fail "$label"; fi
}

HOST="${1:-localhost:8080}"

cleanup() { kill "$SERVER_PID" 2>/dev/null || true; }
trap cleanup EXIT

/tmp/http-engine-test &
SERVER_PID=$!
sleep 0.3

R=$(printf "GET / HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
echo "$R"
check "/ returns 200"            "200 OK"           "$R"
check "/ HTML body"              "<!DOCTYPE html>"  "$R"
check "/ Content-Type text/html" "Content-Type: text/html" "$R"

R=$(printf "GET /ping HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
check "/ping returns 200" "200 OK" "$R"
check "/ping body pong"   "pong"   "$R"

R=$(printf "GET /nonexistent HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
check "/nonexistent 404" "404 Not Found" "$R"

R=$(printf "GET /ping?foo=bar&baz=qux HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
check "relativeURI params 200" "200 OK" "$R"

R=$(printf "GET /ping?foo=bar& HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
if ! echo "$R" | grep -qi "panic"; then ok "trailing & no crash"; else fail "trailing & no crash"; fi

R=$(printf "GET /ping?foo HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
if ! echo "$R" | grep -qi "panic"; then ok "no-value relativeURI no crash"; else fail "no-value relativeURI no crash"; fi

R=$(printf "POST /ping HTTP/1.1\r\nHost: $HOST\r\nContent-Length: 11\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\nhello world" | req localhost 8080)
check "POST /ping 200" "200 OK" "$R"

R=$(printf "PUT /ping HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
check "PUT 405" "405" "$R"

R=$(printf "GET /ping HTTP/1.0\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
check "HTTP/1.0 400" "400" "$R"

R=$(printf "GET /ping HTTP/1.1 extra\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | req localhost 8080)
check "extra param 400" "400\|Too many" "$R"

R=$(printf "GET /ping HTTP/1.1\r\nHost: $HOST\r\n\r\nGET /ping HTTP/1.1\r\nHost: $HOST\r\nConnection: close\r\n\r\n" | timeout 3 nc localhost 8080 2>&1 || true)
COUNT=$(echo "$R" | grep -c "200 OK" || true)
[ "$COUNT" -eq 2 ] && ok "keep-alive 2x200" || fail "keep-alive 2x200 (got $COUNT)"

R=$(printf "POST /ping HTTP/1.1\r\nHost: $HOST\r\nContent-Length: 0\r\nConnection: close\r\n\r\n" | req localhost 8080)
check "Content-Length 0 200" "200 OK" "$R"

echo "  PASS: $PASS  FAIL: $FAIL"

[ "$FAIL" -eq 0 ]
