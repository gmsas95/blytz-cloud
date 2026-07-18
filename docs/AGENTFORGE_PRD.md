---
type: guide
title: Product Requirements Document (PRD)
resource: blytz-cloud
description: "**Version**: 3.0"
tags: [cloudflare, d1, durable-objects, r2, stripe, typescript, workers]
updated: 2026-06-18
---

# Product Requirements Document (PRD)
# AgentForge - AI Employee Platform

**Version**: 3.0  
**Status**: Ready for Development  
**Date**: April 6, 2026

---

## 1. Executive Summary

### Vision
**"Your business has one AI employee - configured with skills you choose, trained on your data"**

### Target User
SMB Owners & Founders in Malaysia/SEA who need help with business operations but don't want to manage multiple AI tools.

### Core Approach
**Single AI Orchestrator + Toggleable Skills** - One AI instance per business, owner enables/disables skills based on their needs, trains it with their documents.

---

## 2. Architecture Overview

### The Stack (Fully Cloudflare)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         AGENTFORGE STACK                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                    FRONTEND (Cloudflare Pages)                    │ │
│  │  Landing → Dashboard → Onboarding → Settings                      │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                    │                                     │
│                                    ▼                                     │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                    API GATEWAY (Workers)                          │ │
│  │  - Auth (magic link)                                               │ │
│  │  - WebSocket routing                                               │ │
│  │  - Stripe webhooks                                                │ │
│  │  - Channel webhooks (WhatsApp, Telegram)                          │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                    │                                     │
│                                    ▼                                     │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │              ORCHESTRATOR (Durable Object per user)               │ │
│  │                                                                        │ │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐       │ │
│  │  │ Business AI    │  │ Skills Engine  │  │ Knowledge Base│       │ │
│  │  │ (Kimi K2.5)    │  │ (enable/disable)│  │ (Vectorize)    │       │ │
│  │  └────────────────┘  └────────────────┘  └────────────────┘       │ │
│  │                                                                        │ │
│  │  State: SQLite (built-in)  │  Memory: Vectorize                    │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                    │                                     │
│                                    ▼                                     │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                       DATA LAYER                                   │ │
│  │                                                                        │ │
│  │  D1: Users, Subscriptions, Configs                                  │ │
│  │  Vectorize: Knowledge embeddings                                   │ │
│  │  R2: Document storage                                              │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. User Experience

### Onboarding Flow

```
STEP 1: Create Account
         ↓
STEP 2: Upload Knowledge
        ├── Upload PDFs, documents
        ├── Paste FAQ
        └── Add business rules
         ↓
STEP 3: Enable Skills (Checkboxes)
        ┌─────────────────────────────────────┐
        │ □ Knowledge Q&A        (Always on)   │
        │ □ Appointment Booking               │
        │ □ Lead Capture                     │
        │ □ Order Processing                 │
        │ □ Lead Qualification                │
        │ □ Follow-up Sequences              │
        │ □ Complaint Handling               │
        │ □ Invoice Handler                  │
        │ □ Report Generation                 │
        │ □ Inventory Alert                   │
        └─────────────────────────────────────┘
         ↓
STEP 4: Configure Behavior
        ├── Response style: [Formal] [Friendly] [Casual]
        ├── Language: [EN] [BM] [CN] [Mix]
        ├── Working hours: [24/7] [9-6] [Custom]
        └── Escalation: [Always] [Smart] [Never]
         ↓
STEP 5: Connect Channels
        ├── WhatsApp Business
        ├── Telegram Bot
        └── Website Chat Widget
         ↓
STEP 6: Subscribe
        ├── Starter: RM99/mo
        ├── Pro: RM299/mo
        └── Business: RM799/mo
         ↓
STEP 7: AI Goes Live
        └── Starts learning from conversations
```

---

## 4. Skills Library

### What the AI Can Do

| Skill | Description | When to Enable |
|-------|-------------|----------------|
| **Knowledge Q&A** | Answer questions from uploaded docs | Always |
| **Appointment Booking** | Schedule meetings, consultations | Service businesses |
| **Lead Capture** | Collect visitor name, email, interest | All businesses |
| **Order Processing** | Handle orders, payments | E-commerce, retail |
| **Lead Qualification** | Score and categorize leads | B2B, services |
| **Follow-up Sequences** | Automated nurturing emails | Sales teams |
| **Complaint Handling** | Triage and escalate issues | Customer service |
| **Invoice Handler** | Send invoices, track payments | Freelancers, agencies |
| **Report Generation** | Summarize daily/weekly data | All businesses |
| **Inventory Alert** | Monitor stock levels | Retail, F&B |

### Skill Implementation

Each skill is a prompt configuration + tool capability:

