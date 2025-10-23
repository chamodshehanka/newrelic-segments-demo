// k6 load test for New Relic segments demo
// - Hits /process-untraced and /process-traced on chamod service
// - Adds a unique X-Request-Id header for log correlation
// - Includes basic checks and configurable stages

import http from 'k6/http';
import { check, sleep } from 'k6';

// ----- Configuration -----
export let options = {
  // Example: ramp up to 20 VUs over 30s, hold 60s, ramp down 10s
  stages: [
    { duration: '15s', target: 5 },
    { duration: '30s', target: 20 },
    { duration: '60s', target: 20 },
    { duration: '10s', target: 0 },
  ],
  // Fail the test if more than 1% of requests fail
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<2000'], // 95% of requests should finish within 2s (adjust as needed)
  },
};

// Base host
const BASE_URL = 'http://localhost:8080';

// Choose endpoint distribution: 50% traced, 50% untraced by default
function pickEndpoint() {
  return Math.random() < 0.5 ? '/process-traced' : '/process-untraced';
}

// Build a simple unique request id using VU and iteration (k6 globals)
function makeRequestId() {
  return `${__VU}-${__ITER}-${Date.now()}`;
}

export default function () {
  const endpoint = pickEndpoint();
  const url = `${BASE_URL}${endpoint}`;

  const headers = {
    'X-Request-Id': makeRequestId(),
    'Accept': 'application/json',
  };

  const res = http.get(url, { headers: headers, tags: { endpoint: endpoint } });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response is json': (r) => r.headers['Content-Type'] && r.headers['Content-Type'].includes('application/json'),
  });

  // wait a short random time between requests to simulate real traffic
  sleep(Math.random() * 1.5 + 0.1);
}

// End of script

