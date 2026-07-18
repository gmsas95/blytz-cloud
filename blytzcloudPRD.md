---
type: guide
title: Blytz Personal AI Assistant Platform - Complete PRD & Documentation
resource: blytz-cloud
description: "1. [Executive Summary](#1-executive-summary)"
tags: [docker, go, stripe]
updated: 2026-06-18
---

# Blytz Personal AI Assistant Platform - Complete PRD & Documentation

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Product Specification](#2-product-specification)
3. [Technical Architecture](#3-technical-architecture)
4. [API Specification](#4-api-specification)
5. [Database Schema](#5-database-schema)
6. [File Specifications](#6-file-specifications)
7. [Implementation Sprint](#7-implementation-sprint)
8. [Deployment Guide](#8-deployment-guide)
9. [Testing Strategy](#9-testing-strategy)
10. [Success Metrics](#10-success-metrics)
11. [Risk Mitigation](#11-risk-mitigation)
12. [Future Roadmap](#12-future-roadmap)

---

## 1. Executive Summary

### Vision

BlytzCloud is a platform that automatically deploys personalized OpenClaw AI assistants for freelancers and contractors. Users sign up, describe what they need help with, provide a Telegram bot token, pay $29/month, and get a working AI assistant within 2 minutes.

### Core Principle

Automate the provisioning, eliminate the complexity. Users never touch a terminal, never edit config files, never manage servers.

### Target Market

| Segment | Description |
|---------|-------------|
| Primary | Freelancers overwhelmed with administrative work |
| Secondary | Contractors needing proposal/research help |
| Tertiary | Solo entrepreneurs wanting AI assistance |

### Value Proposition

"Train your AI assistant in 2 minutes. It knows your context from day one. $29/month."

### Competitive Positioning

Unlike generic OpenClaw hosting (SimpleClaw, Majordomo, etc.), Blytz:
- Targets non-technical users (no SSH, no config editing)
- Pre-configures context from onboarding (AGENTS.md, USER.md, SOUL.md)
- Single purpose: personal assistant (not multi-use agent platform)
- Flat pricing, no usage surprises

---

## 2. Product Specification

### 2.1 Scope

#### In Scope (MVP)

| Feature | Description |
|---------|-------------|
| Single template | Personal Assistant |
| Single channel | Telegram (user provides bot token) |
| Single payment | Stripe ($29/month subscription) |
| Hosting | Single server (Ryzen 7, 32GB RAM, max 20 customers) |
| Customization | Freeform text input for instructions |
| Database | SQLite |
| Reverse proxy | Caddy (subdomain per customer) |

#### Out of Scope (Post-MVP)

| Feature | When |
|---------|------|
| WhatsApp integration | After 10 customers |
| Slack integration | After 20 customers |
| Multi-template (Sales, Content, Admin) | After 15 customers |
| Usage-based pricing | Never (keep flat) |
| Kubernetes migration | After 15 customers |
| White-label offering | After 30 customers |

### 2.2 User Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           LANDING PAGE                                       │
│                                                                             │
│   Headline: "Your Personal AI Assistant"                                    │
│   Subhead: "Train it. Deploy it. $29/month."                               │
│                                                                             │
│   [Email Address __________________]  [Get Started →]                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CONFIGURE YOUR ASSISTANT                             │
│                                                                             │
│   Step 1: What should I call you?                                           │
│   [________________________]                                                │
│   (e.g., "Alex", "Mike", "Assistant")                                       │
│                                                                             │
│   Step 2: What do you want help with?                                       │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                                                                     │   │
│   │  I'm a freelance developer. I need help with:                       │   │
│   │  - Drafting proposals for new clients                               │   │
│   │  - Researching competitors and technologies                         │   │
│   │  - Following up on outstanding invoices                             │   │
│   │  - Summarizing long emails                                          │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│   (Be specific. The more detail, the better your assistant will be.)        │
│                                                                             │
│   Step 3: Telegram Bot Token                                                │
│   [________________________________]                                        │
│   (Get one free from @BotFather → /newbot)                                 │
│                                                                             │
│   [Continue to Payment →]                                                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            STRIPE CHECKOUT                                   │
│                                                                             │
│   Blytz Personal AI Assistant                                               │
│   $29.00 / month                                                            │
│                                                                             │
│   [Card Number          ]                                                   │
│   [Expiry    ] [CVC   ]                                                     │
│                                                                             │
│   [Subscribe]                                                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DEPLOYING YOUR ASSISTANT                           │
│                                                                             │
│   ████████████████████░░░░░░░░░░  67%                                       │
│                                                                             │
│   ✓ Payment confirmed                                                       │
│   ✓ Creating your workspace...                                              │
│   ✓ Starting your assistant...                                              │
│   ○ Connecting to Telegram...                                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            YOUR ASSISTANT IS READY!                          │
│                                                                             │
│   🎉 Your AI assistant is live and waiting for your first message.          │
│                                                                             │
│   [Open in Telegram →] t.me/YourBotName                                     │
│                                                                             │
│   Your assistant URL: https://your-email-com.blytz.cloud                    │
│                                                                             │
│   Tips:                                                                     │
│   • Just start chatting - your assistant already knows your context         │
│   • It remembers conversations and learns over time                          │
│   • Cancel anytime from your dashboard                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 Customer States

| State | Description | Actions Allowed |
|-------|-------------|-----------------|
| `pending` | Signed up, awaiting payment | None |
| `provisioning` | Payment received, container starting | Poll status |
| `active` | Container running, assistant live | Full access |
| `suspended` | Payment failed / cancelled | Read-only dashboard |
| `cancelled` | Subscription ended | Data deletion pending |

---

## 3. Technical Architecture

### 3.1 System Overview

```
                                    ┌─────────────────┐
                                    │   Stripe API    │
                                    └────────┬────────┘
                                             │ webhooks
                                             ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              BLYTZ PLATFORM                                  │
│                           (Go + Gin + SQLite)                                │
│                                                                             │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                    │
│   │   API       │    │  Provisioner │    │  Workspace  │                    │
│   │  Handlers   │───▶│   Service    │───▶│  Generator  │                    │
│   └─────────────┘    └──────┬──────┘    └─────────────┘                    │
│                             │                                                │
│                    Docker SDK│                                                │
│                             ▼                                                │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                        DOCKER HOST                                    │  │
│   │                                                                       │  │
│   │   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                │  │
│   │   │ Customer A  │  │ Customer B  │  │ Customer C  │  ... (max 20)  │  │
│   │   │ Port:30001  │  │ Port:30002  │  │ Port:30003  │                │  │
│   │   └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                │  │
│   │          │                │                │                        │  │
│   └──────────┼────────────────┼────────────────┼────────────────────────┘  │
│              │                │                │                            │
└──────────────┼────────────────┼────────────────┼────────────────────────────┘
               │                │                │
               └────────────────┼────────────────┘
                                │
                         ┌──────▼──────┐
                         │    Caddy    │
                         │ Reverse Proxy│
                         └──────┬──────┘
                                │
                    ┌───────────┼───────────┐
                    │           │           │
              ┌─────▼─────┐ ┌───▼───┐ ┌─────▼─────┐
              │customer-a │ │ cust-b│ │customer-c │
              │.blytz.cloud│ │.blytz │ │.blytz.cloud│
              └───────────┘ └───────┘ └───────────┘
```

### 3.2 Server Requirements

| Resource | Specification |
|----------|---------------|
| CPU | Ryzen 7 (8 cores) |
| RAM | 32GB |
| Storage | 500GB SSD |
| OS | Ubuntu 22.04 LTS |
| Network | Static IP, ports 80/443 open |
| Domain | `blytz.cloud` with wildcard DNS |

### 3.3 Directory Structure

```
/opt/blytz/
├── blytz                         # Compiled Go binary
├── config.env                    # Environment configuration
├── platform/
│   ├── database.sqlite           # Customer database
│   └── templates/
│       └── personal-assistant/
│           ├── AGENTS.md.tmpl    # Template for AGENTS.md
│           ├── USER.md.tmpl      # Template for USER.md
│           └── SOUL.md.tmpl      # Template for SOUL.md
├── customers/                    # Customer data (gitignored)
│   └── {customer-id}/
│       ├── .openclaw/
│       │   ├── openclaw.json     # OpenClaw config
│       │   ├── credentials/      # Auth credentials
│       │   └── workspace/
│       │       ├── AGENTS.md     # Generated workspace rules
│       │       ├── USER.md       # Generated user context
│       │       ├── SOUL.md       # Generated agent personality
│       │       └── memory/       # Agent memory files
│       ├── docker-compose.yml    # Generated compose file
│       └── .env                  # Container environment
├── caddy/
│   └── Caddyfile                 # Dynamic subdomain routing
└── logs/
    └── blytz.log                 # Platform logs
```

### 3.4 Go Project Structure

```
blytz/
├── cmd/
│   └── server/
│       └── main.go               # Entry point
├── internal/
│   ├── api/
│   │   ├── handler.go            # HTTP handlers
│   │   ├── middleware.go         # Auth, logging, CORS
│   │   └── routes.go             # Route definitions
│   ├── config/
│   │   └── config.go             # Configuration loading
│   ├── db/
│   │   ├── db.go                 # Database connection
│   │   ├── migrations.go         # Schema migrations
│   │   └── customer.go           # Customer CRUD operations
│   ├── provisioner/
│   │   ├── provisioner.go        # Container lifecycle management
│   │   ├── compose.go            # Docker-compose generation
│   │   └── ports.go              # Port assignment logic
│   ├── workspace/
│   │   └── generator.go          # Generate AGENTS.md, USER.md, SOUL.md
│   ├── telegram/
│   │   └── validate.go           # Validate bot token with Telegram API
│   ├── stripe/
│   │   ├── checkout.go           # Create Stripe checkout session
│   │   └── webhook.go            # Handle Stripe webhooks
│   └── caddy/
│       └── caddy.go              # Caddy Admin API integration
├── static/
│   ├── index.html                # Landing page (embedded)
│   ├── configure.html            # Configuration form (embedded)
│   ├── success.html              # Success page (embedded)
│   └── dashboard.html            # Customer dashboard (embedded)
├── deployments/
│   ├── blytz.service             # systemd unit file
│   └── install.sh                # Installation script
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 3.5 Go Dependencies

```go
// go.mod
module blytz

go 1.26.0

require (
    github.com/gin-gonic/gin v1.11.0
    modernc.org/sqlite v1.46.1
    github.com/docker/docker v29.2.1
    github.com/stripe/stripe-go/v84 v84.3.0
    github.com/google/uuid v1.6.0
    github.com/joho/godotenv v1.5.1
    go.uber.org/zap v1.27.1
)
```

---

## 4. API Specification

### 4.1 Endpoints Overview

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/` | Landing page | Public |
| `GET` | `/configure` | Configuration form | Public |
| `POST` | `/api/signup` | Create customer | Public |
| `GET` | `/api/checkout/:id` | Get Stripe checkout URL | Public |
| `POST` | `/api/webhook/stripe` | Stripe webhook | Stripe signature |
| `GET` | `/api/status/:id` | Get customer status | Public |
| `GET` | `/api/health` | Platform health check | Public |
| `GET` | `/dashboard/:id` | Customer dashboard | Token (future) |

### 4.2 Request/Response Schemas

#### POST /api/signup

**Request:**
```json
{
  "email": "user@example.com",
  "assistant_name": "Alex",
  "custom_instructions": "I'm a freelance developer. Help me with proposals, research, and scheduling.",
  "telegram_bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
}
```

**Response (201 Created):**
```json
{
  "customer_id": "user-example-com",
  "email": "user@example.com",
  "status": "pending",
  "checkout_url": "https://checkout.stripe.com/c/pay/cs_test_..."
}
```

**Response (400 Bad Request):**
```json
{
  "error": "validation_failed",
  "message": "Invalid Telegram bot token",
  "details": {
    "field": "telegram_bot_token",
    "reason": "Token format should be: <numbers>:<alphanumeric>"
  }
}
```

**Response (409 Conflict):**
```json
{
  "error": "already_exists",
  "message": "An account with this email already exists"
}
```

**Response (503 Service Unavailable):**
```json
{
  "error": "at_capacity",
  "message": "Platform is at maximum capacity. Join our waitlist.",
  "waitlist_url": "https://blytz.cloud/waitlist"
}
```

#### GET /api/status/:id

**Response (200 OK):**
```json
{
  "customer_id": "user-example-com",
  "email": "user@example.com",
  "assistant_name": "Alex",
  "status": "active",
  "container_status": "running",
  "telegram_bot_username": "@UserAssistantBot",
  "url": "https://user-example-com.blytz.cloud",
  "telegram_url": "https://t.me/UserAssistantBot",
  "created_at": "2026-02-18T10:30:00Z",
  "paid_at": "2026-02-18T10:31:00Z",
  "subscription_status": "active",
  "current_period_end": "2026-03-18T10:31:00Z"
}
```

#### POST /api/webhook/stripe

**Webhook Events Handled:**

| Event | Action |
|-------|--------|
| `checkout.session.completed` | Provision customer container |
| `customer.subscription.deleted` | Suspend customer container |
| `invoice.payment_failed` | Mark payment failed, notify customer |
| `customer.subscription.updated` | Update subscription status |

### 4.3 Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `validation_failed` | 400 | Request validation failed |
| `invalid_bot_token` | 400 | Telegram bot token is invalid |
| `already_exists` | 409 | Email or customer ID already exists |
| `not_found` | 404 | Customer not found |
| `at_capacity` | 503 | Platform at max capacity (20 customers) |
| `provisioning_failed` | 500 | Container provisioning failed |
| `internal_error` | 500 | Unexpected server error |

---

## 5. Database Schema

### 5.1 Customers Table

```sql
CREATE TABLE customers (
    id TEXT PRIMARY KEY,                    -- Generated from email: user@example.com -> user-example-com
    email TEXT NOT NULL UNIQUE,
    assistant_name TEXT NOT NULL,
    custom_instructions TEXT NOT NULL,
    telegram_bot_token TEXT NOT NULL,
    telegram_bot_username TEXT,             -- Fetched from Telegram API
    container_port INTEGER,                 -- 30000-30999
    container_id TEXT,                      -- Docker container ID
    status TEXT NOT NULL DEFAULT 'pending', -- pending, provisioning, active, suspended, cancelled
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    stripe_checkout_session_id TEXT,
    subscription_status TEXT,               -- active, past_due, cancelled
    current_period_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,
    suspended_at TIMESTAMP,
    cancelled_at TIMESTAMP
);

CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_stripe_customer ON customers(stripe_customer_id);
```

### 5.2 Audit Log Table

```sql
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id TEXT NOT NULL,
    action TEXT NOT NULL,                   -- created, provisioned, suspended, cancelled, etc.
    details TEXT,                           -- JSON payload
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE INDEX idx_audit_customer ON audit_log(customer_id);
CREATE INDEX idx_audit_created ON audit_log(created_at);
```

### 5.3 Port Allocation Table

```sql
CREATE TABLE port_allocations (
    port INTEGER PRIMARY KEY,               -- 30000-30999
    customer_id TEXT NOT NULL,
    allocated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);
```

---

## 6. File Specifications

### 6.1 Docker Compose Template

**File:** `/opt/blytz/customers/{customer-id}/docker-compose.yml`

```yaml
version: '3.8'
services:
  openclaw:
    image: node:22-alpine
    container_name: blytz-{customer-id}
    working_dir: /app
    command: sh -c "npm install -g openclaw@latest && openclaw gateway --port 18789"
    ports:
      - "{port}:18789"
    volumes:
      - ./.openclaw:/root/.openclaw
    environment:
      - OPENCLAW_STATE_DIR=/root/.openclaw
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - OPENCLAW_GATEWAY_TOKEN=${OPENCLAW_GATEWAY_TOKEN}
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:18789/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
```

### 6.2 Customer Environment File

**File:** `/opt/blytz/customers/{customer-id}/.env`

```env
OPENAI_API_KEY=sk-...
TELEGRAM_BOT_TOKEN=123456789:ABCdef...
OPENCLAW_GATEWAY_TOKEN=<random-uuid>
```

### 6.3 OpenClaw Configuration

**File:** `/opt/blytz/customers/{customer-id}/.openclaw/openclaw.json`

```json5
{
  gateway: {
    port: 18789,
    auth: {
      token: "${OPENCLAW_GATEWAY_TOKEN}"
    }
  },
  agents: {
    defaults: {
      workspace: "/root/.openclaw/workspace"
    }
  },
  channels: {
    telegram: {
      enabled: true,
      botToken: "${TELEGRAM_BOT_TOKEN}",
      dmPolicy: "open",
      allowFrom: ["*"]
    }
  }
}
```

### 6.4 Workspace File Templates

#### AGENTS.md

**Template:** `/opt/blytz/platform/templates/personal-assistant/AGENTS.md.tmpl`

```markdown
# AGENTS.md - Your Workspace

This folder is home. Treat it that way.

## Who You Help

{{.AssistantName}} is helping: {{.UserDescription}}

Your primary responsibilities:
{{.ResponsibilitiesList}}

## Every Session

Before doing anything else:
1. Read `SOUL.md` — this is who you are
2. Read `USER.md` — this is who you're helping
3. Check `memory/YYYY-MM-DD.md` for recent context

## How You Work

- Be proactive but not intrusive
- Remember context from previous conversations
- Actually do things, don't just talk about them
- When in doubt, ask for clarification

## Memory

Write important context to `memory/YYYY-MM-DD.md`.
Update `MEMORY.md` with long-term learnings.

## Safety

- Don't send emails or public posts without asking
- Don't share private client information
- `trash` > `rm` (recoverable beats gone forever)
- When in doubt, ask.

## Group Chats

If added to group chats, be smart about when to contribute:
- Respond when directly mentioned or asked
- Stay silent when it's casual conversation between humans
- Quality > quantity

## Make It Yours

This is a starting point. As you learn more about {{.AssistantName}}'s needs, update this file.
```

#### USER.md

**Template:** `/opt/blytz/platform/templates/personal-assistant/USER.md.tmpl`

```markdown
# USER.md - Your Human

## About

{{.CustomInstructions}}

## Preferences

(This section will grow over time as the assistant learns more)

## Notes

- Communication style: TBD
- Working hours: TBD
- Priority areas: TBD
```

#### SOUL.md

**Template:** `/opt/blytz/platform/templates/personal-assistant/SOUL.md.tmpl`

```markdown
# SOUL.md - Who You Are

You are {{.AssistantName}}, a personal AI assistant.

## Personality

- Helpful but not overbearing
- Proactive but respectful of boundaries
- Clear and concise in communication
- You remember context and use it to provide better help

## Capabilities

You have access to tools for:
- Web browsing and research
- File operations (read, write, edit)
- Executing commands (with permission)
- Managing schedules and reminders
- Processing and summarizing information

## Philosophy

Actually help. Don't just talk about helping.
If someone asks you to draft an email, write the email.
If someone asks you to research something, do the research.

Use your tools to get things done.
```

### 6.5 Caddyfile Structure

**File:** `/opt/blytz/caddy/Caddyfile`

```
{
    email admin@blytz.cloud
}

# Platform main site
blytz.cloud {
    reverse_proxy localhost:8080
    tls internal
}

# Customer subdomains (dynamically added)
customer-1.blytz.cloud {
    reverse_proxy localhost:30001
    tls internal
}

customer-2.blytz.cloud {
    reverse_proxy localhost:30002
    tls internal
}

# ... etc
```

---

## 7. Implementation Sprint

### 7.1 Day-by-Day Breakdown

| Day | Phase | Tasks | Deliverable |
|-----|-------|-------|-------------|
| **1** | Setup | Project init, go mod, config loading, basic server | Server starts, health check works |
| **2** | Database | SQLite connection, migrations, customer CRUD | Can create/query customers |
| **3** | Workspace | Template parsing, file generation | AGENTS.md, USER.md, SOUL.md generated |
| **4** | Docker | Docker SDK integration, compose generation | Container starts on assigned port |
| **5** | Provisioner | Full provisioning flow, port assignment | End-to-end: signup → container running |
| **6** | API | POST /signup, GET /status endpoints | API works with curl |
| **7** | Frontend | Embed HTML files, serve landing page | Form submission works |
| **8** | Caddy | Admin API integration, subdomain routing | Subdomain routes to container |
| **9** | Stripe | Checkout session creation | Payment flow starts |
| **10** | Webhooks | Handle payment success → trigger provisioning | Payment → container |
| **11** | Polish | Error handling, validation, logging | Robust error handling |
| **12** | systemd | Service file, install script | Runs as service |
| **13** | Testing | Integration tests, manual testing | All flows verified |
| **14** | Launch | Deploy to production, first pilot | Live platform |

### 7.2 Detailed Task Breakdown

#### Day 1: Project Setup

```
Tasks:
├── Initialize Go module
│   └── go mod init blytz
├── Create directory structure
├── Implement config loading
│   ├── Load from environment variables
│   ├── Load from config.env file
│   └── Validate required fields
├── Create basic Gin server
│   ├── Health endpoint
│   └── Request logging middleware
└── Test: curl localhost:8080/health returns 200

Files created:
├── go.mod
├── main.go
├── internal/config/config.go
└── internal/api/routes.go
```

#### Day 2: Database

```
Tasks:
├── Add sqlite dependency
│   └── go get modernc.org/sqlite
├── Implement database connection
│   ├── Connection pooling
│   └── Graceful shutdown
├── Implement migrations
│   ├── customers table
│   ├── audit_log table
│   └── port_allocations table
├── Implement Customer model
│   └── struct with all fields
├── Implement CRUD operations
│   ├── CreateCustomer
│   ├── GetCustomerByID
│   ├── GetCustomerByEmail
│   ├── UpdateCustomerStatus
│   └── CountActiveCustomers
└── Test: Can insert and query customers

Files created:
├── internal/db/db.go
├── internal/db/migrations.go
└── internal/db/customer.go
```

#### Day 3: Workspace Generator

```
Tasks:
├── Create template files
│   ├── AGENTS.md.tmpl
│   ├── USER.md.tmpl
│   └── SOUL.md.tmpl
├── Implement template parsing
│   └── Use Go's text/template
├── Implement file generation
│   ├── Parse custom instructions
│   ├── Extract responsibilities
│   ├── Generate all three files
│   └── Write to customer directory
├── Implement OpenClaw config generation
│   └── Generate openclaw.json
└── Test: Given input, files generated correctly

Files created:
├── internal/workspace/generator.go
└── internal/workspace/templates.go
```

#### Day 4: Docker Integration

```
Tasks:
├── Add Docker SDK dependency
│   └── go get github.com/docker/docker
├── Implement container operations
│   ├── List containers
│   ├── Create container
│   ├── Start container
│   ├── Stop container
│   └── Remove container
├── Implement compose file generation
│   ├── Parse template
│   ├── Inject port, customer ID
│   └── Write to customer directory
├── Implement port assignment
│   ├── Get next available port (30000-30999)
│   ├── Track in port_allocations table
│   └── Release on container removal
└── Test: Container starts, responds on assigned port

Files created:
├── internal/provisioner/provisioner.go
├── internal/provisioner/compose.go
└── internal/provisioner/ports.go
```

#### Day 5: Full Provisioning Flow

```
Tasks:
├── Wire all components together
│   ├── Create customer record
│   ├── Generate workspace files
│   ├── Generate compose file
│   ├── Create and start container
│   └── Update customer status
├── Implement Telegram token validation
│   └── Call getMe API, verify token works
├── Implement provisioning status polling
│   └── Container health check
├── Implement error handling
│   ├── Rollback on failure
│   └── Log detailed errors
└── Test: End-to-end provisioning works

Files created:
└── internal/telegram/validate.go
```

#### Day 6: API Endpoints

```
Tasks:
├── Implement POST /api/signup
│   ├── Validate input
│   ├── Check for duplicates
│   ├── Check capacity
│   ├── Validate Telegram token
│   ├── Create customer
│   └── Return checkout URL
├── Implement GET /api/status/:id
│   ├── Query customer
│   ├── Check container status
│   └── Return full status
├── Implement input validation
│   ├── Email format
│   ├── Telegram token format
│   └── Instruction length limits
└── Test: API works with curl

Files updated:
└── internal/api/handler.go
```

#### Day 7: Frontend

```
Tasks:
├── Create landing page HTML
│   ├── Hero section
│   ├── Email signup form
│   └── Basic styling
├── Create configuration page HTML
│   ├── Assistant name field
│   ├── Instructions textarea
│   ├── Telegram token field
│   └── Progress indicator
├── Create success page HTML
│   ├── Telegram link
│   ├── Subdomain link
│   └── Tips section
├── Embed files in binary
│   └── Use Go's embed package
├── Implement page serving
│   ├── GET / → landing
│   ├── GET /configure → form
│   └── GET /success → success
└── Test: Form submission creates customer

Files created:
├── static/index.html
├── static/configure.html
├── static/success.html
└── internal/api/static.go (embedding)
```

#### Day 8: Caddy Integration

```
Tasks:
├── Implement Caddyfile generation
│   ├── Base config
│   └── Dynamic subdomain entries
├── Implement Caddy Admin API client
│   ├── POST /load to reload config
│   └── Error handling
├── Add subdomain on provisioning
│   └── After container starts
├── Remove subdomain on suspension
│   └── After container stops
└── Test: Subdomain routes to container

Files created:
└── internal/caddy/caddy.go
```

#### Day 9-10: Stripe Integration

```
Tasks:
├── Add Stripe SDK dependency
│   └── go get github.com/stripe/stripe-go
├── Implement checkout session creation
│   ├── Create Stripe customer
│   ├── Create checkout session
│   └── Return checkout URL
├── Implement webhook handler
│   ├── Verify signature
│   ├── Handle checkout.session.completed
│   ├── Handle customer.subscription.deleted
│   ├── Handle invoice.payment_failed
│   └── Handle customer.subscription.updated
├── Update signup flow
│   └── Redirect to Stripe checkout
└── Test: Payment triggers provisioning

Files created:
├── internal/stripe/checkout.go
└── internal/stripe/webhook.go
```

#### Day 11: Polish

```
Tasks:
├── Comprehensive error handling
│   ├── User-friendly error messages
│   ├── Structured logging
│   └── Error tracking
├── Input validation
│   ├── All fields validated
│   ├── Length limits enforced
│   └── Sanitization
├── Logging
│   ├── Request logging
│   ├── Provisioning logs
│   └── Error logs
├── Graceful shutdown
│   └── Drain connections, stop containers
└── Test: All error cases handled

Files updated:
├── internal/api/middleware.go
└── internal/api/handler.go
```

#### Day 12: systemd Service

```
Tasks:
├── Create systemd unit file
├── Create install script
│   ├── Create blytz user
│   ├── Set up directories
│   ├── Copy binary
│   ├── Install service
│   └── Start service
├── Create uninstall script
├── Document installation
└── Test: Service starts on boot

Files created:
├── deployments/blytz.service
├── deployments/install.sh
└── deployments/uninstall.sh
```

#### Day 13: Testing

```
Tasks:
├── Write unit tests
│   ├── Config loading
│   ├── Database operations
│   ├── Template generation
│   └── Port allocation
├── Write integration tests
│   ├── Signup flow
│   ├── Provisioning flow
│   └── Status checking
├── Manual testing
│   ├── Full user flow
│   ├── Payment flow
│   └── Error scenarios
└── Fix any bugs found

Files created:
├── internal/config/config_test.go
├── internal/db/customer_test.go
├── internal/workspace/generator_test.go
└── internal/api/handler_test.go
```

#### Day 14: Launch

```
Tasks:
├── Deploy to production server
│   ├── Run install script
│   ├── Configure environment
│   └── Verify service running
├── Configure DNS
│   └── *.blytz.cloud → server IP
├── Configure Stripe webhooks
│   └── Add production endpoint
├── Monitor first provisioning
│   └── Watch logs closely
├── Onboard first pilot customer
│   └── Walk through flow manually
└── Document any issues
```

---

## 8. Deployment Guide

### 8.1 Prerequisites

```bash
# On Ubuntu 22.04 server

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Node.js (for OpenClaw in containers)
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs

# Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# Create blytz user
sudo useradd -r -s /bin/false blytz
sudo usermod -aG docker blytz

# Create directories
sudo mkdir -p /opt/blytz/{platform/templates/personal-assistant,customers,caddy,logs}
sudo chown -R blytz:blytz /opt/blytz
```

### 8.2 Configuration

**File:** `/opt/blytz/config.env`

```env
# API Keys
OPENAI_API_KEY=sk-xxx
STRIPE_SECRET_KEY=sk_live_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
STRIPE_PRICE_ID=price_xxx

# Platform Config
DATABASE_PATH=/opt/blytz/platform/database.sqlite
CUSTOMERS_DIR=/opt/blytz/customers
TEMPLATES_DIR=/opt/blytz/platform/templates
CADDYFILE_PATH=/opt/blytz/caddy/Caddyfile
LOG_PATH=/opt/blytz/logs/blytz.log
MAX_CUSTOMERS=20
PORT_RANGE_START=30000
PORT_RANGE_END=30999
BASE_DOMAIN=blytz.cloud
PLATFORM_PORT=8080

# Security
OPENCLAW_GATEWAY_TOKEN_PREFIX=blytz_
```

### 8.3 Installation

```bash
# Build binary
go build -o blytz ./cmd/server

# Copy binary
sudo cp blytz /opt/blytz/
sudo chmod +x /opt/blytz/blytz

# Copy templates
sudo cp -r internal/workspace/templates/* /opt/blytz/platform/templates/personal-assistant/

# Install systemd service
sudo cp deployments/blytz.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable blytz
sudo systemctl start blytz

# Check status
sudo systemctl status blytz
```

### 8.4 Caddy Setup

```bash
# Initial Caddyfile
cat << EOF | sudo tee /opt/blytz/caddy/Caddyfile
{
    email admin@blytz.cloud
}

blytz.cloud {
    reverse_proxy localhost:8080
    tls internal
}
EOF

# Start Caddy
sudo systemctl enable caddy
sudo systemctl start caddy
```

### 8.5 DNS Configuration

```
# Add wildcard DNS record
Type: A
Name: *.blytz.cloud
Value: <your-server-ip>
TTL: 300

# Add apex record
Type: A
Name: blytz.cloud
Value: <your-server-ip>
TTL: 300
```

### 8.6 Stripe Configuration

1. Create product in Stripe Dashboard
2. Create recurring price ($29/month)
3. Copy Price ID to config.env
4. Add webhook endpoint: `https://blytz.cloud/api/webhook/stripe`
5. Select events: `checkout.session.completed`, `customer.subscription.*`, `invoice.*`
6. Copy Webhook Secret to config.env

---

## 9. Testing Strategy

### 9.1 Unit Tests

| Package | Test Coverage Target |
|---------|---------------------|
| `config` | 90% |
| `db` | 85% |
| `workspace` | 90% |
| `provisioner` | 70% |
| `api` | 80% |

### 9.2 Integration Tests

```go
// Test full signup and provisioning flow
func TestSignupAndProvision(t *testing.T) {
    // 1. POST /api/signup
    // 2. Verify customer created in DB
    // 3. Verify workspace files generated
    // 4. Verify container started
    // 5. Verify Caddy subdomain added
    // 6. GET /api/status/:id returns active
}

// Test capacity limit
func TestCapacityLimit(t *testing.T) {
    // Create 20 customers
    // 21st signup should return 503
}
```

### 9.3 Manual Test Checklist

```
[ ] Landing page loads
[ ] Email signup works
[ ] Configuration form validates input
[ ] Invalid Telegram token shows error
[ ] Stripe checkout opens
[ ] Payment success triggers provisioning
[ ] Container starts
[ ] Subdomain routes to container
[ ] Telegram bot responds
[ ] Status endpoint shows active
[ ] Subscription cancellation suspends container
[ ] Error messages are user-friendly
[ ] Logs are written correctly
```

### 9.4 Load Testing

```bash
# Using hey for load testing
hey -n 100 -c 10 http://localhost:8080/api/health

# Should handle:
# - 100 requests/second to health endpoint
# - 10 concurrent signups (though only 1 will succeed per email)
```

---

## 10. Success Metrics

### 10.1 Day 14 Checklist

- [ ] Can signup with email
- [ ] Can configure assistant
- [ ] Stripe checkout works
- [ ] Container provisions in < 2 minutes
- [ ] Subdomain works (customer-id.blytz.cloud)
- [ ] Telegram bot responds to messages
- [ ] 1 paying pilot customer onboarded

### 10.2 Week 4 Goals

- [ ] 3 paying customers
- [ ] < 5% churn rate
- [ ] Average response time < 3 seconds
- [ ] Zero critical bugs
- [ ] NPS score > 8

### 10.3 Month 2 Goals

- [ ] 10 paying customers
- [ ] Feature requests documented
- [ ] Second channel (WhatsApp) in development
- [ ] Kubernetes migration planned

---

## 11. Risk Mitigation

### 11.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Docker daemon crashes | Low | High | systemd auto-restart, health checks |
| OpenClaw container OOM | Medium | Medium | 1GB limit per container, monitoring |
| SQLite corruption | Low | High | Daily backups, WAL mode |
| Stripe webhook missed | Low | Medium | Idempotent handling, manual sync job |
| Telegram API changes | Low | Medium | Pin OpenClaw version, monitor changelog |

### 11.2 Business Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Customer overload (>20) | Medium | High | Waitlist page, Kubernetes migration plan |
| High support burden | Medium | Medium | Self-service dashboard, FAQ docs |
| Payment disputes | Low | Low | Stripe handles, clear TOS |
| Competitor enters market | High | Medium | Focus on niche, build loyalty |

### 11.3 Backup Strategy

```bash
#!/bin/bash
# /opt/blytz/scripts/backup.sh
# Run daily via cron

DATE=$(date +%Y%m%d)
BACKUP_DIR="/opt/blytz/backups/$DATE"

mkdir -p $BACKUP_DIR

# Backup database
cp /opt/blytz/platform/database.sqlite $BACKUP_DIR/

# Backup customer workspaces (not container data)
tar -czf $BACKUP_DIR/customers.tar.gz /opt/blytz/customers/*/.openclaw/workspace

# Upload to S3 (optional)
# aws s3 sync $BACKUP_DIR s3://blytz-backups/$DATE/

# Keep only last 7 days locally
find /opt/blytz/backups -type d -mtime +7 -exec rm -rf {} +
```

### 11.4 Monitoring

```bash
# Simple health check script
# /opt/blytz/scripts/health-check.sh

#!/bin/bash

# Check platform is running
if ! systemctl is-active --quiet blytz; then
    echo "ALERT: blytz service down"
    # Send Telegram alert
    curl -s "https://api.telegram.org/bot${ALERT_BOT_TOKEN}/sendMessage" \
        -d chat_id="${ALERT_CHAT_ID}" \
        -d text="ALERT: Blytz platform service is down!"
fi

# Check each customer container
for CONTAINER in $(docker ps --filter "name=blytz-" --format "{{.Names}}"); do
    if ! docker inspect --format='{{.State.Health.Status}}' $CONTAINER | grep -q "healthy"; then
        echo "ALERT: Container $CONTAINER unhealthy"
    fi
done
```

---

## 12. Future Roadmap

### 12.1 Phase 2: Growth (After 10 Customers)

| Feature | Effort | Value |
|---------|--------|-------|
| Customer dashboard | 2 days | High |
| WhatsApp channel | 3 days | High |
| Usage analytics | 2 days | Medium |
| Template switching | 1 day | Medium |

### 12.2 Phase 3: Scale (After 15 Customers)

| Feature | Effort | Value |
|---------|--------|-------|
| Kubernetes migration | 1 week | High |
| Slack channel | 3 days | Medium |
| Team accounts | 5 days | High |
| API access | 3 days | Medium |

### 12.3 Phase 4: Enterprise (After 30 Customers)

| Feature | Effort | Value |
|---------|--------|-------|
| White-label offering | 2 weeks | High |
| SSO integration | 1 week | Medium |
| Dedicated hosting | 1 week | High |
| Custom model fine-tuning | 2 weeks | Medium |

---

## Appendix A: Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENAI_API_KEY` | Yes | - | OpenAI API key for customer containers |
| `STRIPE_SECRET_KEY` | Yes | - | Stripe secret key |
| `STRIPE_WEBHOOK_SECRET` | Yes | - | Stripe webhook signing secret |
| `STRIPE_PRICE_ID` | Yes | - | Stripe price ID for $29/month |
| `DATABASE_PATH` | No | `/opt/blytz/platform/database.sqlite` | SQLite database path |
| `CUSTOMERS_DIR` | No | `/opt/blytz/customers` | Customer data directory |
| `TEMPLATES_DIR` | No | `/opt/blytz/platform/templates` | Template files directory |
| `CADDYFILE_PATH` | No | `/opt/blytz/caddy/Caddyfile` | Caddyfile path |
| `LOG_PATH` | No | `/opt/blytz/logs/blytz.log` | Log file path |
| `MAX_CUSTOMERS` | No | `20` | Maximum customers |
| `PORT_RANGE_START` | No | `30000` | Starting port for containers |
| `PORT_RANGE_END` | No | `30999` | Ending port for containers |
| `BASE_DOMAIN` | No | `blytz.cloud` | Base domain for subdomains |
| `PLATFORM_PORT` | No | `8080` | Platform API port |

---

## Appendix B: Error Codes Reference

| Code | HTTP | User Message | Technical Details |
|------|------|--------------|-------------------|
| `VALIDATION_FAILED` | 400 | Please check your input | Field-specific validation errors |
| `INVALID_EMAIL` | 400 | Invalid email address | Email format check failed |
| `INVALID_BOT_TOKEN` | 400 | Invalid Telegram bot token | Telegram API returned error |
| `INSTRUCTIONS_TOO_LONG` | 400 | Instructions too long | Max 5000 characters |
| `ALREADY_EXISTS` | 409 | Account already exists | Email already registered |
| `NOT_FOUND` | 404 | Customer not found | Invalid customer ID |
| `AT_CAPACITY` | 503 | Platform at capacity | 20 customers already active |
| `PROVISIONING_FAILED` | 500 | Setup failed, contact support | Container start failed |
| `PAYMENT_FAILED` | 402 | Payment failed | Stripe returned error |
| `INTERNAL_ERROR` | 500 | Something went wrong | Unexpected error |

---

## Appendix C: API Request Examples

### Create Customer

```bash
curl -X POST https://blytz.cloud/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "assistant_name": "JARVIS",
    "custom_instructions": "I am a freelance software developer. I need help with:\n- Drafting client proposals\n- Researching new technologies\n- Managing my calendar\n- Following up on unpaid invoices",
    "telegram_bot_token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
  }'
```

### Check Status

```bash
curl https://blytz.cloud/api/status/john-example-com
```

### Stripe Webhook (handled by Stripe)

```json
{
  "id": "evt_123",
  "object": "event",
  "type": "checkout.session.completed",
  "data": {
    "object": {
      "id": "cs_123",
      "customer": "cus_123",
      "subscription": "sub_123",
      "metadata": {
        "customer_id": "john-example-com"
      }
    }
  }
}
```

---

**Document Version:** 1.0  
**Last Updated:** 2026-02-18  
**Author:** Blytz Team  
**Status:** Ready for Implementation
