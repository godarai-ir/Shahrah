import http from 'k6/http';
import { check, group, sleep } from 'k6';
import exec from 'k6/execution';
import { Counter, Rate, Trend } from 'k6/metrics';

const baseUrl = (__ENV.BASE_URL || 'http://localhost:4000').replace(/\/$/, '');
const scenario = __ENV.SCENARIO || 'mixed';
const apiKey = __ENV.API_KEY || __ENV.MASTER_KEY || 'sk-dev-master-key-12345';
const tenant = __ENV.TENANT || 'tenant1';
const model = __ENV.MODEL || 'gpt-4o-mini';
const timeout = __ENV.TIMEOUT || '60s';

const gatewayErrors = new Counter('gateway_errors');
const cacheHits = new Counter('gateway_cache_hits');
const successfulGatewayCalls = new Rate('gateway_success_rate');
const upstreamDuration = new Trend('gateway_upstream_duration', true);

const scenarios = {
  smoke: {
    executor: 'constant-vus',
    vus: Number(__ENV.VUS || 2),
    duration: __ENV.DURATION || '1m',
  },
  ramp: {
    executor: 'ramping-vus',
    stages: [
      { duration: __ENV.RAMP_UP || '5m', target: Number(__ENV.TARGET_VUS || 50) },
      { duration: __ENV.HOLD || '10m', target: Number(__ENV.TARGET_VUS || 50) },
      { duration: __ENV.RAMP_DOWN || '2m', target: 0 },
    ],
  },
  spike: {
    executor: 'ramping-vus',
    stages: [
      { duration: '1m', target: Number(__ENV.BASELINE_VUS || 20) },
      { duration: '30s', target: Number(__ENV.SPIKE_VUS || 200) },
      { duration: '2m', target: Number(__ENV.SPIKE_VUS || 200) },
      { duration: '30s', target: Number(__ENV.BASELINE_VUS || 20) },
      { duration: '1m', target: 0 },
    ],
  },
  soak: {
    executor: 'constant-vus',
    vus: Number(__ENV.VUS || 40),
    duration: __ENV.DURATION || '2h',
  },
  cache: {
    executor: 'constant-arrival-rate',
    rate: Number(__ENV.RATE || 30),
    timeUnit: '1s',
    duration: __ENV.DURATION || '15m',
    preAllocatedVUs: Number(__ENV.PREALLOCATED_VUS || 50),
    maxVUs: Number(__ENV.MAX_VUS || 200),
  },
  mixed: {
    executor: 'ramping-arrival-rate',
    startRate: Number(__ENV.START_RATE || 5),
    timeUnit: '1s',
    preAllocatedVUs: Number(__ENV.PREALLOCATED_VUS || 50),
    maxVUs: Number(__ENV.MAX_VUS || 300),
    stages: [
      { duration: __ENV.RAMP_UP || '5m', target: Number(__ENV.TARGET_RATE || 80) },
      { duration: __ENV.HOLD || '15m', target: Number(__ENV.TARGET_RATE || 80) },
      { duration: __ENV.RAMP_DOWN || '3m', target: 0 },
    ],
  },
};

export const options = {
  scenarios: {
    [scenario]: scenarios[scenario] || scenarios.mixed,
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<30000', 'p(99)<60000'],
    gateway_success_rate: ['rate>0.95'],
  },
};

function headers(extra = {}) {
  return {
    Authorization: `Bearer ${apiKey}`,
    'Content-Type': 'application/json',
    'x-godarai-tenant': tenant,
    ...extra,
  };
}

function record(name, response) {
  const ok = check(response, {
    [`${name}: status is 2xx`]: (r) => r.status >= 200 && r.status < 300,
    [`${name}: has body`]: (r) => r.body && r.body.length > 0,
  });

  successfulGatewayCalls.add(ok);
  if (!ok) {
    gatewayErrors.add(1);
  }

  const cacheStatus = response.headers['X-Godarai-Cache-Status'] || response.headers['x-godarai-cache-status'];
  if (cacheStatus && cacheStatus.toLowerCase() === 'hit') {
    cacheHits.add(1);
  }

  upstreamDuration.add(response.timings.duration);
}

function chat(cacheMode = 'off') {
  const prompt = cacheMode === 'repeat'
    ? 'Summarize Shahrah gateway load-test behavior in one short paragraph.'
    : `Summarize request ${exec.scenario.iterationInTest} for tenant gateway testing.`;

  const response = http.post(`${baseUrl}/v1/chat/completions`, JSON.stringify({
    model,
    messages: [
      { role: 'system', content: 'You are a concise gateway test assistant.' },
      { role: 'user', content: prompt },
    ],
    temperature: 0.1,
    max_tokens: 96,
  }), {
    headers: headers(cacheMode === 'repeat' ? { 'x-godarai-cache-config': 'semantic=true,ttl=600' } : {}),
    timeout,
    tags: { endpoint: 'chat_completions', cache_mode: cacheMode },
  });

  record('chat completions', response);
}

function responses() {
  const response = http.post(`${baseUrl}/v1/responses`, JSON.stringify({
    model,
    input: `Give one operational risk for load test iteration ${exec.scenario.iterationInTest}.`,
    max_output_tokens: 96,
  }), {
    headers: headers(),
    timeout,
    tags: { endpoint: 'responses' },
  });

  record('responses', response);
}

function embeddings() {
  const response = http.post(`${baseUrl}/v1/embeddings`, JSON.stringify({
    model: __ENV.EMBEDDING_MODEL || model,
    input: [`gateway load test ${exec.scenario.iterationInTest}`, 'latency and contention profiling'],
  }), {
    headers: headers(),
    timeout,
    tags: { endpoint: 'embeddings' },
  });

  record('embeddings', response);
}

function models() {
  const response = http.get(`${baseUrl}/v1/models`, {
    headers: headers(),
    timeout,
    tags: { endpoint: 'models' },
  });

  record('models', response);
}

function health() {
  const response = http.get(`${baseUrl}/health`, {
    timeout: '5s',
    tags: { endpoint: 'health' },
  });

  record('health', response);
}

export default function () {
  if (scenario === 'cache') {
    group('cache repeated prompt', () => chat('repeat'));
    sleep(Number(__ENV.SLEEP || 0.2));
    return;
  }

  const roll = Math.random();
  if (roll < 0.50) {
    group('chat completions', () => chat('off'));
  } else if (roll < 0.70) {
    group('responses', responses);
  } else if (roll < 0.85) {
    group('embeddings', embeddings);
  } else if (roll < 0.95) {
    group('models', models);
  } else {
    group('health', health);
  }

  sleep(Number(__ENV.SLEEP || 1));
}
