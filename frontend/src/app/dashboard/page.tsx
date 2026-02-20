'use client'

import { 
  Activity, 
  Bot, 
  Globe, 
  MessageSquare,
  ArrowUpRight,
  CheckCircle2,
  XCircle,
  Loader2
} from 'lucide-react'

export default function DashboardPage() {
  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Overview</h1>
        <p className="text-zinc-400">Manage and monitor your AI assistant</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-blue-500/10 rounded-lg flex items-center justify-center">
              <Bot className="w-5 h-5 text-blue-500" />
            </div>
            <span className="px-2 py-1 bg-amber-500/10 text-green-400 text-xs rounded-full flex items-center gap-1">
              <CheckCircle2 className="w-3 h-3" />
              Active
            </span>
          </div>
          <p className="text-zinc-400 text-sm mb-1">Status</p>
          <p className="text-2xl font-semibold">Running</p>
        </div>

        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-purple-500/10 rounded-lg flex items-center justify-center">
              <Globe className="w-5 h-5 text-purple-500" />
            </div>
            <a
              href="https://demo.blytz.cloud"
              target="_blank"
              className="text-zinc-400 hover:text-white"
            >
              <ArrowUpRight className="w-5 h-5" />
            </a>
          </div>
          <p className="text-zinc-400 text-sm mb-1">Domain</p>
          <p className="text-2xl font-semibold truncate">demo.blytz.cloud</p>
        </div>

        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-amber-500/10 rounded-lg flex items-center justify-center">
              <MessageSquare className="w-5 h-5 text-amber-500" />
            </div>
          </div>
          <p className="text-zinc-400 text-sm mb-1">Messages</p>
          <p className="text-2xl font-semibold">1,234</p>
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Agent Details */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold">Your Agent</h2>
              <button className="px-4 py-2 bg-blue-500 hover:bg-blue-600 rounded-lg text-sm font-medium transition-colors">
                Configure
              </button>
            </div>

            <div className="space-y-4">
              <div className="flex items-center gap-4 p-4 bg-zinc-800/50 rounded-lg">
                <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                  <Bot className="w-6 h-6 text-white" />
                </div>
                <div>
                  <p className="font-medium">My AI Assistant</p>
                  <p className="text-sm text-zinc-400">OpenClaw Gateway</p>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 bg-zinc-800/50 rounded-lg">
                  <p className="text-sm text-zinc-400 mb-1">Framework</p>
                  <p className="font-medium">OpenClaw</p>
                </div>
                <div className="p-4 bg-zinc-800/50 rounded-lg">
                  <p className="text-sm text-zinc-400 mb-1">Version</p>
                  <p className="font-medium">v1.0.0</p>
                </div>
                <div className="p-4 bg-zinc-800/50 rounded-lg">
                  <p className="text-sm text-zinc-400 mb-1">Region</p>
                  <p className="font-medium">US East</p>
                </div>
                <div className="p-4 bg-zinc-800/50 rounded-lg">
                  <p className="text-sm text-zinc-400 mb-1">Port</p>
                  <p className="font-medium">30001</p>
                </div>
              </div>
            </div>
          </div>

          {/* Recent Activity */}
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
            <h2 className="text-xl font-semibold mb-6">Recent Activity</h2>
            
            <div className="space-y-4">
              {[
                { action: 'Agent deployed', time: '2 minutes ago', status: 'success' },
                { action: 'Configuration updated', time: '1 hour ago', status: 'success' },
                { action: 'Payment processed', time: '1 day ago', status: 'success' },
              ].map((activity, i) => (
                <div key={i} className="flex items-center gap-4 p-4 bg-zinc-800/50 rounded-lg">
                  <div className="w-8 h-8 bg-blue-500/10 rounded-full flex items-center justify-center">
                    <Activity className="w-4 h-4 text-blue-500" />
                  </div>
                  <div className="flex-1">
                    <p className="font-medium">{activity.action}</p>
                    <p className="text-sm text-zinc-400">{activity.time}</p>
                  </div>
                  {activity.status === 'success' ? (
                    <CheckCircle2 className="w-5 h-5 text-amber-500" />
                  ) : (
                    <XCircle className="w-5 h-5 text-red-500" />
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Sidebar Info */}
        <div className="space-y-6">
          {/* Quick Actions */}
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
            <h2 className="text-lg font-semibold mb-4">Quick Actions</h2>
            
            <div className="space-y-3">
              <button className="w-full px-4 py-3 bg-zinc-800 hover:bg-zinc-700 rounded-lg text-left transition-colors flex items-center gap-3">
                <ArrowUpRight className="w-4 h-4" />
                Open Agent UI
              </button>
              
              <button className="w-full px-4 py-3 bg-zinc-800 hover:bg-zinc-700 rounded-lg text-left transition-colors flex items-center gap-3">
                <MessageSquare className="w-4 h-4" />
                View Telegram Bot
              </button>
              
              <button className="w-full px-4 py-3 bg-zinc-800 hover:bg-zinc-700 rounded-lg text-left transition-colors flex items-center gap-3">
                <Activity className="w-4 h-4" />
                View Logs
              </button>
            </div>
          </div>

          {/* Billing Info */}
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
            <h2 className="text-lg font-semibold mb-4">Billing</h2>
            
            <div className="space-y-4">
              <div>
                <p className="text-sm text-zinc-400">Current Plan</p>
                <p className="text-xl font-semibold">Pro</p>
              </div>
              
              <div>
                <p className="text-sm text-zinc-400">Monthly Cost</p>
                <p className="text-xl font-semibold">$29.00</p>
              </div>
              
              <div className="p-3 bg-amber-500/10 border border-amber-500/20 rounded-lg">
                <p className="text-sm text-green-400">✓ Payment up to date</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
