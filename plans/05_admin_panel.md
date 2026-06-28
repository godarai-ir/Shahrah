# GodarAI Admin Panel - Support & Management System

## Overview

This document describes the GodarAI Admin Panel, a separate management system for Shahrah's internal support and operations team. This panel is built with Next.js + TailwindCSS + shadcn/ui and maintains its own independent database, separate from the main GodarAI service database.

## 1. Architecture Overview

### 1.1 System Separation

```mermaid
graph TB
    subgraph "GodarAI Main Service"
        A[Tenant API] --> B[(GodarAI DB)]
        C[Tenant Admin Panel] --> B
    end
    
    subgraph "GodarAI Admin Panel"
        D[Support Dashboard] --> E[(Admin DB)]
        F[Ticket Management] --> E
        G[Customer Management] --> E
        H[Billing Overview] --> E
        I[Analytics] --> E
    end
    
    B -. Read-Only Sync .-> E
    
    style A fill:#e1f5fe
    style D fill:#fff3e0
```

### 1.2 Key Design Principles

1. **Independent Database**: Admin panel has its own PostgreSQL database
2. **Read-Only Sync**: Data from GodarAI main service is synced read-only
3. **Separate Deployment**: Can be deployed, scaled, and updated independently
4. **Separate Authentication**: Admin users have their own auth system
5. **Support-First Design**: Optimized for support team workflows

## 2. Technology Stack

### 2.1 Frontend

| Technology | Purpose | Version |
|------------|---------|---------|
| Next.js | React framework | 16.x (App Router) |
| TailwindCSS | Utility-first CSS | 4.x |
| shadcn/ui | UI Components | Latest |
| React Query | Data fetching | 5.x |
| Zustand | State management | 5.x |
| React Hook Form | Form handling | 7.x |
| Zod | Schema validation | 3.x |

### 2.2 Backend

| Technology | Purpose |
|------------|---------|
| Next.js API Routes | Backend endpoints |
| Prisma ORM | Database ORM |
| PostgreSQL | Database |
| Redis | Caching & sessions |

### 2.3 Project Structure

```
shahrah-admin/
├── src/
│   ├── app/                          # Next.js App Router
│   │   ├── (auth)/                   # Auth routes (login, etc.)
│   │   │   ├── login/
│   │   │   └── layout.tsx
│   │   ├── (dashboard)/              # Main dashboard
│   │   │   ├── tickets/
│   │   │   │   ├── page.tsx
│   │   │   │   ├── [id]/
│   │   │   │   └── components/
│   │   │   ├── customers/
│   │   │   │   ├── page.tsx
│   │   │   │   ├── [id]/
│   │   │   │   └── components/
│   │   │   ├── billing/
│   │   │   ├── analytics/
│   │   │   └── layout.tsx
│   │   ├── api/                      # API routes
│   │   │   ├── trpc/
│   │   │   ├── tickets/
│   │   │   └── customers/
│   │   └── layout.tsx
│   ├── components/                   # Shared components
│   │   ├── ui/                       # shadcn/ui components
│   │   ├── tickets/
│   │   ├── customers/
│   │   └── layout/
│   ├── lib/                          # Utilities
│   │   ├── db.ts
│   │   ├── auth.ts
│   │   └── utils.ts
│   ├── hooks/                        # React hooks
│   ├── services/                     # Business logic
│   │   ├── ticket-service.ts
│   │   ├── customer-service.ts
│   │   └── sync-service.ts
│   └── types/                        # TypeScript types
│       └── index.ts
├── prisma/
│   └── schema.prisma
├── public/
├── .env.local
├── next.config.ts
├── tailwind.config.ts
└── package.json
```

## 3. Database Schema (Admin DB)

### 3.1 Core Tables

