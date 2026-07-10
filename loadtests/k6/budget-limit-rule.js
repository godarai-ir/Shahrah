import http from 'k6/http';
import { check, fail, sleep } from 'k6';
import exec from 'k6/execution';
import { Counter, Trend } from 'k6/metrics';

const baseUrl = (__ENV.BASE_URL || 'http://localhost:4000').replace(/\/$/, '');
const apiKey = __ENV.API_KEY || __ENV.MASTER_KEY || 'sk-dev-master-key-12345';
const tenant = __ENV.TENANT || 'tenant1';
const model = __ENV.MODEL || 'gpt-4o-mini';
const timeout = __ENV.TIMEOUT || '30s';
const ruleID = __ENV.BUDGET_RULE_ID || 'budget-loadtest';

const budgetLimitIRR = positiveNumber('BUDGET_LIMIT_IRR', __ENV.BUDGET_LIMIT_IRR || 400);
const costToleranceIRR = nonNegativeNumber('COST_TOLERANCE_IRR', __ENV.COST_TOLERANCE_IRR || 1);
const burstRequests = positiveInteger('BURST_REQUESTS', __ENV.BURST_REQUESTS || 30);
const burstVUs = positiveInteger('BURST_VUS', __ENV.BURST_VUS || Math.min(burstRequests, 30));
const saturationRequests = positiveInteger('SATURATION_REQUESTS', __ENV.SATURATION_REQUESTS || 3);
const minAllowed = positiveInteger('MIN_ALLOWED', __ENV.MIN_ALLOWED || 1);
const minRejected = positiveInteger('MIN_REJECTED', __ENV.MIN_REJECTED || 1);
const maxTokens = positiveInteger('MAX_TOKENS', __ENV.MAX_TOKENS || 8);
const verificationSleep = nonNegativeNumber('VERIFICATION_SLEEP', __ENV.VERIFICATION_SLEEP || 0.1);
const metadata = parseMetadata(__ENV.METADATA_JSON);

if (minAllowed + minRejected > burstRequests) {
  throw new Error('MIN_ALLOWED + MIN_REJECTED must not exceed BURST_REQUESTS');
}

const allowed = new Counter('budget_allowed');
const rejected = new Counter('budget_rejected');
const successfulCostIRR = new Counter('budget_successful_cost_irr');
const saturationRejected = new Counter('budget_saturation_rejected');
const unexpectedStatuses = new Counter('budget_unexpected_statuses');
const contractViolations = new Counter('budget_contract_violations');
const rateLimitInterference = new Counter('budget_rate_limit_interference');
const allowedDuration = new Trend('budget_allowed_duration', true);
const rejectedDuration = new Trend('budget_rejected_duration', true);
const expectedStatuses = http.expectedStatuses({ min: 200, max: 299 }, 402);

export const options = {
  setupTimeout: timeout,
  teardownTimeout: __ENV.TEARDOWN_TIMEOUT || '1m',
  scenarios: {
    budget_burst: {
      executor: 'shared-iterations',
      exec: 'budgetBurst',
      vus: burstVUs,
      iterations: burstRequests,
      maxDuration: __ENV.BURST_MAX_DURATION || '2m',
    },
  },
  thresholds: {
    checks: ['rate==1'],
    http_req_failed: ['rate==0'],
    'budget_allowed{phase:burst}': [`count>=${minAllowed}`],
    'budget_rejected{phase:burst}': [`count>=${minRejected}`],
    'budget_successful_cost_irr{phase:burst}': [`count<=${budgetLimitIRR + costToleranceIRR}`],
    'budget_saturation_rejected{phase:saturation}': [`count==${saturationRequests}`],
    budget_unexpected_statuses: ['count==0'],
    budget_contract_violations: ['count==0'],
    budget_rate_limit_interference: ['count==0'],
  },
};

function positiveInteger(name, value) {
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error(`${name} must be a positive integer; received ${value}`);
  }
  return parsed;
}

function positiveNumber(name, value) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    throw new Error(`${name} must be a positive number; received ${value}`);
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

function parseMetadata(raw) {
  if (!raw) return undefined;
  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (error) {
    throw new Error(`METADATA_JSON must be valid JSON: ${error.message}`);
  }
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error('METADATA_JSON must contain a JSON object');
  }
  return parsed;
}

function header(response, name) {
  const expected = name.toLowerCase();
  for (const key of Object.keys(response.headers)) {
    if (key.toLowerCase() === expected) return response.headers[key];
  }
  return '';
}

function errorType(response) {
  try {
    return response.json('error.type') || '';
  } catch (_) {
    return '';
  }
}

function budgetRequest(phase, iteration, runID) {
  const body = {
    model,
    messages: [
      { role: 'system', content: 'Return only the word ok.' },
      { role: 'user', content: `Budget load test ${ruleID}-${runID}-${phase}-${iteration}` },
    ],
    temperature: 0,
    max_tokens: maxTokens,
  };
  if (metadata) body.metadata = metadata;

  return http.post(`${baseUrl}/v1/chat/completions`, JSON.stringify(body), {
    headers: {
      Authorization: `Bearer ${apiKey}`,
      'Content-Type': 'application/json',
      'x-godarai-tenant': tenant,
    },
    timeout,
    tags: { endpoint: 'chat_completions', phase, tenant, budget_rule: ruleID },
    responseCallback: expectedStatuses,
  });
}

function classify(response, phase) {
  const isAllowed = response.status >= 200 && response.status < 300;
  const isRejected = response.status === 402;
  const isRateLimited = response.status === 429;
  const expected = isAllowed || isRejected;
  let violations = 0;

  allowed.add(isAllowed ? 1 : 0, { phase });
  rejected.add(isRejected ? 1 : 0, { phase });
  unexpectedStatuses.add(expected ? 0 : 1, { phase });
  rateLimitInterference.add(isRateLimited ? 1 : 0, { phase });

  if (isAllowed) {
    const costIRR = Number(header(response, 'X-Godarai-Cost-IRR'));
    successfulCostIRR.add(Number.isFinite(costIRR) ? costIRR : 0, { phase });
    allowedDuration.add(response.timings.duration, { phase });
    if (!Number.isFinite(costIRR) || costIRR <= 0) violations += 1;
  } else {
    successfulCostIRR.add(0, { phase });
  }

  if (isRejected) {
    rejectedDuration.add(response.timings.duration, { phase });
    if (errorType(response) !== 'budget_exceeded') violations += 1;
  }

  contractViolations.add(violations, { phase });
  check(response, {
    [`${phase}: status is 2xx or budget 402`]: () => expected,
    [`${phase}: response matches budget contract`]: () => violations === 0,
  });

  return { isAllowed, isRejected };
}

export function setup() {
  const response = http.get(`${baseUrl}/health`, {
    timeout: '5s',
    tags: { endpoint: 'health', phase: 'setup' },
  });
  if (response.status !== 200) {
    fail(`Gateway health preflight failed with HTTP ${response.status}: ${response.body}`);
  }
  return { runID: `${Date.now()}` };
}

export function budgetBurst(data) {
  const response = budgetRequest('burst', exec.scenario.iterationInTest, data.runID);
  classify(response, 'burst');
}

export function teardown(data) {
  for (let i = 0; i < saturationRequests; i += 1) {
    const response = budgetRequest('saturation', i, data.runID);
    const result = classify(response, 'saturation');
    saturationRejected.add(result.isRejected ? 1 : 0, { phase: 'saturation' });
    check(response, {
      'saturation: exhausted budget remains blocked': () => result.isRejected,
    });
    if (verificationSleep > 0) sleep(verificationSleep);
  }
}
