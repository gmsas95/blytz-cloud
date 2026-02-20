'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  Squares2X2Icon,
  CommandLineIcon,
  Cog6ToothIcon,
  CreditCardIcon,
  ShoppingBagIcon,
  ArrowLeftOnRectangleIcon,
  ArrowTopRightOnSquareIcon
} from '@heroicons/react/24/outline'

const navItems = [
  {
    href: '/dashboard',
    label: 'Overview',
    icon: Squares2X2Icon
  },
  {
    href: '/dashboard/agents',
    label: 'My Agents',
    icon: CommandLineIcon
  },
  {
    href: '/dashboard/marketplace',
    label: 'Marketplace',
    icon: ShoppingBagIcon
  },
  {
    href: '/dashboard/billing',
    label: 'Billing',
    icon: CreditCardIcon
  },
  {
    href: '/dashboard/settings',
    label: 'Settings',
    icon: Cog6ToothIcon
  }
]

export function DashboardSidebar() {
  const pathname = usePathname()

  return (
    <aside className="fixed left-0 top-0 h-screen w-64 bg-amber-950/20 border-r border-amber-900/30 flex flex-col backdrop-blur-xl">
      {/* Logo */}
      <div className="p-6 border-b border-amber-900/30">
        <Link href="/" className="flex items-center gap-3">
          <div className="w-8 h-8 bg-gradient-to-br from-amber-500 to-orange-600 rounded-lg flex items-center justify-center">
            <CommandLineIcon className="w-5 h-5 text-white" />
          </div>
          <span className="text-xl font-bold text-white">Blytz</span>
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
              className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                isActive
                  ? 'bg-amber-500/10 text-amber-400 border-r-2 border-amber-500'
                  : 'text-amber-200/50 hover:text-amber-100 hover:bg-amber-500/5'
              }`}
            >
              <Icon className="w-5 h-5" />
              <span>{item.label}</span>
            </Link>
          )
        })}
      </nav>

      {/* User section */}
      <div className="p-4 border-t border-amber-900/30 space-y-4">
        {/* Subdomain link */}
        <div className="px-4">
          <p className="text-xs text-amber-600 mb-2">Your Agent</p>
          <a
            href="https://demo.blytz.cloud"
            target="_blank"
            className="flex items-center gap-2 text-sm text-amber-400 hover:text-amber-300"
          >
            demo.blytz.cloud
            <ArrowTopRightOnSquareIcon className="w-3 h-3" />
          </a>
        </div>

        {/* Logout */}
        <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-amber-200/50 hover:text-amber-100 hover:bg-amber-500/5 transition-all">
          <ArrowLeftOnRectangleIcon className="w-5 h-5" />
          <span>Sign Out</span>
        </button>
      </div>
    </aside>
  )
}