```prisma
// Admin Panel Database Schema

model AdminUser {
    id              String   @id @default(uuid())
    email           String   @unique
    name            String
    role            AdminRole @default(SUPPORT_AGENT)
    
    // Auth
    password_hash   String
    mfa_enabled     Boolean @default(false)
    mfa_secret      String?
    
    // Status
    is_active       Boolean @default(true)
    last_login_at   DateTime?
    
    created_at      DateTime @default(now())
    updated_at      DateTime @default(now()) @updatedAt
    
    // Relations
    assigned_tickets SupportTicket[]
    
    @@index([email])
    @@index([is_active])
}

enum AdminRole {
    SUPER_ADMIN
    SUPPORT_MANAGER
    SUPPORT_AGENT
    BILLING_AGENT
    VIEWER
}

model SupportTicket {
    id              String   @id @default(uuid())
    
    // Ticket info
    ticket_number   String   @unique  // e.g., "SHR-2024-001234"
    subject         String
    status          TicketStatus @default(OPEN)
    priority        TicketPriority @default(MEDIUM)
    category        TicketCategory
    
    // Customer reference (from GodarAI main service)
    tenant_id       String?
    tenant_slug     String?
    customer_email  String
    customer_name   String?
    
    // Assignment
    assigned_to_id  String?
    
    // Content
    messages        SupportTicketMessage[]
    
    // Metadata
    tags            String[]
    source          TicketSource @default(EMAIL)
    
    // Timing
    first_response_at DateTime?
    resolved_at     DateTime?
    sla_deadline    DateTime?
    
    created_at      DateTime @default(now())
    updated_at      DateTime @default(now()) @updatedAt
    
    assigned_to     AdminUser?    @relation(fields: [assigned_to_id], references: [id])
    
    @@index([status])
    @@index([assigned_to_id])
    @@index([tenant_id])
    @@index([created_at])
}

enum TicketStatus {
    OPEN
    IN_PROGRESS
    WAITING_CUSTOMER
    RESOLVED
    CLOSED
}

enum TicketPriority {
    LOW
    MEDIUM
    HIGH
    CRITICAL
}

enum TicketCategory {
    TECHNICAL_SUPPORT
    BILLING_QUESTION
    FEATURE_REQUEST
    BUG_REPORT
    ACCOUNT_ISSUE
    API_INTEGRATION_HELP
    OTHER
}

enum TicketSource {
    EMAIL
    DASHBOARD
    API
    PHONE
    CHAT
}

model SupportTicketMessage {
    id              String   @id @default(uuid())
    ticket_id       String
    
    // Message content
    content         String
    content_html    String?  // Rendered HTML
    
    // Author
    author_type     MessageAuthorType
    author_id       String?
    author_name     String
    author_email    String
    
    // Attachments
    attachments     Json?    @default("[]")
    
    // Internal note
    is_internal     Boolean @default(false)
    
    created_at      DateTime @default(now())
    
    ticket          SupportTicket @relation(fields: [ticket_id], references: [id])
    
    @@index([ticket_id])
    @@index([created_at])
}

enum MessageAuthorType {
    CUSTOMER
    SUPPORT_AGENT
    SYSTEM
}
```

### 3.2 Customer Sync Table

```prisma
model ShahrahCustomerSync {
    id              String   @id @default(uuid())
    
    // Reference to GodarAI main service
    tenant_id       String   @unique
    tenant_slug     String
    tenant_name     String
    tenant_type     String
    
    // Synced data snapshot
    status          String
    wallet_balance_irr Float?
    subscription_status String?
    
    // Contact info
    contact_emails  String[]
    contact_phones  String[]
    
    // Sync metadata
    last_synced_at  DateTime @default(now())
    sync_version    Int @default(1)
    
    created_at      DateTime @default(now())
    updated_at      DateTime @default(now()) @updatedAt
    
    @@index([tenant_slug])
    @@index([status])
}
```

## 4. Key Features

### 4.1 Ticket Management

```mermaid
graph LR
    A[New Ticket] --> B{Triage}
    B --> C[Auto-Assign by Category]
    B --> D[Manual Assign]
    C --> E[Agent Queue]
    D --> E
    E --> F[Agent Responds]
    F --> G{Resolved?}
    G -->|No| F
    G -->|Yes| H[Mark Resolved]
    H --> I[Customer Confirmation]
    I --> J{Satisfied?}
    J -->|No| E
    J -->|Yes| K[Close Ticket]
```

### 4.2 Ticket List View

```typescript
// src/app/(dashboard)/tickets/page.tsx
import { DataTable } from "@/components/ui/data-table"
import { columns } from "./columns"
import { fetchTickets } from "@/services/ticket-service"

export default async function TicketsPage({
  searchParams,
}: {
  searchParams: {
    status?: string
    priority?: string
    assigned_to?: string
    page?: string
  }
}) {
  const tickets = await fetchTickets({
    status: searchParams.status,
    priority: searchParams.priority,
    assignedTo: searchParams.assigned_to,
    page: parseInt(searchParams.page || "1"),
  })

  return (
    <div className="container mx-auto py-10">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Support Tickets</h1>
        <Badge variant="outline">
          {tickets.totalCount} total
        </Badge>
      </div>
      <DataTable 
        columns={columns} 
        data={tickets.items}
        pageCount={tickets.pageCount}
      />
    </div>
  )
}
```

### 4.3 Ticket Detail View

