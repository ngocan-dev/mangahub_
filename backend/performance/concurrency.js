import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const WS_URL = __ENV.WS_URL || 'ws://localhost:8081/ws';
const MAX_VUS = Number(__ENV.MAX_VUS || 100);
const WS_VUS = Number(__ENV.WS_VUS || 50);
const WS_HOLD_MS = Number(__ENV.WS_HOLD_MS || 5000);

const httpErrorRate = new Rate('http_error_rate');
const wsConnectionSuccess = new Rate('ws_connection_success');

const thresholds = {
  http_req_duration: ['p(95)<500'],
  http_error_rate: ['rate<0.01'],
  ws_connection_success: ['rate>0.99'],
  checks: ['rate>0.99'],
};

const scenarios = {
  browse: {
    executor: 'constant-vus',
    vus: MAX_VUS,
    duration: '1m',
    exec: 'browse',
  },
  websockets: {
    executor: 'constant-vus',
    vus: WS_VUS,
    duration: '1m',
    exec: 'wsHold',
    startTime: '0s',
  },
};

export const options = { scenarios, thresholds };

function authHeaders() {
  const token = __ENV.AUTH_TOKEN;
  if (!token) return {};
  return { Authorization: `Bearer ${token}` };
}

function httpStep(url, description) {
  const res = http.get(url, { headers: authHeaders(), tags: { description } });
  const ok = check(res, {
    [`${description} status`]: (r) => r.status === 200,
    [`${description} <500ms`]: (r) => r.timings.duration < 500,
  });
  httpErrorRate.add(ok ? 0 : 1);
  if (!ok) {
    console.error(`${description} failed`, res.status, res.body?.slice?.(0, 120));
  }
  return res;
}

export function browse() {
  httpStep(`${BASE_URL}/manga/popular`, 'popular manga');
  httpStep(`${BASE_URL}/manga/search?query=action`, 'search manga');
  httpStep(`${BASE_URL}/manga/1`, 'manga details');
  sleep(1);
}

export function wsHold() {
  let success = 0;
  ws.connect(WS_URL, { headers: authHeaders() }, (socket) => {
    socket.on('open', () => {
      success = 1;
    });

    socket.on('error', (e) => {
      success = 0;
      console.error('websocket error', e.error());
    });

    socket.setTimeout(() => {
      socket.close();
    }, WS_HOLD_MS);
  });
  wsConnectionSuccess.add(success);
}
