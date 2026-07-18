---
type: architecture
title: Architecture Specification: CF-Native AI Agent Platform
resource: blytz-cloud
description: "**Version**: 1.0"
tags: [architecture, cloudflare, d1, docker, durable-objects, go, nextjs, r2, redis, stripe, typescript, workers]
updated: 2026-06-18
---

# Architecture Specification: CF-Native AI Agent Platform

## Project: Agent-as-a-Service on Cloudflare

**Version**: 1.0  
**Status**: Draft  
**Last Updated**: April 2026

---

## 1. Executive Summary

### Purpose
Build a serverless AI agent platform targeting Malaysian and Southeast Asian users. The platform allows users to deploy personal AI agents that scale to zero, with minimal infrastructure costs.

### Key Technology Choices
| Component | Choice | Rationale |
|-----------|--------|------------|
| Compute | Cloudflare Workers + Durable Objects | Stateful, edge-first, scale-to-zero |
| AI | Workers AI (Kimi K2.5) | 256k context, tool calling, 77% cheaper than proprietary |
| Database | D1 (SQLite) | Serverless, familiar API, generous free tier |
| Memory | Vectorize | Semantic search, RAG pipeline |
| Storage | R2 | No egress fees, file storage |
| Frontend | Next.js on Cloudflare Pages | SSR, edge rendering |

### Success Metrics
- **Cold start**: < 2 seconds
- **Monthly cost**: < $50 for 1K users (excluding LLM)
- **Uptime**: 99.9%
- **Latency**: < 100ms from Malaysia

---

## 2. Architecture Overview

### High-Level Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              USER LAYER                                      │
│                                                                              │
│   ┌────────────┐   ┌────────────┐   ┌────────────┐   ┌─────────────────┐   │
│   │  Web UI    │   │ Telegram   │   │ Discord    │   │  REST API       │   │
│   │  (Pages)   │   │  (MCP)     │   │  (MCP)     │   │  (Mobile/3rdP)  │   │
│   └─────┬──────┘   └─────┬──────┘   └─────┬──────┘   └────────┬────────┘   │
│         │                │                │                    │             │
└─────────┼────────────────┼────────────────┼────────────────────┼─────────────┘
          │                │                │                    │
          ▼                ▼                ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API GATEWAY LAYER                                   │
│                                                                              │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │                        Cloudflare Workers                           │   │
│   │                                                                       │   │
│   │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌─────────────┐  │   │
│   │  │  Auth      │  │ Rate Limit │  │ WebSocket  │  │  Webhook    │  │   │
│   │  │  (Zero     │  │ (AI        │  │  Gateway   │  │  Handler    │  │   │
│   │  │   Trust)   │  │  Gateway)  │  │            │  │             │  │   │
│   │  └────────────┘  └────────────┘  └────────────┘  └─────────────┘  │   │
│   │                                                                       │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                    │                                          │
│                                    ▼                                          │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │                     Durable Object Dispatcher                       │   │
│   │                     (Routes to per-user agent DO)                   │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          AGENT EXECUTION LAYER                               │
│                                                                              │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │                PER-USER DURABLE OBJECTS (Pool)                     │   │
│   │                                                                      │   │
│   │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │   │
│   │  │   Agent A   │ │   Agent B   │ │   Agent C   │ │   Agent N   │   │   │
│   │  │  (User 1)   │ │  (User 2)   │ │  (User 3)   │ │  (User N)   │   │   │
│   │  │             │ │             │ │             │ │             │   │   │
│   │  │ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │   │   │
│   │  │ │SQLite  │ │ │ │SQLite  │ │ │ │SQLite  │ │ │ │SQLite  │ │   │   │
│   │  │ │ State  │ │ │ │ State  │ │ │ │ State  │ │ │ │ State  │ │   │   │
│   │  │ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │   │   │
│   │  │             │ │             │ │             │ │             │   │   │
│   │  │ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │   │   │
│   │  │ │ Message │ │ │ │ Message │ │ │ │ Message │ │ │ │ Message │ │   │   │
│   │  │ │ History │ │ │ │ History │ │ │ │ History │ │ │ │ History │ │   │   │
│   │  │ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │   │   │
│   │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │   │
│   │                                                                      │   │
│   │              Each DO:                                                │   │
│   │              - Activates on first request                           │   │
│   │              - Hibernates after 30s idle                            │   │
│   │              - Pays $0 when sleeping                                │   │
│   │                                                                      │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                    │                                          │
│                                    ▼                                          │
│   ┌─────────────────────┐  ┌─────────────────────┐  ┌───────────────────┐ │
│   │    Workers AI       │  │    AI Gateway       │  │   Tool Executor   │ │
│   │   (Kimi K2.5)       │  │   (Rate/Cache)      │  │   (HTTP/Browser)  │ │
│   └─────────────────────┘  └─────────────────────┘  └───────────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            DATA LAYER                                        │
│                                                                              │
│   ┌──────────────────┐  ┌──────────────────┐  ┌─────────────────────────┐ │
│   │       D1          │  │    Vectorize     │  │          R2             │ │
│   │  (User DB,       │  │  (Agent Memory,  │  │  (Files, Images,       │ │
│   │   Subscriptions) │  │   RAG Context)   │  │   Attachments)        │ │
│   └──────────────────┘  └──────────────────┘  └─────────────────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Component Specifications

