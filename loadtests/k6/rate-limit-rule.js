import http from 'k6/http';
import { check, fail, sleep } from 'k6';
import exec from 'k6/execution';
import { Counter, Trend } from 'k6/metrics';

const baseUrl = (__ENV.BASE_URL || 'http://localhost:4000').replace(/\/$/, '');
const apiKey = __ENV.API_KEY || __ENV.MASTER_KEY || 'sk-dev-master-key-12345';
const tenant = __ENV.TENANT || 'tenant1';
const model = __ENV.MODEL || 'gpt-4o-mini';
const timeout = __ENV.TIMEOUT || '30s';

const ruleLimit = positiveInteger('RULE_LIMIT', __ENV.RULE_LIMIT || 10);
const windowSeconds = positiveInteger('WINDOW_SECONDS', __ENV.WINDOW_SECONDS || 60);
const burstRequests = positiveInteger('BURST_REQUESTS', __ENV.BURST_REQUESTS || ruleLimit * 2);
const burstVUs = positiveInteger('BURST_VUS', __ENV.BURST_VUS || Math.min(burstRequests, 20));
const recoveryRequests = positiveInteger('RECOVERY_REQUESTS', __ENV.RECOVERY_REQUESTS || Math.min(ruleLimit, 3));
const graceSeconds = nonNegativeNumber('GRACE_SECONDS', __ENV.GRACE_SECONDS || 2);
const alignWindow = booleanEnv('ALIGN_WINDOW', true);
const expectRetryAfter = booleanEnv('EXPECT_RETRY_AFTER', true);

if (burstRequests <= ruleLimit) {
  throw new Error(`BURST_REQUESTS (${burstRequests}) must be greater than RULE_LIMIT (${ruleLimit})`);
}
if (recoveryRequests > ruleLimit) {
  throw new Error(`RECOVERY_REQUESTS (${recoveryRequests}) must not exceed RULE_LIMIT (${ruleLimit})`);
}

const allowed = new Counter('rate_limit_allowed');
const rejected = new Counter('rate_limit_rejected');
const recovered = new Counter('rate_limit_recovered');
const unexpectedStatuses = new Counter('rate_limit_unexpected_statuses');
const contractViolations = new Counter('rate_limit_contract_violations');
const retryAfterSeconds = new Trend('rate_limit_retry_after_seconds');
const expectedStatuses = http.expectedStatuses({ min: 200, max: 299 }, 429);

export const options = {
  setupTimeout: `${windowSeconds + graceSeconds + 30}s`,
  scenarios: {
    burst: {
      executor: 'shared-iterations',
      exec: 'burst',
      vus: burstVUs,
      iterations: burstRequests,
      maxDuration: __ENV.BURST_MAX_DURATION || '1m',
    },
    recovery: {
      executor: 'shared-iterations',
      exec: 'recovery',
      startTime: `${windowSeconds + graceSeconds}s`,
      vus: Math.min(recoveryRequests, burstVUs),
      iterations: recoveryRequests,
      maxDuration: __ENV.RECOVERY_MAX_DURATION || '1m',
    },
  },
  thresholds: {
    checks: ['rate==1'],
    http_req_failed: ['rate==0'],
    'rate_limit_allowed{phase:burst}': [`count==${ruleLimit}`],
    'rate_limit_rejected{phase:burst}': [`count==${burstRequests - ruleLimit}`],
    'rate_limit_recovered{phase:recovery}': [`count==${recoveryRequests}`],
    rate_limit_unexpected_statuses: ['count==0'],
    rate_limit_contract_violations: ['count==0'],
  },
};

function positiveInteger(name, value) {
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error(`${name} must be a positive integer; received ${value}`);
  }
  return parsed;
}

function nonNegativeNumber(name, value) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed < 0) {
    throw new Error(`${name} must be a non-negative number; received ${value}`);
  }
  return parsed;
}

function booleanEnv(name, fallback) {
  const raw = __ENV[name];
  if (raw === undefined || raw === '') return fallback;
  return ['1', 'true', 'yes', 'on'].includes(raw.toLowerCase());
}

function header(response, name) {
  const expected = name.toLowerCase();
  for (const key of Object.keys(response.headers)) {
    if (key.toLowerCase() === expected) return response.headers[key];
  }
  return '';
}