```typescript
// Skills are enabled/disabled in the orchestrator
interface UserConfig {
  enabledSkills: Set<SkillType>;
  knowledgeBaseId: string;
  behavior: BehaviorConfig;
  channels: ChannelConfig[];
}

// The orchestrator decides which skill to use based on message intent
class BusinessOrchestrator extends AIChatAgent<Env> {
  async onChatMessage(message: string) {
    // 1. Detect intent
    const intent = await this.detectIntent(message);
    
    // 2. Find matching enabled skill
    const skill = this.getMatchingSkill(intent, this.config.enabledSkills);
    
    // 3. Retrieve relevant knowledge
    const context = await this.retrieveKnowledge(message);
    
    // 4. Execute skill with context
    const response = await this.executeSkill(skill, message, context);
    
    // 5. Return response
    return response;
  }
}
```

---

## 5. Pricing Tiers (Global)

### Price by Region

| Tier | MYR | USD | SGD | IDR |
|------|-----|-----|-----|-----|
| **Starter** | RM 99/mo | $25 | $33 | 400K |
| **Pro** | RM 299/mo | $75 | $100 | 1.2M |
| **Business** | RM 799/mo | $200 | $267 | 3.2M |
| **Enterprise** | Custom | Custom | Custom | Custom |

### Tier Features

| Feature | Starter | Pro | Business |
|---------|---------|-----|----------|
| Skills enabled | 4 | 8 | All 10 |
| Knowledge pages | 50 | 200 | Unlimited |
| Messages/month | 500 | 2,000 | Unlimited |
| Channels | 1 | 3 | All |
| Custom training | ❌ | ✅ | ✅ |
| Priority support | ❌ | ❌ | ✅ |
| Analytics | Basic | Advanced | Full |
| API access | ❌ | ✅ | ✅ |
| Team members | 1 | 3 | 10 |

---

## 6. Technical Implementation

### Project Structure

```
agentforge/
├── src/
│   ├── workers/           # API Gateway
│   │   ├── index.ts       # Entry point
│   │   ├── auth.ts        # Magic link auth
│   │   ├── api.ts         # REST endpoints
│   │   ├── webhooks.ts    # Stripe, WhatsApp, Telegram
│   │   └── dispatcher.ts  # Route to DO
│   │
│   ├── agents/
│   │   ├── orchestrator.ts    # Main AIChatAgent
│   │   ├── skills/
│   │   │   ├── knowledge.ts    # Q&A skill
│   │   │   ├── appointment.ts  # Scheduling skill
│   │   │   ├── lead-capture.ts # Lead gen skill
│   │   │   └── ...             # Other skills
│   │   ├── prompts/
│   │   │   ├── system.md       # Base system prompt
│   │   │   └── skills.md       # Skill definitions
│   │   └── tools/
│   │       ├── knowledge.ts    # Knowledge retrieval
│   │       └── actions.ts      # Execute actions
│   │
│   ├── utils/
│   │   ├── llm.ts         # Workers AI client
│   │   ├── vector.ts      # Vectorize operations
│   │   └── channels.ts    # Channel integrations
│   │
│   └── types/
│       └── index.ts       # TypeScript types
│
├── frontend/              # Cloudflare Pages
│   ├── index.html         # Landing
│   ├── dashboard/         # User dashboard
│   └── onboarding/        # Setup flow
│
├── migrations/            # D1 migrations
│   └── 001_initial.sql
│
├── wrangler.toml         # CF config
├── package.json
└── tsconfig.json
```

### Wrangler Configuration

```toml
name = "agentforge"
main = "src/workers/index.ts"
compatibility_date = "2025-04-06"

[ai]
binding = "AI"

[durable_objects]
bindings = [
  { name = "Orchestrator", class_name = "BusinessOrchestrator" }
]

[[d1.databases]]
binding = "DB"
database_name = "agentforge"
database_id = "..."

[[vectorize]]
binding = "KNOWLEDGE"
```

### The Orchestrator Agent

```typescript
import { AIChatAgent } from "agents/ai-chat-agent";
import { createWorkersAI } from "workers-ai-provider";
import { streamText, convertToModelMessages } from "ai";

interface Env {
  AI: Ai;
  DB: D1Database;
  KNOWLEDGE: Vectorize;
}

interface State {
  userId: string;
  enabledSkills: string[];
  conversationHistory: Array<{role: string, content: string}>;
  knowledgeIndexId: string;
}

export class BusinessOrchestrator extends AIChatAgent<Env, State> {
  initialState: State = {
    userId: "",
    enabledSkills: [],
    conversationHistory: [],
    knowledgeIndexId: "",
  };

  async onChatMessage(message: string) {
    const workersai = createWorkersAI({ binding: this.env.AI });
    
    // Build system prompt from enabled skills
    const systemPrompt = this.buildSystemPrompt(this.state.enabledSkills);
    
    // Retrieve relevant knowledge
    const knowledgeContext = await this.retrieveKnowledge(message);
    
    // Build messages
    const messages = [
      { role: "system", content: systemPrompt },
      ...knowledgeContext ? [{ role: "system", content: `Context: ${knowledgeContext}` }] : [],
      ...convertToModelMessages(this.messages),
      { role: "user", content: message },
    ];
    
    // Call Kimi K2.5
    const result = streamText({
      model: workersai("@cf/moonshotai/kimi-k2.5"),
      messages,
      maxTokens: 4096,
    });
    
    return result.toUIMessageStreamResponse();
  }

  private async retrieveKnowledge(query: string): Promise<string> {
    // Vectorize similarity search
    const results = await this.env.KNOWLEDGE.query({
      query,
      topK: 3,
    });
    
    return results.map(r => r.vector?.text || "").join("\n\n");
  }
}
```

