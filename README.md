# HTTP Engine Go

A monorepo of networking components built from scratch in Go — **no frameworks, no third-party dependencies**, just the standard library touching the raw TCP wire.

The goal is to understand how HTTP and TCP actually work under the hood: what `net/http` hides, what a load balancer really does, and how the pieces of the network stack fit together.

```
┌─────────────────────────────────────────────────────────────┐
│                         Client                               │
└──────────────────────────────┬──────────────────────────────┘
                               │ TCP
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  l4-rproxy-lb   (Layer 4 reverse proxy / load balancer)      │
│  listens on :9000 · round-robins to backends                 │
└──────┬──────────────────────────────┬───────────────────────┘
       │ TCP                         │ TCP
       ▼                             ▼
┌────────────────────┐   ┌────────────────────┐
│  l7-http-engine    │   │  l7-http-engine     │
│  backend on :8080  │   │  backend on :8081   │
│  raw HTTP/1.1      │   │  raw HTTP/1.1       │
└────────────────────┘   └────────────────────┘
```

## Projects

### [`l7-http-engine/`](l7-http-engine/README.md) — Minimal HTTP/1.1 Engine

A hand-rolled HTTP/1.1 server built directly on `net.Listen`, `net.Accept`, and `net.Conn`. No `net/http`.

- Raw TCP listener with per-connection goroutines and a 5s connection deadline
- Keep-alive request loop, `bufio`-buffered parsing
- Request-line / header / query-string parsing with validation
- `GET` + `POST` only (`405` otherwise), `HTTP/1.1` only (`400` otherwise)
- Static file serving with MIME detection and directory-traversal protection
- Hand-built response builder (`HTTPMessage` / `HTTPFileServe`)
- POST body reading capped at ~10 MB via `io.LimitReader`

### [`l4-rproxy-lb/`](l4-rproxy-lb/README.md) — Layer 4 Reverse Proxy / Load Balancer

A TCP-level reverse proxy that sits in front of backends and balances connections.

- Listens on `:9000`, accepts connections in a loop
- Round-robins each connection to a backend (`localhost:8080`, `localhost:8081`)
- Bidirectional byte relay with `io.Copy` — client → server and server → client
- Protocol-agnostic: it balances raw TCP streams, not HTTP requests

## Requirements

| Tool  | Version         |
|-------|-----------------|
| Go    | 1.25+ (L7 uses `go 1.26.3`, L4 uses `go 1.25.0`) |

## Getting Started

Clone, then build and run each component:

```sh
# 1. Start a couple of backend HTTP engines
cd l7-http-engine && make build && ./bin/http-go-engine        # :8081

# 2. Start the load balancer in front of them
cd l4-rproxy-lb && make run                                    # :9000

# 3. Hit the balancer
curl -v http://localhost:9000/
```

Each project has its own `Makefile` with shared targets:

| Target  | Description                                   |
|---------|-----------------------------------------------|
| `make run`    | Run the server (`go run`)                    |
| `make build`  | Build a stripped binary into `bin/`          |
| `make test`   | Run tests with `-race -buildvcs`             |
| `make audit`  | `go mod verify` + `go vet` + `golangci-lint` |
| `make tidy`   | `go fmt` + `go mod tidy`                     |
| `make clean`  | Remove build artifacts and coverage          |

## Testing

The L7 engine ships a scripted end-to-end suite that exercises the wire protocol:

```sh
cd l7-http-engine
./test_endpoints.sh        # status codes, keep-alive, malformed requests
./path_traversal_test.sh   # directory-traversal and injection attempts
```

## Design Philosophy

**Scope:** This applies to both projects in the monorepo — `l7-http-engine` and `l4-rproxy-lb` — unless noted otherwise.

I wrote this code entirely myself without relying on AI for core logic or conventions. I used AI solely for syntax, boilerplate generation, as a code-review tool, and as a critique tool to better understand underlying systems and hand-picked advised implementation on engineering the systems. AI is also used to surface edge cases, memory-safety concerns, and robustness gaps — which I then investigate and fix myself rather than accepting AI-generated fixes wholesale.

AI can also be unreliable at times — for example its understanding of path-traversal defenses isn't always trustworthy. To verify that surface, I wrote `l7-http-engine/path_traversal_test.sh`, which exercises traversal and injection attempts and confirms the server rejects them. It may be that I've missed something and it's still vulnerable, but for now that behavior is verified by the script.

The systems might be flawed as they don't include extensive error handling and may contain logical bugs — these are learning projects, not production servers. Where AI was used to produce boilerplate reasoning for specific features, those notes live in the relevant project README.

- **Standard library only.** Every byte crosses your own code before it hits the socket.
- **Educational first.** Error handling is deliberately lean; this is a learning project, not a production server.
- **Explicit over abstract.** Routing is a `switch`, headers are a `map`, responses are hand-assembled strings — you see the whole pipeline.
- **YAGNI.** No routing frameworks, no middleware, no HTTP/2, no connection pooling — those are listed as deliberate omissions in the project READMEs.

## License

Released under the [MIT License](LICENSE). This is a personal learning project — see the license file for the full terms.
