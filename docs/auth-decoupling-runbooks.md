# Auth Decoupling - Deployment, Security & Monitoring Runbooks

## Table of Contents
1. [Deployment RunBOOK](#deployment-runbook)
2. [Security Validation Runbook](#security-validation-runbook)
3. [Monitoring & Error Rate Runbook](#monitoring--error-rate-runbook)

---

## Deployment Runbook

### Pre-Deployment Checklist
- [ ] All new TypeScript files have been merged
- [ ] Go control-plane auth handlers are deployed
- [ ] Go control-plane routes are registered (`/api/v1/auth/*`)
- [ ] PostgreSQL tables exist: `tenant_admin_captchas`, `tenant_admin_sessions`, `tenant_admin_otp_requests`
- [ ] Database migration script has been executed in staging/production
- [ ] `SHAHRAH_MASTER_KEY` is configured in control-plane environment
- [ ] Network connectivity from tenant-admin to control-plane is verified

### Staging Deployment
```bash
# 1. Set environment variables
export USE_API_AUTH=true
export CONTROL_PLANE_URL=https://control-plane.staging.internal:8443
export SHAHRAH_MASTER_KEY=<staging-master-key>

# 2. Deploy tenant-admin
cd apps/tenant-admin
pnpm build
# Deploy via your CI/CD pipeline

# 3. Verify health endpoint
curl https://tenant-admin-staging.shahrah.local/health

# 4. Run smoke tests
# See security validation runbook below
```

### Production Deployment (Blue/Green)
```bash
# 1. Deploy green environment with USE_API_AUTH=false first
# This ensures code compatibility without switching backends

# 2. Verify green environment works with Prisma mode

# 3. Switch traffic to green environment

# 4. After 24h monitoring, enable API mode on green
export USE_API_AUTH=true

# 5. Monitor for 1 hour, then rollback if needed
export USE_API_AUTH=false
```

### Rollback Procedure
```bash
# Immediate rollback
export USE_API_AUTH=false
# Restart the application

# If code rollback is needed
git revert HEAD~5..HEAD  # Revert the auth decoupling commits
```

---

## Security Validation Runbook

### 1. CAPTCHA Security
- [ ] CAPTCHA challenges expire within 2 minutes
- [ ] CAPTCHA can only be used once (consumed)
- [ ] Answer comparison is case-insensitive
- [ ] Expired CAPTCHAs are cleaned up automatically

```bash
# Test CAPTCHA lifecycle
curl -X GET https://tenant-admin.shahrah.local/api/auth/captcha
# Verify response has captcha_id, image_data_url, expires_in_seconds <= 120
```

### 2. OTP Security
- [ ] OTP codes are 6 digits
- [ ] OTP codes expire within 5 minutes
- [ ] Maximum 5 attempts per OTP request
- [ ] OTP requests are rate-limited (per identifier, per tenant)
- [ ] OTP codes are never logged in production

```bash
# Test OTP expiration
curl -X POST https://tenant-admin.shahrah.local/api/auth/request-otp \
  -H "Content-Type: application/json" \
  -d '{"identifier":"test@example.com","captcha_id":"...","captcha_answer":"..."}'

# Verify rate limiting (send 10 requests rapidly)
for i in {1..10}; do
  curl -X POST ...  # Same payload
  sleep 0.1
done
```

### 3. Session Security
- [ ] Session IDs are cryptographically random (48 hex characters)
- [ ] Sessions expire after 7 days
- [ ] Revoked sessions are immediately invalidated
- [ ] Session cookies are HttpOnly, Secure, SameSite=Lax

```bash
# Test session cookie flags
curl -i -X POST https://tenant-admin.shahrah.local/api/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"request_id":"...","otp_code":"..."}'
# Verify Set-Cookie header includes: HttpOnly, Secure, SameSite=Lax
```

### 4. API Authentication (Go Backend)
- [ ] All auth endpoints require valid master key
- [ ] Invalid API keys are rejected with 401
- [ ] Expired API keys are rejected
- [ ] Rate limiting is applied to auth endpoints

```bash
# Test auth without API key
curl -X POST https://control-plane.internal:8443/api/v1/auth/captcha/create
# Expected: 401 Unauthorized

# Test auth with invalid API key
curl -X POST https://control-plane.internal:8443/api/v1/auth/captcha/create \
  -H "Authorization: Bearer invalid-key"
# Expected: 401 Unauthorized
```

### 5. Tenant Isolation
- [ ] OTP requests for wrong tenant are rejected
- [ ] Session tenant is used as source of truth (not header hint)
- [ ] Users cannot access data from other tenants

```bash
# Test tenant isolation
# Request OTP for tenant A, then verify with tenant B header
# Expected: "این درخواست OTP مربوط به Tenant فعلی نیست."
```

---

## Monitoring & Error Rate Runbook

### Key Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `auth_otp_request_rate` | OTP requests per minute | > 100/min per tenant |
| `auth_otp_verify_success_rate` | OTP verification success rate | < 95% |
| `auth_session_create_rate` | Session creations per minute | > 50/min |
| `auth_captcha_solve_rate` | CAPTCHA solve success rate | < 80% |
| `auth_api_error_rate` | HTTP 5xx errors from auth endpoints | > 1% |
| `auth_api_latency_p99` | 99th percentile latency for auth API | > 2000ms |
| `auth_api_latency_p50` | 50th percentile latency for auth API | > 500ms |

### Grafana Dashboard Configuration

```json
{
  "dashboard": {
    "title": "Tenant Admin Auth - Decoupled",
    "panels": [
      {
        "title": "OTP Success Rate",
        "targets": [{"expr": "rate(auth_otp_verify_success_total[5m]) / rate(auth_otp_verify_total[5m])"}]
      },
      {
        "title": "Auth API Latency P99",
        "targets": [{"expr": "histogram_quantile(0.99, rate(auth_api_request_duration_seconds_bucket[5m]))"}]
      },
      {
        "title": "Auth Error Rate by Endpoint",
        "targets": [{"expr": "rate(auth_api_errors_total{status=~\"5..\"}[5m]) by (endpoint)"}]
      }
    ]
  }
}
```

### Alerting Rules (Prometheus)

```yaml
groups:
  - name: tenant-admin-auth
    rules:
      - alert: HighAuthErrorRate
        expr: rate(auth_api_errors_total[5m]) / rate(auth_api_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High auth error rate detected"
          description: "Auth error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

      - alert: HighAuthLatency
        expr: histogram_quantile(0.99, rate(auth_api_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High auth API latency detected"
          description: "P99 latency is {{ $value }}s for the last 5 minutes"

      - alert: OTPRateLimitExceeded
        expr: rate(auth_otp_request_total[1m]) > 100
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "High OTP request rate detected"
          description: "OTP request rate is {{ $value }}/min"
```

### Incident Response

**Scenario: Users cannot login**
1. Check `auth_api_error_rate` metric
2. If API mode: check control-plane health
3. If Prisma mode: check database connectivity
4. Check logs for specific error messages:
   ```bash
   kubectl logs -l app=tenant-admin --tail=100 | grep -i "auth\|otp\|session"
   ```
5. If API mode issues: switch to Prisma mode (`USE_API_AUTH=false`)

**Scenario: High auth latency**
1. Check `auth_api_latency_p99` metric
2. If API mode: check control-plane load and network latency
3. If Prisma mode: check database query performance
4. Check connection pool utilization

**Scenario: OTP brute force attack**
1. Check `auth_otp_request_rate` metric per tenant
2. Verify rate limiting is active
3. If needed, manually block suspicious IPs
4. Alert security team

### Post-Deployment Validation Checklist
- [ ] All auth metrics are within normal thresholds
- [ ] No 5xx errors in auth endpoints for the last 30 minutes
- [ ] Session creation rate matches expected user sign-in rate
- [ ] OTP verification success rate > 95%
- [ ] No alerts fired in the last 1 hour
- [ ] Error logs do not contain unexpected patterns