---

## 7. Data Models

### D1 Schema

```sql
-- Users table
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT,
  created_at INTEGER NOT NULL,
  subscription_tier TEXT DEFAULT 'starter',
  subscription_status TEXT DEFAULT 'active',
  stripe_customer_id TEXT,
  api_key TEXT
);

-- Business configs
CREATE TABLE business_configs (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  
  -- Behavior
  response_style TEXT DEFAULT 'friendly',
  language TEXT DEFAULT 'en',
  working_hours TEXT DEFAULT '24/7',
  escalation_mode TEXT DEFAULT 'smart',
  
  -- Skills enabled
  skills JSON NOT NULL DEFAULT '[]',
  
  -- Channels
  whatsapp_number TEXT,
  telegram_bot_token TEXT,
  website_widget_enabled INTEGER DEFAULT 0,
  
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

-- Knowledge documents
CREATE TABLE knowledge_docs (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  doc_type TEXT DEFAULT 'text',
  
  -- Vectorize index ID
  vector_index_id TEXT,
  
  created_at INTEGER NOT NULL
);

-- Conversations (for analytics)
CREATE TABLE conversations (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  channel TEXT NOT NULL,
  message TEXT NOT NULL,
  response TEXT,
  skill_used TEXT,
  tokens_used INTEGER,
  created_at INTEGER NOT NULL
);
```

---

## 8. Channel Integrations

### WhatsApp (Meta Business)

```
User sends WhatsApp message
         ↓
Meta webhook → CF Worker
         ↓
Route to user's DO (Orchestrator)
         ↓
Process with enabled skills
         ↓
Respond via WhatsApp API
```

### Telegram

```
User sends to bot
         ↓
Telegram webhook → CF Worker
         ↓
Route to user's DO
         ↓
Process & respond
```

### Website Widget

```
User visits website
         ↓
Embed chat widget (static HTML)
         ↓
Connect via WebSocket to DO
         ↓
Real-time chat
```

---

## 9. Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **One DO per user** | Each business gets isolated instance, scale-to-zero when idle |
| **Kimi K2.5** | 256k context fits business docs, tool calling for actions |
| **Skills as config** | No code changes needed to add capabilities |
| **Magic link auth** | No passwords, simple for non-tech users |
| **SQLite in DO** | Built-in, fast, no external DB needed per user |
| **Vectorize for knowledge** | Semantic search over business documents |

---

## 10. Roadmap

### Phase 1: MVP (Weeks 1-4)

| Week | Focus | Deliverables |
|------|-------|--------------|
| 1 | Setup + Landing | CF project, landing page, user signup |
| 2 | Onboarding + Config | Skill selection, knowledge upload, behavior config |
| 3 | Orchestrator + AI | Implement DO, integrate Kimi K2.5, tool calling |
| 4 | WhatsApp + Stripe | Channel integration, payment flow |

### Phase 2: Growth (Weeks 5-8)

| Week | Focus | Deliverables |
|------|-------|--------------|
| 5 | Telegram + Widget | More channels |
| 6 | Analytics | Usage dashboard |
| 7 | Team features | Multi-user support |
| 8 | More Skills | Add remaining skills |

### Phase 3: Scale (Weeks 9-12)

| Week | Focus | Deliverables |
|------|-------|--------------|
| 9-10 | Enterprise | SSO, custom contracts |
| 11 | API | Developer integrations |
| 12 | Regional | SG, ID, TH launch |

---

## 11. Success Metrics

### Targets (12 months)

| Metric | Target |
|--------|--------|
| Customers | 2,000 |
| MRR | RM 500,000 |
| Daily Active Users | 5,000 |
| NPS | > 50 |

---

## 12. Dependencies

### NPM Packages

```json
{
  "dependencies": {
    "agents": "latest",
    "workers-ai-provider": "latest",
    "ai": "latest",
    "hono": "latest"
  },
  "devDependencies": {
    "@cloudflare/workers-types": "latest",
    "typescript": "latest",
    "wrangler": "latest"
  }
}
```

---

## 13. Next Steps

1. ✅ PRD Approved (v3.0)
2. Initialize CF project: `npm create cloudflare@latest`
3. Implement Sprint 1

---

**Document Status**: READY FOR DEVELOPMENT  
**Version**: 3.0  
**Last Updated**: April 6, 2026

---

*Changelog:*
- v1.0: Initial draft with multiple options
- v2.0: Focused on SMB owners, marketplace-first, tiered pricing
- v3.0: Skills + Single Orchestrator model, full CF stack, latest Agents SDK