### 3.1 API Gateway (Workers)

**Purpose**: Handle all incoming requests, authentication, rate limiting

**Components**:
- `src/workers/entry.ts` - Main entry point
- `src/workers/auth.ts` - Authentication middleware
- `src/workers/ratelimit.ts` - Rate limiting via AI Gateway
- `src/workers/websocket.ts` - WebSocket upgrade handler
- `src/workers/api.ts` - REST API routes

**Endpoints**:
```
POST   /api/v1/auth/signup        - Create new user
POST   /api/v1/auth/login         - Login, get JWT
GET    /api/v1/agent              - Get user's agent status
POST   /api/v1/agent/message     - Send message to agent
GET    /api/v1/agent/history     - Get conversation history
PUT    /api/v1/agent/config      - Update agent configuration
POST   /api/v1/webhooks/stripe   - Stripe webhook handler
GET    /health                   - Health check
```

**Authentication**:
- JWT-based (CF Access or custom)
- Short-lived access tokens (15 min)
- Refresh tokens stored in D1

### 3.2 Durable Objects (Agent Runtime)

**Purpose**: Per-user agent instance with state persistence

**Class**: `AIChatAgent` from Agents SDK

**Structure**:
```
src/
├── agents/
│   ├── base-agent.ts      - Base class with SQLite
│   ├── chat-agent.ts      - Chat-specific logic
│   ├── tools/
│   │   ├── http.ts        - HTTP fetch tool
│   │   ├── browser.ts     - Headless browser tool
│   │   ├── memory.ts      - Long-term memory tool
│   │   └── ...            - Additional tools
│   └── prompts/
│       ├── system.md      - System prompt
│       └── tools.md       - Tool definitions
```

**State Schema** (SQLite in DO):
```sql
CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  role TEXT NOT NULL,        -- 'user' | 'assistant' | 'system'
  content TEXT NOT NULL,
  tool_calls TEXT,           -- JSON array of tool calls
  created_at INTEGER NOT NULL
);

CREATE TABLE agent_config (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE INDEX idx_messages_created ON messages(created_at);
```

**Tool Calling**:
- Kimi K2.5 native function calling
- Server-side tools (execute in DO)
- Client-side tools (execute in browser)

### 3.3 Workers AI Integration

**Provider**: Kimi K2.5 (`@cf/moonshotai/kimi-k2.5`)

**Configuration**:
```typescript
const modelSettings = {
  model: "@cf/moonshotai/kimi-k2.5",
  max_tokens: 4096,
  temperature: 0.7,
  tools: [
    // Tool definitions
  ]
};
```

**Optimization**:
- Session affinity header for prefix caching
- Async API for non-real-time tasks
- Streaming for real-time responses

### 3.4 Data Layer

#### D1 Database Schema

```sql
-- Users table
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  created_at INTEGER NOT NULL,
  subscription_tier TEXT DEFAULT 'free',
  stripe_customer_id TEXT,
  api_key TEXT UNIQUE
);

-- Subscriptions table
CREATE TABLE subscriptions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  status TEXT NOT NULL,           -- 'active' | 'cancelled' | 'past_due'
  plan_id TEXT NOT NULL,
  current_period_end INTEGER,
  created_at INTEGER NOT NULL
);

-- API Keys table
CREATE TABLE api_keys (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  key_hash TEXT NOT NULL,
  name TEXT,
  created_at INTEGER NOT NULL,
  last_used_at INTEGER
);
```