function request(phase) {
  const iteration = phase === 'preflight' ? 'setup' : exec.scenario.iterationInTest;
  return http.post(`${baseUrl}/v1/chat/completions`, JSON.stringify({
    model,
    messages: [
      { role: 'system', content: 'Return only the word ok.' },
      { role: 'user', content: `Rate-limit test ${phase}-${iteration}` },
    ],
    temperature: 0,
    max_tokens: 8,
  }), {
    headers: {
      Authorization: `Bearer ${apiKey}`,
      'Content-Type': 'application/json',
      'x-godarai-tenant': tenant,
    },
    timeout,
    tags: { endpoint: 'chat_completions', phase, tenant },
    responseCallback: expectedStatuses,
  });
}

function responseErrorType(response) {
  try {
    return response.json('error.type') || '';
  } catch (_) {
    return '';
  }
}

function validateRateLimitHeaders(response, shouldBeRejected) {
  const limit = Number(header(response, 'X-RateLimit-Limit'));
  const remaining = Number(header(response, 'X-RateLimit-Remaining'));
  const reset = Number(header(response, 'X-RateLimit-Reset'));
  const now = Date.now() / 1000;
  let violations = 0;

  if (limit !== ruleLimit) violations += 1;
  if (!Number.isFinite(remaining) || remaining < 0 || remaining >= ruleLimit) violations += 1;
  if (shouldBeRejected && remaining !== 0) violations += 1;
  if (!Number.isFinite(reset) || reset < now - 1 || reset > now + windowSeconds + graceSeconds + 1) violations += 1;

  if (shouldBeRejected && expectRetryAfter) {
    const retryAfter = Number(header(response, 'Retry-After'));
    retryAfterSeconds.add(Number.isFinite(retryAfter) ? retryAfter : 0);
    if (!Number.isFinite(retryAfter) || retryAfter < 0 || retryAfter > windowSeconds + graceSeconds) {
      violations += 1;
    }
  }

  return violations;
}

function classifyBurst(response) {
  const isAllowed = response.status >= 200 && response.status < 300;
  const isRejected = response.status === 429;
  const expected = isAllowed || isRejected;

  allowed.add(isAllowed ? 1 : 0, { phase: 'burst' });
  rejected.add(isRejected ? 1 : 0, { phase: 'burst' });
  unexpectedStatuses.add(expected ? 0 : 1, { phase: 'burst' });

  let violations = validateRateLimitHeaders(response, isRejected);
  if (isRejected && responseErrorType(response) !== 'rate_limit_error') violations += 1;
  contractViolations.add(violations, { phase: 'burst' });

  check(response, {
    'burst status is 2xx or rate-limit 429': () => expected,
    'burst response has expected rate-limit headers': () => violations === 0,
  });
}

export function setup() {
  if (!alignWindow) return;

  const response = request('preflight');
  const observedLimit = Number(header(response, 'X-RateLimit-Limit'));
  if (!((response.status >= 200 && response.status < 300) || response.status === 429)) {
    fail(`Preflight failed with HTTP ${response.status}: ${response.body}`);
  }
  if (observedLimit !== ruleLimit) {
    fail(`Expected X-RateLimit-Limit=${ruleLimit}, received ${header(response, 'X-RateLimit-Limit') || '<missing>'}. Check that the tenant rule is enabled and active.`);
  }

  const resetAt = Number(header(response, 'X-RateLimit-Reset'));
  if (!Number.isFinite(resetAt)) {
    fail('Preflight response is missing a valid X-RateLimit-Reset header.');
  }

  const waitSeconds = Math.max(0, resetAt - Date.now() / 1000) + graceSeconds;
  if (waitSeconds > 0) sleep(waitSeconds);
}

export function burst() {
  classifyBurst(request('burst'));
}

export function recovery() {
  const response = request('recovery');
  const isRecovered = response.status >= 200 && response.status < 300;
  recovered.add(isRecovered ? 1 : 0, { phase: 'recovery' });
  unexpectedStatuses.add(isRecovered ? 0 : 1, { phase: 'recovery' });

  const violations = isRecovered ? validateRateLimitHeaders(response, false) : 1;
  contractViolations.add(violations, { phase: 'recovery' });

  check(response, {
    'request is accepted after window reset': () => isRecovered,
    'recovery response has expected rate-limit headers': () => violations === 0,
  });
}
