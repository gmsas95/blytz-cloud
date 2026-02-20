# Multi-Agent Marketplace Architecture

## Concept

Platform provides **infrastructure-as-a-service** for personal AI agents.

**Key Principle: 1 Container Slot = 1 Agent**
- User pays for 1 container slot ($29/month)
- User chooses which agent runs in that slot
- Container runs a single agent instance
- Easy to switch agents (terminate old, deploy new)

**User Journey:**
1. Sign up → Gets 1 container slot
2. Choose Agent Framework (OpenClaw, Myrai, etc.) → Runs in their slot
3. Choose LLM Provider (OpenAI, Anthropic, Groq, etc.)
4. Configure (API keys, channels, etc.)
5. Deploy → Container spins up with chosen agent
6. Can switch agent later (swap what's in their slot)

## Supported Agents

### Tier 1: Fully Supported

| Agent | Language | Port | Resources | Status |
|-------|----------|------|-----------|--------|
| **OpenClaw** | Node.js | 18789 | 512MB / 0.25 CPU | ✅ Ready |
| **Myrai** | Go | 8080 | 512MB / 0.25 CPU | ✅ Ready |
| **Nanobot** | Python | 5000 | 512MB / 0.25 CPU | 🚧 Planned |
| **ZeptoClaw** | Node.js | 3000 | 256MB / 0.2 CPU | 🚧 Planned |
| **PicoClaw** | Node.js | 3000 | 256MB / 0.2 CPU | 🚧 Planned |

### Tier 2: Coming Soon
- Custom Docker images (bring your own agent)
- Community-contributed agents

## Database Schema

### agent_types table
```sql
CREATE TABLE agent_types (
    id TEXT PRIMARY KEY,           -- 'openclaw', 'myrai', 'nanobot'
    name TEXT NOT NULL,            -- 'OpenClaw'
    description TEXT,
    language TEXT,                 -- 'nodejs', 'go', 'python'
    base_image TEXT,               -- Docker base image
    internal_port INTEGER,         -- 18789, 8080, etc.
    health_endpoint TEXT,          -- '/health', '/api/health'
    min_memory TEXT,               -- '256M', '512M'
    min_cpu TEXT,                  -- '0.2', '0.25'
    config_template TEXT,          -- JSON template for config
    env_vars TEXT,                 -- JSON array of required env vars
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP
);
```

### llm_providers table
```sql
CREATE TABLE llm_providers (
    id TEXT PRIMARY KEY,           -- 'openai', 'anthropic', 'groq'
    name TEXT NOT NULL,            -- 'OpenAI'
    description TEXT,
    env_key TEXT,                  -- 'OPENAI_API_KEY'
    base_url TEXT,                 -- Optional custom endpoint
    is_active BOOLEAN DEFAULT true
);
```

### Updated customers table
```sql
ALTER TABLE customers ADD COLUMN agent_type_id TEXT;
ALTER TABLE customers ADD COLUMN llm_provider_id TEXT;
ALTER TABLE customers ADD COLUMN custom_config TEXT; -- JSON
```

## Architecture

### 1. Agent Registry
- Stores all supported agent types
- Version management
- Configuration schemas

### 2. LLM Provider Registry
- Available providers
- API key validation
- Custom endpoint support

### 3. Dynamic Compose Generator
- Template-based
- Per-agent configuration
- Environment variable injection

### 4. Port Management
- Different internal ports per agent
- External port allocation
- Health check endpoints

## API Endpoints

### Marketplace
```
GET  /api/marketplace/agents          # List available agents
GET  /api/marketplace/llm-providers   # List LLM providers
GET  /api/marketplace/stacks          # Popular combinations
```

### Customer with Agent Selection
```
POST /api/signup
{
  "email": "user@example.com",
  "agent_type_id": "openclaw",
  "llm_provider_id": "anthropic",
  "llm_api_key": "sk-ant-...",
  "telegram_bot_token": "...",
  "config": { /* agent-specific config */ }
}
```

## Docker Compose Templates

### OpenClaw Template
```yaml
version: '3.8'
services:
  agent:
    image: node:22-bookworm
    command: >
      sh -c "npm install -g openclaw@latest &&
             openclaw gateway --port {{.InternalPort}} --bind lan"
    ports:
      - "{{.ExternalPort}}:{{.InternalPort}}"
      - "{{.ExternalPortBridge}}:{{.InternalPortBridge}}"
    environment:
      - HOME=/home/node
      - {{.LLMEnvKey}}={{.LLMKey}}
    volumes:
      - ./config:/home/node/.openclaw
    user: "1000:1000"
```

### Myrai Template
```yaml
version: '3.8'
services:
  agent:
    image: ghcr.io/gmsas95/myrai:latest
    command: ["myrai", "server", "--port", "{{.InternalPort}}"]
    ports:
      - "{{.ExternalPort}}:{{.InternalPort}}"
    environment:
      - {{.LLMEnvKey}}={{.LLMKey}}
      - MYRAI_GATEWAY_TOKEN={{.GatewayToken}}
    volumes:
      - ./data:/app/data
```

## Configuration Flow

1. **User selects agent** → We load agent template + config schema
2. **User selects LLM** → We show appropriate API key field
3. **User configures** → We validate config against schema
4. **Deploy** → We render template with values + spin up container

## Pricing Model

**Base:** $29/month for container slot (512MB / 0.25 CPU)

**Add-ons:**
- +$10/month: Additional memory (1GB total)
- +$10/month: Additional CPU (0.5 total)
- +$5/month: Priority support

**User pays for:**
- Container resources (to us)
- LLM API usage (to OpenAI/Anthropic/etc.)

## Benefits

**For Users:**
- Choose best agent for their needs
- Use their preferred LLM
- One-click deployment
- Easy switching between agents

**For Platform:**
- Not tied to one agent's success
- Marketplace dynamics (popular agents rise)
- Easier to add new agents
- Focus on infrastructure excellence

## Implementation Phases

### Phase 1: Foundation (Week 1)
- Database schema updates
- Agent registry
- LLM provider registry
- Template system

### Phase 2: OpenClaw + Myrai (Week 2)
- Support 2 agents
- Dynamic compose generation
- Port management per agent
- Basic marketplace UI

### Phase 3: Additional Agents (Week 3-4)
- Nanobot
- ZeptoClaw
- PicoClaw
- Custom agent support

### Phase 4: Polish (Week 5)
- Agent switching
- Migration tools
- Analytics
- Documentation