```typescript
// src/app/(dashboard)/tickets/[id]/page.tsx
export default function TicketDetailPage({ params }: Props) {
  const ticket = useTicket(params.id)
  const messages = useTicketMessages(params.id)
  
  return (
    <div className="grid grid-cols-3 gap-6">
      {/* Ticket Info Sidebar */}
      <Card className="col-span-1">
        <CardHeader>
          <CardTitle>Ticket Details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <Label>Status</Label>
              <TicketStatusBadge status={ticket.status} />
            </div>
            <div>
              <Label>Priority</Label>
              <TicketPriorityBadge priority={ticket.priority} />
            </div>
            <div>
              <Label>Assigned To</Label>
              <AssigneeSelect 
                value={ticket.assignedTo}
                onChange={handleAssign}
              />
            </div>
            <div>
              <Label>Customer</Label>
              <CustomerInfo 
                email={ticket.customerEmail}
                tenantSlug={ticket.tenantSlug}
              />
            </div>
          </div>
        </CardContent>
      </Card>
      
      {/* Message Thread */}
      <div className="col-span-2">
        <TicketMessages messages={messages} />
        <ReplyForm ticketId={params.id} />
      </div>
    </div>
  )
}
```

### 4.4 Customer 360 View

```typescript
// Customer detail page showing all relevant info
interface Customer360 {
  tenant: ShahrahCustomerSync
  tickets: SupportTicket[]
  usage: {
    totalRequests: number
    totalTokens: number
    totalSpendUSD: number
    topModels: ModelUsage[]
  }
  billing: {
    walletBalanceIRR: number
    subscription: SubscriptionInfo
    transactions: Transaction[]
  }
  health: {
    errorRate: number
    avgLatency: number
    lastActiveAt: Date
  }
}
```

## 5. Data Sync from GodarAI Main Service

### 5.1 Sync Architecture

```mermaid
sequenceDiagram
    participant S as GodarAI Main DB
    participant W as Sync Worker
    participant Q as Message Queue
    participant A as Admin DB
    
    S --> W: Change Data Capture (CDC)
    W --> Q: Publish customer update
    Q --> W: Consume update
    W --> A: Upsert ShahrahCustomerSync
    
    Note over S,A: Sync happens in near real-time
```

### 5.2 Sync Implementation

```typescript
// src/services/sync-service.ts
import { db } from "@/lib/db"
import { shahrahCustomerSync } from "@/lib/db/schema"

interface SyncCustomerPayload {
  tenant_id: string
  tenant_slug: string
  tenant_name: string
  tenant_type: string
  status: string
  wallet_balance_irr: number | null
  subscription_status: string | null
  contact_emails: string[]
}

export async function syncCustomer(payload: SyncCustomerPayload) {
  await db
    .insert(shahrahCustomerSync)
    .values({
      tenantId: payload.tenant_id,
      tenantSlug: payload.tenant_slug,
      tenantName: payload.tenant_name,
      tenantType: payload.tenant_type,
      status: payload.status,
      walletBalanceIrr: payload.wallet_balance_irr,
      subscriptionStatus: payload.subscription_status,
      contactEmails: payload.contact_emails,
      lastSyncedAt: new Date(),
    })
    .onConflictDoUpdate({
      target: shahrahCustomerSync.tenantId,
      set: {
        tenantSlug: payload.tenant_slug,
        tenantName: payload.tenant_name,
        status: payload.status,
        walletBalanceIrr: payload.wallet_balance_irr,
        subscriptionStatus: payload.subscription_status,
        contactEmails: payload.contact_emails,
        lastSyncedAt: new Date(),
        syncVersion: sql`${shahrahCustomerSync.syncVersion} + 1`,
      },
    })
}
```

### 5.3 Sync Methods

| Method | Description | Latency |
|--------|-------------|---------|
| CDC (Debezium) | Capture database changes | < 1s |
| Webhook | GodarAI pushes updates | < 5s |
| Scheduled Poll | Admin pulls updates periodically | 1-5 min |
| Manual Sync | Trigger from admin UI | On-demand |

## 6. Authentication & Authorization

### 6.1 Admin Auth Flow

```mermaid
sequenceDiagram
    participant U as Admin User
    participant A as Admin Panel
    participant D as Admin DB
    
    U->>A: Login (email + password)
    A->>D: Verify credentials
    D->>A: Return user + role
    A->>U: Set session cookie + redirect
    
    U->>A: Access resource
    A->>A: Check role permissions
    A->>U: Allow/Deny
```

### 6.2 Role-Based Access Control

```typescript
// src/lib/auth.ts
export const rolePermissions = {
  SUPER_ADMIN: [
    "manage_users",
    "manage_tickets",
    "view_billing",
    "view_analytics",
    "manage_settings",
  ],
  SUPPORT_MANAGER: [
    "manage_tickets",
    "assign_tickets",
    "view_billing",
    "view_analytics",
  ],
  SUPPORT_AGENT: [
    "view_tickets",
    "update_tickets",
    "respond_tickets",
  ],
  BILLING_AGENT: [
    "view_billing",
    "update_billing",
    "view_tickets",
  ],
  VIEWER: [
    "view_tickets",
    "view_billing",
    "view_analytics",
  ],
} as const
```