#### Vectorize Index

```typescript
// Index configuration
const memoryIndex = {
  indexName: "agent-memories",
  dimension: 1024,  // Kimi embedding dimension
  metric: "cosine"
};
```

### 3.5 Frontend (Cloudflare Pages)

**Stack**: Next.js or plain HTML/JS

**Pages**:
- `/` - Landing page
- `/login` - Login form
- `/signup` - Signup form
- `/dashboard` - Main chat interface
- `/settings` - Agent configuration
- `/billing` - Subscription management

---

## 4. Data Flow

### 4.1 User Sends Message

```
1. User (browser/mobile) sends message
   │
   ▼
2. POST /api/v1/agent/message
   │
   ▼
3. Worker validates JWT, checks rate limit
   │
   ▼
4. Worker looks up user's DO ID from D1
   │
   ▼
5. Worker forwards to Durable Object
   │
   ▼
6. DO receives message, stores in SQLite
   │
   ▼
7. DO calls Workers AI (Kimi K2.5)
   │
   ▼
8. Kimi responds with text + tool calls
   │
   ▼
9. DO executes tools if needed
   │
   ▼
10. DO updates conversation history
    │
    ▼
11. DO streams response back to Worker
    │
    ▼
12. Worker streams to client via WebSocket
    │
    ▼
13. Response displayed in UI
    │
    ▼
14. (Optional) DO stores summary in Vectorize
```

### 4.2 Agent Initialization (First Request)

```
1. User visits dashboard (first time)
   │
   ▼
2. Worker checks if DO exists in D1
   │
   ▼
3. If not, create new DO with unique ID
   │
   ▼
4. Store DO ID in D1 user record
   │
   ▼
5. Initialize DO state (empty SQLite)
   │
   ▼
6. Return WebSocket connection to client
   │
   ▼
7. DO is now ready for messages
```

---

## 5. Security

### 5.1 Authentication Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   AUTHENTICATION FLOW                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   User                                                        │
│     │                                                        │
│     ▼                                                        │
│   POST /api/v1/auth/login {email, password}                  │
│     │                                                        │
│     ▼                                                        │
│   ┌──────────────────────────────────────────────┐          │
│   │           Worker validates credentials       │          │
│   │           against hashed password in D1      │          │
│   └──────────────────────────────────────────────┘          │
│     │                                                        │
│     ▼                                                        │
│   JWT Token generated (15 min expiry)                      │
│     │                                                        │
│     ▼                                                        │
│   Refresh token stored in D1                                │
│     │                                                        │
│     ▼                                                        │
│   Return {access_token, refresh_token}                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 Security Checklist

- [ ] All API endpoints require authentication (except /health, /login, /signup)
- [ ] JWT tokens are short-lived (15 min)
- [ ] Passwords hashed with bcrypt/argon2
- [ ] Rate limiting per user via AI Gateway
- [ ] Input sanitization on all user inputs
- [ ] No secrets in code (use CF Secrets)
- [ ] CORS configured for frontend origin only
- [ ] SQL queries use parameterized statements
- [ ] WebSocket connections authenticated

---

## 6. Scalability Design

### 6.1 Scale-to-Zero Pattern

```
                    Active Usage
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
    ┌─────────┐    ┌─────────┐    ┌─────────┐
    │  DO     │    │  DO     │    │  DO     │
    │ Running │    │ Running │    │ Running │
    └─────────┘    └─────────┘    └─────────┘
         │               │               │
         │   30s idle    │   30s idle    │
         ▼               ▼               ▼
    ┌─────────┐    ┌─────────┐    ┌─────────┐
    │   DO    │    │   DO    │    │   DO    │
    │Hibernated│    │Hibernated│    │Hibernated│
    └─────────┘    └─────────┘    └─────────┘
         │               │               │
         └───────────────┼───────────────┘
                         │
                    $0 cost per DO
               (no duration billing)
```

### 6.2 Capacity Planning

| Resource | Free Limit | Paid Limit | Per User |
|----------|------------|------------|----------|
| DO Requests | 100K/day | $0.20/1M | ~100/day |
| DO Duration | 13K GB-s/day | $12.50/1M GB-s | ~1 GB-s/day |
| D1 Reads | 5M/day | $1/1M | ~5K/day |
| D1 Writes | 100K/day | $1/1M | ~100/day |
| Workers AI | 10K neurons/day | $0.011/1K | ~500/day |

