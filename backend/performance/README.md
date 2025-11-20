# Concurrent user validation

This check is tailored to the "Support Concurrent Users" use case (50–100 simultaneous users) with the success criteria:

- API responses stay under 500ms.
- Database queries complete efficiently while requests are concurrent.
- TCP/WebSocket connections remain stable under load.
- No data corruption or dropped writes under concurrent access.

## What is covered

The `concurrency.js` k6 script exercises the production entrypoints that already exist in the backend:

- HTTP: `/manga/popular`, `/manga/search`, and `/manga/1` for cached and uncached paths.
- WebSocket: `/ws` handshake and hold connections to detect churn or disconnects during bursts.
- TCP listener: use `tcp_smoke.go` to open many short-lived sockets against the broadcast server to verify it accepts concurrent connections.

Thresholds enforce the <500ms latency target for p(95) HTTP requests and require >99% successful WebSocket/TCP connection attempts. You can adjust VUs or duration to reflect different traffic profiles.

## How to run

1. Start the API server on port 8080 and the WebSocket server on 8081 (default from `cmd/api-server` and `cmd/ws-server`). Ensure the TCP broadcaster is enabled if you want the TCP check.
2. From the repository root, run:

   ```bash
   BASE_URL=http://localhost:8080 \
   WS_URL=ws://localhost:8081/ws \
   k6 run backend/performance/concurrency.js
   ```

3. Review the summary at the end of the run. A passing run keeps HTTP p(95) < 500ms, maintains >99% successful socket handshakes, and shows no threshold failures.

4. (Optional) Validate the TCP broadcaster with concurrent sockets:

   ```bash
   go run ./backend/performance/tcp_smoke.go -addr localhost:9000 -clients 100 -hold 2s
   ```

## Interpreting results

- **HTTP thresholds** validate the 500ms SLA and will fail if median or p(95) latency rises.
- **WebSocket and TCP connection rates** highlight instability (disconnects or refused connections) during spikes.
- **Error samples** printed from the script show which endpoint or protocol failed so the underlying handler can be tuned.

## Extending

- Add authenticated routes by exporting `AUTH_TOKEN` and including it in the `Authorization` header inside the script.
- Point `BASE_URL`, `WS_URL`, or `TCP_ADDR` to staging/production to compare environments.
- Raise `MAX_VUS` or extend the duration to explore headroom beyond 100 concurrent users.
