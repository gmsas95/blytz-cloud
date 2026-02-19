'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  Bot,
  Settings,
  CreditCard,
  Store,
  LogOut,
  ExternalLink
} from 'lucide-react'

const navItems = [
  {
    href: '/dashboard',
    label: 'Overview',
    icon: LayoutDashboard
  },
  {
    href: '/dashboard/agents',
    label: 'My Agents',
    icon: Bot
  },
  {
    href: '/dashboard/marketplace',
    label: 'Marketplace',
    icon: Store
  },
  {
    href: '/dashboard/billing',
    label: 'Billing',
    icon: CreditCard
  },
  {
    href: '/dashboard/settings',
    label: 'Settings',
    icon: Settings
  }
]

export function DashboardSidebar() {
  const pathname = usePathname()

  return (
    <aside className="fixed left-0 top-0 h-screen w-64 bg-zinc-900/50 border-r border-zinc-800 flex flex-col">
      {/* Logo */}
      <div className="p-6 border-b border-zinc-800">
        <Link href="/" className="flex items-center gap-2">
          <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <Bot className="w-5 h-5 text-white" />
          </div>
          <span className="text-xl font-bold">Blytz</span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = pathname === item.href
          
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-all ${
                isActive
                  ? 'bg-blue-500/10 text-blue-400 border-r-2 border-blue-500'
                  : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
              }`}
            >
              <Icon className="w-5 h-5" />
              <span>{item.label}</span>
            </Link>
          )
        })}
      </nav>

      {/* User section */}
      <div className="p-4 border-t border-zinc-800 space-y-4">
        {/* Subdomain link */}
        <div className="px-4">
          <p className="text-xs text-zinc-500 mb-2">Your Agent</p>
          <a
            href="https://demo.blytz.cloud"
            target="_blank"
            className="flex items-center gap-2 text-sm text-blue-400 hover:text-blue-300"
          >
            demo.blytz.cloud
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>

        {/* Logout */}
        <button className="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800/50 transition-all">
          <LogOut className="w-5 h-5" />
          <span>Sign Out</span>
        </button>
      </div>
    </aside>
  )
}
