'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { motion, AnimatePresence } from 'framer-motion'
import {
  LayoutDashboard,
  Bot,
  Settings,
  CreditCard,
  Store,
  LogOut,
  ExternalLink,
  ChevronRight,
  Menu,
  X
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

interface DashboardSidebarProps {
  isMobileMenuOpen: boolean
  setIsMobileMenuOpen: (open: boolean) => void
}

export function DashboardSidebar({ isMobileMenuOpen, setIsMobileMenuOpen }: DashboardSidebarProps) {
  const pathname = usePathname()

  const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen)

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        onClick={toggleMobileMenu}
        className="fixed top-4 right-4 z-[60] lg:hidden p-3 bg-black/80 backdrop-blur-xl border border-white/[0.1] rounded-xl text-white"
        aria-label="Toggle menu"
      >
        {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
      </button>

      {/* Sidebar */}
      <aside className={`
        fixed left-0 top-0 h-screen w-72 flex flex-col z-50
        transform transition-transform duration-300 ease-in-out
        lg:translate-x-0
        ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
      `}>
        {/* Background with blur */}
        <div className="absolute inset-0 bg-black/95 lg:bg-black/80 backdrop-blur-2xl border-r border-white/[0.06]" />
        
        {/* Ambient glow */}
        <div className="absolute top-0 left-0 w-full h-32 bg-gradient-to-b from-amber-500/10 to-transparent pointer-events-none" />

        <div className="relative flex flex-col h-full">
          {/* Logo */}
          <div className="p-6 pt-20 lg:pt-6">
            <Link href="/" className="flex items-center gap-3 group">
              <div className="relative w-10 h-10">
                <div className="absolute inset-0 bg-gradient-to-br from-amber-500 to-orange-600 rounded-xl blur-md opacity-50 group-hover:opacity-75 transition-opacity" />
                <div className="relative w-10 h-10 bg-gradient-to-br from-amber-500 to-orange-600 rounded-xl flex items-center justify-center">
                  <svg className="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path strokeLinecap="round" strokeLinejoin="round" d="m8 9 3 3-3 3m5 0h3M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
                  </svg>
                </div>
              </div>
              <span className="text-xl font-bold text-white">Blytz</span>
            </Link>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-4 space-y-1">
            {navItems.map((item) => {
              const Icon = item.icon
              const isActive = pathname === item.href
              
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className="relative flex items-center gap-3 px-4 py-3 rounded-xl transition-all group"
                >
                  {/* Active background */}
                  {isActive && (
                    <motion.div
                      layoutId="activeNav"
                      className="absolute inset-0 bg-white/[0.06] border border-white/[0.1] rounded-xl"
                      transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                    />
                  )}
                  
                  {/* Active indicator line */}
                  {isActive && (
                    <motion.div
                      layoutId="activeIndicator"
                      className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-8 bg-gradient-to-b from-amber-500 to-orange-500 rounded-r-full"
                      transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                    />
                  )}
                  
                  <span className={`relative z-10 ${isActive ? 'text-amber-400' : 'text-white/50 group-hover:text-white/80'}`}>
                    <Icon className="w-5 h-5" />
                  </span>
                  <span className={`relative z-10 font-medium ${isActive ? 'text-white' : 'text-white/50 group-hover:text-white/80'}`}>
                    {item.label}
                  </span>
                  
                  {isActive && (
                    <ChevronRight className="relative z-10 w-4 h-4 text-white/30 ml-auto" />
                  )}
                </Link>
              )
            })}
          </nav>

          {/* Bottom section */}
          <div className="p-4 space-y-3">
            {/* Agent quick link */}
            <div className="p-4 bg-white/[0.03] border border-white/[0.06] rounded-xl">
              <p className="text-xs text-white/40 mb-2">Your Agent</p>
              <a
                href="https://demo.blytz.cloud"
                target="_blank"
                className="flex items-center gap-2 text-sm text-amber-400 hover:text-amber-300 transition-colors"
              >
                <span className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                demo.blytz.cloud
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>

            {/* Sign out */}
            <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-white/40 hover:text-white hover:bg-white/[0.05] transition-all"
            >
              <LogOut className="w-5 h-5" />
              <span className="font-medium">Sign Out</span>
            </button>
          </div>
        </div>
      </aside>
    </>
  )
}