### 6.3 Limits and Throttling

- **Message rate**: 10 messages/minute per user
- **Token limit**: 100K tokens/conversation
- **Concurrent connections**: 10 WebSocket connections per user
- **Tool calls per message**: 5 max

---

## 7. Cost Optimization

### 7.1 Free Tier Usage Strategy

1. **DO hibernation**: Minimize active duration
2. **Prefix caching**: Reuse context via session affinity
3. **D1 query optimization**: Use indexes, batch reads
4. **Vectorize**: Limit memory storage, implement eviction

### 7.2 Caching Strategy

```
┌─────────────────────────────────────────────────────────────────┐
│                      CACHING LAYERS                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │
│  │ AI Gateway  │    │  D1 Cache   │    │Vectorize    │       │
│  │ (Rate/Cache)│    │  (Redis)    │    │  (Memory)   │       │
│  └─────────────┘    └─────────────┘    └─────────────┘       │
│                                                                  │
│  - Cache LLM responses for identical prompts                   │
│  - Cache user lookups                                           │
│  - Cache frequently accessed Vectorize results                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 8. Monitoring & Observability

### 8.1 Metrics to Track

| Metric | Source | Alert Threshold |
|--------|--------|-----------------|
| Request latency (p50, p95, p99) | Cloudflare Analytics | > 500ms |
| DO cold starts | Custom | > 10% of requests |
| LLM token usage | Workers AI logs | > 10x baseline |
| Error rate | Cloudflare Analytics | > 1% |
| DO duration | Cloudflare Analytics | > 80% of limit |

### 8.2 Logging

- **Structured logs**: JSON format with request ID, user ID
- **Log levels**: DEBUG, INFO, WARN, ERROR
- **Retention**: 7 days (free), 30 days (paid)

---

## 9. Implementation Notes

### 9.1 Key Dependencies

```json
{
  "dependencies": {
    "agents": "latest",
    "workers-ai-provider": "latest",
    "ai": "latest",
    "hono": "latest",
    "@cloudflare/workers-types": "latest"
  }
}
```

### 9.2 Project Structure

```
cf-agent-platform/
├── src/
│   ├── workers/          # API Workers
│   ├── agents/           # Agent implementations
│   ├── utils/            # Shared utilities
│   └── types/            # TypeScript types
├── migrations/           # D1 migrations
├── frontend/             # UI (Next.js or plain)
├── test/                 # Tests
├── wrangler.toml         # CF config
├── package.json
├── tsconfig.json
└── README.md
```

### 9.3 Environment Variables

```bash
# Required
CF_ACCOUNT_ID=...
CF_API_TOKEN=...
JWT_SECRET=...

# Optional
STRIPE_SECRET_KEY=...
STRIPE_WEBHOOK_SECRET=...
OPENAI_API_KEY=...  # Fallback
```

---

## 10. Appendix

### A. Kimi K2.5 Pricing Reference

| Token Type | Price (per 1M) |
|------------|----------------|
| Input | $0.60 |
| Cached Input | $0.10 |
| Output | $3.00 |

### B. Free Tier Limits Reference

| Resource | Limit |
|----------|-------|
| Workers Requests | 100K/day |
| DO Requests | 100K/day |
| DO Duration | 13K GB-s/day |
| D1 Reads | 5M/day |
| D1 Writes | 100K/day |
| D1 Storage | 5GB |
| Vectorize Vectors | 5M |
| Workers AI | 10K neurons/day |

### C. Comparison with BlytzCloud (VPS)

| Aspect | BlytzCloud (VPS) | CF Agent Platform |
|--------|------------------|-------------------|
| Deployment | Docker containers | Durable Objects |
| Scaling | Manual | Auto (scale-to-zero) |
| Cost (1K users) | ~$160/mo | ~$50/mo |
| Latency | Depends on VPS | < 100ms (edge) |
| Maintenance | High | Low |
| Complexity | Medium | Medium |

---

## 11. Approval

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Architect | Claude | 2026-04-06 | [Pending] |
| Lead Developer | [Pending] | [Pending] | [Pending] |
| Product Owner | [Pending] | [Pending] | [Pending] |

---

*Document Version: 1.0*  
*Next Review: After Sprint 2*