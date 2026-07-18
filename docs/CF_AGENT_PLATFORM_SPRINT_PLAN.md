---
type: plan
title: Sprint Plan: AgentForge - AI Employee Platform
resource: blytz-cloud
description: "**Goal**: Build a skill-based AI orchestrator for SMB owners - 'one AI, your skills, your data'"
tags: [cloudflare, d1, durable-objects, planning, r2, stripe, typescript, workers]
updated: 2026-06-18
---

# Sprint Plan: AgentForge - AI Employee Platform

## Project Overview

**Goal**: Build a skill-based AI orchestrator for SMB owners - "one AI, your skills, your data"

**Target User**: Small/Medium Business Owners & Founders in Malaysia/SEA

**Key Assumptions**:
- New project (fresh start, not modifying BlytzCloud)
- Full Cloudflare stack (Workers, DO, D1, Vectorize, Workers AI)
- Single AI orchestrator + toggleable skills (not marketplace)
- Kimi K2.5 as primary LLM (256k context for business knowledge)
- Pricing: Starter (RM99), Pro (RM299), Business (RM799)

---

## Sprint 1: Foundation & Landing

**Duration**: 1 week  
**Focus**: Project setup, landing page, user signup

### Tasks

| # | Task | Size | Depends On | Files |
|---|------|------|-----------|-------|
| 1 | Initialize CF Workers project | S | - | `wrangler.toml`, `package.json` |
| 2 | Configure TypeScript and dependencies | S | 1 | `tsconfig.json`, `package.json` |
| 3 | Set up project structure | S | 1 | `src/workers/`, `src/agents/`, `src/utils/` |
| 4 | Configure wrangler (AI, D1, Vectorize, R2) | M | 1 | `wrangler.toml` |
| 5 | Create D1 database schema | M | - | `migrations/001_initial.sql` |
| 6 | Create landing page (Skills concept) | M | - | `frontend/index.html` |
| 7 | Implement magic link auth | L | - | `src/workers/auth.ts` |
| 8 | Create basic user dashboard | M | 7 | `frontend/dashboard/index.html` |

### Deliverables
- Working CF project with all bindings
- Landing page showcasing skills
- User registration + basic dashboard
- Database schema ready

---

## Sprint 2: Onboarding & Skills Configuration

**Duration**: 1 week  
**Focus**: Onboarding flow - knowledge, skills, behavior

### Tasks

| # | Task | Size | Depends On | Files |
|---|------|------|-----------|-------|
| 9 | Build skills selection UI (10 skills) | M | - | `frontend/onboarding/skills.html` |
| 10 | Create knowledge upload UI | M | - | `frontend/onboarding/knowledge.html` |
| 11 | Implement document storage (R2) | L | 5 | `src/utils/storage.ts` |
| 12 | Build behavior configurator | M | 9 | `frontend/onboarding/behavior.html` |
| 13 | Implement channel selection | M | - | `frontend/onboarding/channels.html` |
| 14 | Create pricing page with 3 tiers | S | - | `frontend/pricing.html` |
| 15 | Integrate Stripe checkout | L | - | `src/workers/checkout.ts` |
| 16 | Webhook handler for payment | M | 15 | `src/workers/webhooks.ts` |

### Deliverables
- Complete onboarding flow
- Knowledge upload and storage
- Skills selection UI with all 10 skills
- Stripe integration for payments

---

## Sprint 3: Orchestrator + AI Integration

**Duration**: 1-2 weeks  
**Focus**: The AI brain - Durable Object with Kimi K2.5

### Tasks

| # | Task | Size | Depends On | Files |
|---|------|------|-----------|-------|
| 17 | Implement DO dispatcher | L | 5, 16 | `src/workers/dispatcher.ts` |
| 18 | Create BusinessOrchestrator (AIChatAgent) | L | - | `src/agents/orchestrator.ts` |
| 19 | Integrate Kimi K2.5 with prompt engineering | M | 18 | `src/utils/llm.ts` |
| 20 | Implement skills system prompt builder | M | 18 | `src/agents/prompts/skills.ts` |
| 21 | Add Vectorize knowledge retrieval | L | 11 | `src/utils/vector.ts` |
| 22 | Implement skills execution engine | M | 20 | `src/agents/skills/engine.ts` |
| 23 | Add streaming response support | M | 19 | `src/utils/stream.ts` |
| 24 | Build conversation storage | M | 18 | `src/utils/db.ts` |

### Deliverables
- Functional AI orchestrator with Kimi K2.5
- Skills can be enabled/disabled
- Knowledge retrieval working
- Streaming responses

---

## Sprint 4: Channel Integrations

**Duration**: 1 week  
**Focus**: Connect WhatsApp, Telegram, and web

### Tasks

| # | Task | Size | Depends On | Files |
|---|------|------|-----------|-------|
| 25 | Implement WhatsApp webhook handler | M | 16, 17 | `src/workers/whatsapp.ts` |
| 26 | Implement Telegram bot handler | M | 17 | `src/workers/telegram.ts` |
| 27 | Create website chat widget | M | - | `frontend/widget/chat.js` |
| 28 | Implement WebSocket handler | M | 17 | `src/workers/websocket.ts` |
| 29 | Add channel message routing | M | 25, 26, 28 | `src/agents/channels/router.ts` |
| 30 | Build analytics dashboard | M | 24 | `frontend/dashboard/analytics.html` |

