# GodarAI Framework - Documentation Index

## Overview

This directory contains the foundational documentation for **Shahrah** (شاهراه), a unified LLM gateway framework designed for the Iranian market. GodarAI includes Rial wallet payments, controlled European provider egress, and multi-tenant subdomain isolation.

## Document Index

| # | Document | File | Status |
|---|----------|------|--------|
| 1 | Iranian Payment System (Wallet & Billing) | [`01_iranian_payment_system.md`](01_iranian_payment_system.md) | ✅ Complete |
| 2 | Multi-Tenant Architecture with On-Premise Support | [`02_multi_tenant_architecture.md`](02_multi_tenant_architecture.md) | ✅ Complete |
| 3 | Subdomain-Based Tenant Access | [`03_subdomain_tenant_access.md`](03_subdomain_tenant_access.md) | ✅ Complete |
| 5 | Admin Panel for GodarAI Management | [`05_admin_panel.md`](05_admin_panel.md) | ✅ Complete |
| 6 | Docusaurus Documentation System | [`06_docusaurus_documentation.md`](06_docusaurus_documentation.md) | ✅ Complete |
| 7 | Blog Service (Future) | [`07_blog_service.md`](07_blog_service.md) | ✅ Complete |
| 8 | SDK Development Strategy | [`08_sdk_development_strategy.md`](08_sdk_development_strategy.md) | ✅ Complete |
| 9 | GodarAI Feature Set | [`09_GodarAI_features.md`](09_GodarAI_features.md) | ✅ Complete |
| 10 | Performance & Availability Architecture | [`10_performance_availability.md`](10_performance_availability.md) | ✅ Complete |

| 11 | Observability Platform Implementation Plan | [`observability-platform-implementation-plan.md`](observability-platform-implementation-plan.md) | âœ… Complete |

## Architecture Summary

```mermaid
graph TB
    subgraph "User Access Layer"
        A["*.sharah.ir - API"] --> B[API Gateway - Iran]
        C["*.app.sharah.ir - Admin"] --> D[Tenant Admin Panel]
    end
    
    subgraph "Iran Control Plane"
        B --> E[Tenant Resolver]
        E --> F[Auth Service]
        F --> G[Rate Limiter]
        G --> H[Billing Service]
        H --> I[Silk Road]
        
        D --> J[Support Panel]
        J --> K[(Admin DB)]
    end
    
    subgraph "Provider Egress"
        I -. HTTP CONNECT .-> L[EU Egress Proxy]
    end
    
    subgraph "Europe Services (Silk Road)"
        L --> M[Silk Road Proxy]
        M --> N[OpenAI]
        M --> O[Anthropic]
        M --> P[Other LLMs]
    end
    
    subgraph "Supporting Services"
        Q[(GodarAI DB)] -. sync .-> K
        R[Redis Cache]
        S[Docusaurus Docs]
        T[Blog Service]
    end
    
    H --> Q
    F --> R
    G --> R
```

## Key Design Decisions

### 1. Payment System
- **Rial Wallet**: Customers top up in IRR through Iranian payment gateways
- **USD Token Pricing**: Token consumption priced in USD, converted at real-time rates
- **Exchange Rate API**: Real-time rate fetching from external providers
- **Tax Handling**: 9% VAT on subscription fees

### 2. Multi-Tenancy
- **Row-Level Isolation**: Shared database with `tenant_id` filtering
- **On-Premise Ready**: Architecture designed for future extraction to dedicated deployments
- **Subdomain Access**: `tenant.sharah.ir` for API, `tenant.app.sharah.ir` for admin

### 3. Controlled European Provider Egress
- **Direct Internal Dispatch**: Control Plane sends provider operations directly to Silk Road
- **Stateless Silk Road**: Provider adapters and request forwarding run without persistent application state
- **Restricted Egress**: Provider traffic exits through the explicitly configured European proxy path

### 4. Technology Stack
- **Backend**: Go (chi router, net/http) for both GodarAI Control Plane and Silk Road
- **Frontend**: Next.js + TailwindCSS + shadcn/ui for admin panels
- **Database**: PostgreSQL with Prisma ORM (Python) / GORM (Go)
- **Cache**: Redis Cluster with Sentinel (go-redis)
- **Provider Egress**: HTTP dispatch to Silk Road with an explicit HTTP proxy for European egress
- **SDKs**: Turborepo monorepo with TypeScript, Python, Next.js, NestJS packages

## Performance Targets

| Metric | Target |
|--------|--------|
| P95 Control Plane Latency | < 10ms |
| Availability | 99.99% |
| Throughput (single instance) | 10,000+ req/sec |

## Next Steps

1. **Review & Approve**: Review these documents and provide feedback
2. **Switch to Code Mode**: Begin implementation starting with:
   - Core project structure and monorepo setup
   - Database schema (Prisma)
   - Tenant resolution middleware
   - Provider-egress transport and proxy policy
   - Payment gateway integration
3. **Iterative Development**: Build in phases, starting with MVP features

## Recent Delivery Notes

- **2026-04-28**: Wallet top-up intent/callback/verify flow delivered in `services/control-plane` with provider adapters for `mock`, `zarinpal`, and `idpay`, plus idempotent wallet crediting and transaction auditing.

## Active Provider Readiness Plans

- [xAI four-model full-support execution plan](xai-four-model-production-readiness.md): production-readiness work for `grok-4.3`, `grok-4.5`, `grok-imagine-image`, and `grok-imagine-image-quality`.
- [Moonshot four-model production-readiness plan](moonshot-four-model-production-readiness.md): guarded delivery and release gates for `kimi-k3`, `kimi-k2.7-code`, `kimi-k2.7-code-highspeed`, and `kimi-k2.6`.