### 6.3 Session Management

```typescript
// Using next-auth or custom session
import { getIronSession } from "iron-session"

const sessionOptions = {
  password: process.env.SESSION_PASSWORD!,
  cookieName: "shahrah-admin-session",
  cookieOptions: {
    secure: process.env.NODE_ENV === "production",
    maxAge: 8 * 60 * 60, // 8 hours
  },
}

export async function getSession(req: NextRequest) {
  return getIronSession<AdminSession>(req, sessionOptions)
}

interface AdminSession {
  userId: string
  email: string
  role: AdminRole
  permissions: string[]
}
```

## 7. UI Components

### 7.1 Ticket Status Badge

```typescript
// src/components/tickets/ticket-status-badge.tsx
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

const statusStyles = {
  open: "bg-blue-100 text-blue-800",
  in_progress: "bg-yellow-100 text-yellow-800",
  waiting_customer: "bg-orange-100 text-orange-800",
  resolved: "bg-green-100 text-green-800",
  closed: "bg-gray-100 text-gray-800",
}

export function TicketStatusBadge({ status }: { status: TicketStatus }) {
  return (
    <Badge variant="secondary" className={cn(statusStyles[status])}>
      {status.replace("_", " ").toUpperCase()}
    </Badge>
  )
}
```

### 7.2 Priority Indicator

```typescript
// src/components/tickets/priority-indicator.tsx
import { AlertTriangle, ArrowUp, Minus, ArrowDown } from "lucide-react"

const priorityConfig = {
  critical: { icon: AlertTriangle, color: "text-red-600", label: "Critical" },
  high: { icon: ArrowUp, color: "text-orange-600", label: "High" },
  medium: { icon: Minus, color: "text-yellow-600", label: "Medium" },
  low: { icon: ArrowDown, color: "text-green-600", label: "Low" },
}

export function PriorityIndicator({ priority }: { priority: TicketPriority }) {
  const config = priorityConfig[priority]
  const Icon = config.icon
  
  return (
    <span className={cn("flex items-center gap-1", config.color)}>
      <Icon className="h-4 w-4" />
      {config.label}
    </span>
  )
}
```

## 8. API Routes

### 8.1 Ticket API

```typescript
// src/app/api/tickets/route.ts
export async function GET(request: NextRequest) {
  const session = await getSession(request)
  if (!hasPermission(session, "view_tickets")) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 403 })
  }
  
  const { searchParams } = new URL(request.url)
  const tickets = await fetchTickets({
    status: searchParams.get("status"),
    priority: searchParams.get("priority"),
    assignedTo: searchParams.get("assigned_to"),
    page: parseInt(searchParams.get("page") || "1"),
    limit: parseInt(searchParams.get("limit") || "20"),
  })
  
  return NextResponse.json(tickets)
}

export async function POST(request: NextRequest) {
  const session = await getSession(request)
  if (!hasPermission(session, "manage_tickets")) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 403 })
  }
  
  const body = await request.json()
  const ticket = await createTicket({
    ...body,
    createdBy: session.userId,
  })
  
  return NextResponse.json(ticket, { status: 201 })
}
```

## 9. Deployment

### 9.1 Docker Configuration

```dockerfile
# Dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine AS runner
WORKDIR /app
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3000
ENV PORT=3000
CMD ["node", "server.js"]
```

### 9.2 Docker Compose

```yaml
# docker-compose.admin.yml
version: '3.8'

services:
  admin-panel:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgresql://admin:password@admin-db:5432/shahrah_admin
      - REDIS_URL=redis://admin-redis:6379
      - SESSION_PASSWORD=${SESSION_PASSWORD}
    depends_on:
      - admin-db
      - admin-redis
  
  admin-db:
    image: postgres:16
    environment:
      - POSTGRES_USER=admin
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=shahrah_admin
    volumes:
      - admin-db-data:/var/lib/postgresql/data
  
  admin-redis:
    image: redis:7-alpine

volumes:
  admin-db-data:
```

## 10. Monitoring & Alerts

### 10.1 Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| Open Tickets | Unassigned tickets | > 10 |
| Avg Response Time | Time to first response | > 4 hours |
| SLA Breaches | Tickets past deadline | > 5 |
| Ticket Backlog | Total open + in_progress | > 100 |

### 10.2 Dashboard Widgets

```typescript
// Dashboard overview components
interface DashboardMetrics {
  tickets: {
    open: number
    inProgress: number
    resolvedToday: number
    avgResponseTime: string
  }
  customers: {
    active: number
    newThisWeek: number
    atRisk: number
  }
  billing: {
    totalRevenueIRR: number
    outstandingInvoices: number
  }
  system: {
    apiHealth: string
    avgLatency: string
  }
}
```