### Deliverables
- WhatsApp integration working
- Telegram bot working
- Website chat widget
- Basic analytics dashboard

---

## Sprint 5: Production Readiness

**Duration**: 1 week  
**Focus**: Polish, monitoring, enterprise

### Tasks

| # | Task | Size | Depends On | Files |
|---|------|------|-----------|-------|
| 31 | Set up CI/CD with GitHub Actions | M | - | `.github/workflows/deploy.yml` |
| 32 | Add comprehensive error handling | M | All | `src/utils/errors.ts` |
| 33 | Implement health check endpoint | S | - | `src/workers/health.ts` |
| 34 | Add budget alerts | S | - | `wrangler.toml` |
| 35 | Performance optimization | L | All | `src/utils/cache.ts` |
| 36 | Add team members feature | M | 24 | `src/workers/team.ts` |
| 37 | API access for developers | M | - | `src/workers/api.ts` |
| 38 | Prepare Enterprise tier page | S | - | `frontend/enterprise.html` |

### Deliverables
- Production-ready deployment
- Team collaboration
- API for integrations
- Enterprise landing page

---

## Total Effort Estimate

| Sprint | Tasks | Duration |
|--------|-------|----------|
| Sprint 1 | 8 tasks | 1 week |
| Sprint 2 | 8 tasks | 1 week |
| Sprint 3 | 8 tasks | 1-2 weeks |
| Sprint 4 | 6 tasks | 1 week |
| Sprint 5 | 8 tasks | 1 week |
| **Total** | **38 tasks** | **5-6 weeks** |

---

## Technical Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     AGENTFORGE PLATFORM                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  FRONTEND (Cloudflare Pages)                                    │
│  ├── Landing page (skills showcase)                            │
│  ├── Onboarding (skills, knowledge, behavior, channels)      │
│  ├── Dashboard (analytics, settings)                            │
│  └── Widget (website chat)                                      │
│                                                                  │
│  API LAYER (Cloudflare Workers)                                 │
│  ├── Auth (magic link)                                         │
│  ├── Onboarding API                                            │
│  ├── Stripe checkout                                            │
│  ├── Webhooks (payment, WhatsApp, Telegram)                     │
│  └── Dispatcher → Durable Objects                               │
│                                                                  │
│  AGENT LAYER (Durable Objects)                                   │
│  ├── BusinessOrchestrator (AIChatAgent)                        │
│  ├── Skills Engine (enable/disable, execute)                   │
│  ├── Knowledge Retrieval (Vectorize)                            │
│  └── State (SQLite in DO)                                       │
│                                                                  │
│  AI LAYER (Workers AI)                                          │
│  ├── Kimi K2.5 (primary)                                       │
│  ├── System prompt from skills                                  │
│  └── Tool calling                                               │
│                                                                  │
│  DATA LAYER                                                      │
│  ├── D1: Users, Subscriptions, Configs                         │
│  ├── Vectorize: Knowledge embeddings                            │
│  └── R2: Document storage                                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Skills Implementation

### Core Skills (Always Available)
1. **Knowledge Q&A** - Answer from uploaded docs
2. **Appointment Booking** - Schedule meetings
3. **Lead Capture** - Collect visitor info
4. **Order Processing** - Handle transactions

### Advanced Skills
5. **Lead Qualification** - Score leads
6. **Follow-up Sequences** - Automated nurturing
7. **Complaint Handling** - Triage issues
8. **Invoice Handler** - Send invoices
9. **Report Generation** - Summarize data
10. **Inventory Alert** - Monitor stock

---

## Cost Estimates (MYR)

| Users | CF Infra | LLM Cost (pass-through) | Total |
|-------|----------|-------------------------|-------|
| 100 (free tier) | Free | ~RM 200 | ~RM 200 |
| 500 | RM 20 | ~RM 1,000 | ~RM 1,000 |
| 1,000 | RM 40 | ~RM 2,000 | ~RM 2,000 |
| 5,000 | RM 100 | ~RM 10,000 | ~RM 10,000 |

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Single orchestrator | One AI per business, skills are toggles |
| Skills as config | No code needed to enable capabilities |
| DO per user | Isolation, scale-to-zero when idle |
| Kimi K2.5 | 256k context fits business docs |
| Magic link auth | Simple for non-tech users |
| WhatsApp first | Dominant in SEA (90%+) |

---

## Definition of Done

- [ ] All 38 tasks completed
- [ ] Landing page showcasing skills concept
- [ ] Full onboarding flow working
- [ ] AI orchestrator with Kimi K2.5
- [ ] 10 skills implemented (toggleable)
- [ ] WhatsApp integration working
- [ ] Stripe payment working
- [ ] Analytics dashboard
- [ ] Deployed to CF production
- [ ] No critical security issues

---

## Next Steps

1. ✅ Sprint Plan Updated (v3.0 - Skills + Orchestrator)
2. Start Sprint 1: Foundation
3. Set up CF account and get API keys
4. Run `npm create cloudflare@latest -- agentforge`

---

**Version**: 3.0  
**Updated**: April 6, 2026