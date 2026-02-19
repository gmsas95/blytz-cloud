# Blytz Frontend

Modern React/Next.js frontend for the Blytz AI Assistant Platform.

## 🎨 Architecture

This frontend provides a modern PaaS-style dashboard similar to Vercel, Netlify, or Railway:

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Next.js)                      │
├─────────────────────────────────────────────────────────────┤
│  Landing Page          │  Dashboard (app.blytz.work)         │
│  - Signup form         │  - Overview                         │
│  - Marketing copy      │  - My Agents                        │
│  - Feature showcase    │  - Marketplace (coming soon)        │
│                        │  - Billing                          │
│                        │  - Settings                         │
└────────────────────────┴─────────────────────────────────────┘
         │                           │
         │                           │
         ▼                           ▼
   ┌──────────────┐         ┌──────────────┐
   │  Go Backend  │         │  Go Backend  │
   │  :8080       │         │  :8080       │
   └──────────────┘         └──────────────┘
```

## 🚀 Getting Started

### Prerequisites

- Node.js 18+
- Go backend running on port 8080

### Installation

```bash
cd frontend
npm install
```

### Development

```bash
# Run Next.js dev server (port 3000)
npm run dev

# The frontend proxies API requests to Go backend on :8080
```

### Build for Production

```bash
npm run build
npm start
```

## 📁 Project Structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── page.tsx              # Landing page
│   │   ├── layout.tsx            # Root layout
│   │   ├── globals.css           # Global styles
│   │   └── dashboard/            # Dashboard routes
│   │       ├── layout.tsx        # Dashboard layout with sidebar
│   │       ├── page.tsx          # Overview page
│   │       ├── agents/           # My Agents page
│   │       ├── marketplace/      # Agent Marketplace
│   │       ├── billing/          # Billing & subscription
│   │       └── settings/         # Agent settings
│   └── components/
│       └── dashboard-sidebar.tsx # Sidebar navigation
├── public/                        # Static assets
└── next.config.ts                # Next.js config with API proxy
```

## 🌐 URL Structure

### User-Facing URLs

- `app.blytz.work` - Main dashboard
- `app.blytz.work/dashboard` - Overview
- `app.blytz.work/dashboard/agents` - My Agents
- `app.blytz.work/dashboard/marketplace` - Agent Marketplace
- `app.blytz.work/dashboard/billing` - Billing
- `app.blytz.work/dashboard/settings` - Settings

### Tenant Subdomains

- `alice.blytz.work` → Alice's OpenClaw UI
- `bob.blytz.work` → Bob's OpenClaw UI

## 🔧 Configuration

### API Proxy

Next.js proxies all `/api/*` requests to the Go backend:

```typescript
// next.config.ts
async rewrites() {
  return [
    {
      source: '/api/:path*',
      destination: 'http://localhost:8080/api/:path*',
    },
  ];
}
```

### Environment Variables

Create `.env.local`:

```bash
# API URL for client-side requests
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 🎨 Design System

### Colors
- Background: `#000000` (black)
- Foreground: `#ffffff` (white)
- Muted: `#a1a1aa` (zinc-400)
- Border: `#27272a` (zinc-800)
- Primary: `#3b82f6` (blue-500)
- Success: `#22c55e` (green-500)

### Components

- **Cards**: `bg-zinc-900/50 border border-zinc-800 rounded-xl`
- **Buttons**: Gradient backgrounds with hover effects
- **Inputs**: Dark inputs with focus rings
- **Sidebar**: Fixed left sidebar with active states

## 📦 Features

### Implemented ✅

- [x] Modern dark theme UI
- [x] Landing page with signup form
- [x] Dashboard with sidebar navigation
- [x] Overview page with stats
- [x] My Agents page
- [x] Billing page
- [x] Settings page
- [x] Agent Marketplace (UI only)
- [x] Responsive design
- [x] API proxy to Go backend

### Coming Soon 🚧

- [ ] Agent Marketplace backend integration
- [ ] Real-time logs viewer
- [ ] Usage analytics charts
- [ ] Dark/light theme toggle
- [ ] Mobile app

## 🤝 Integration with Go Backend

The frontend expects these API endpoints:

```
POST /api/signup              - Create customer
GET  /api/status/:id          - Get customer status
POST /api/webhook/stripe      - Stripe webhooks
GET  /api/health              - Health check
```

## 📝 Development Notes

### Adding New Pages

1. Create page component in `src/app/dashboard/[page]/page.tsx`
2. Add navigation link in `src/components/dashboard-sidebar.tsx`
3. Update `next.config.ts` if needed

### Styling

- Use Tailwind CSS classes
- Follow existing color scheme
- Use `card-hover` class for interactive cards
- Use `gradient-text` for highlighted text

### Icons

Use Lucide React icons:

```tsx
import { Bot, Settings, CreditCard } from 'lucide-react'
```

## 📄 License

MIT
