'use client'

import { useState } from 'react'
import Link from 'next/link'
import { motion } from 'framer-motion'
import { 
  Zap, 
  Activity, 
  Globe, 
  MessageSquare,
  TrendingUp,
  Clock,
  ExternalLink,
  MoreHorizontal,
  Play,
  Pause,
  RotateCcw,
  Terminal,
  FileText,
  Settings,
  ChevronRight
} from 'lucide-react'

interface Agent {
  id: string
  name: string
  status: 'running' | 'stopped' | 'error'
  domain: string
  messagesToday: number
  messagesTotal: number
  lastActive: string
  uptime: string
}

interface ActivityItem {
  id: string
  action: string
  timestamp: string
  type: 'deploy' | 'message' | 'config' | 'error'
}

const mockAgent: Agent = {
  id: '1',
  name: 'Business Assistant',
  status: 'running',
  domain: 'assistant.blytz.cloud',
  messagesToday: 47,
  messagesTotal: 1234,
  lastActive: '2 minutes ago',
  uptime: '99.9%'
}

const recentActivity: ActivityItem[] = [
  { id: '1', action: 'Agent responded to 3 messages', timestamp: '2 min ago', type: 'message' },
  { id: '2', action: 'Configuration updated', timestamp: '1 hour ago', type: 'config' },
  { id: '3', action: 'Successfully deployed', timestamp: '3 hours ago', type: 'deploy' },
  { id: '4', action: 'Connected to Telegram', timestamp: '5 hours ago', type: 'config' },
]

const getActivityIcon = (type: ActivityItem['type']) => {
  switch (type) {
    case 'deploy': return <Zap className="w-4 h-4" />
    case 'message': return <MessageSquare className="w-4 h-4" />
    case 'config': return <Settings className="w-4 h-4" />
    case 'error': return <Activity className="w-4 h-4" />
  }
}

const getActivityColor = (type: ActivityItem['type']) => {
  switch (type) {
    case 'deploy': return 'text-emerald-400 bg-emerald-400/10 border-emerald-400/20'
    case 'message': return 'text-blue-400 bg-blue-400/10 border-blue-400/20'
    case 'config': return 'text-amber-400 bg-amber-400/10 border-amber-400/20'
    case 'error': return 'text-red-400 bg-red-400/10 border-red-400/20'
  }
}

export default function DashboardPage() {
  const [agent, setAgent] = useState<Agent>(mockAgent)
  const [isLoading, setIsLoading] = useState(false)

  const toggleAgent = () => {
    setIsLoading(true)
    setTimeout(() => {
      setAgent(prev => ({
        ...prev,
        status: prev.status === 'running' ? 'stopped' : 'running'
      }))
      setIsLoading(false)
    }, 800)
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header className="mb-8">
        <div className="flex items-end justify-between">
          <div>
            <h1 className="text-3xl font-bold text-white mb-1">Dashboard</h1>
            <p className="text-white/40">Welcome back. Your agent is {agent.status}.</p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/dashboard/agents"
              className="flex items-center gap-2 px-4 py-2 text-sm text-white/60 hover:text-white transition-colors"
            >
              View All Agents
              <ChevronRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </header>

      {/* Main Agent Card */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative mb-8"
      >
        {/* Glow effect */}
        <div className="absolute -inset-1 bg-gradient-to-r from-amber-500/20 via-orange-500/20 to-amber-500/20 rounded-2xl blur-xl opacity-50" />
        
        <div className="relative bg-gradient-to-br from-white/[0.08] to-white/[0.02] backdrop-blur-xl border border-white/[0.1] rounded-2xl p-8 overflow-hidden">
          {/* Background pattern */}
          <div className="absolute inset-0 opacity-30">
            <div className="absolute top-0 right-0 w-96 h-96 bg-amber-500/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
            <div className="absolute bottom-0 left-0 w-64 h-64 bg-orange-500/10 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2" />
          </div>

          <div className="relative flex flex-col lg:flex-row lg:items-center justify-between gap-8">
            {/* Agent Info */}
            <div className="flex items-center gap-6">
              {/* Status indicator */}
              <div className="relative">
                <div className={`w-16 h-16 rounded-2xl flex items-center justify-center ${
                  agent.status === 'running' 
                    ? 'bg-emerald-500/20 border border-emerald-500/30' 
                    : 'bg-amber-500/20 border border-amber-500/30'
                }`}>
                  <Terminal className={`w-8 h-8 ${
                    agent.status === 'running' ? 'text-emerald-400' : 'text-amber-400'
                  }`} />
                </div>
                {agent.status === 'running' && (
                  <span className="absolute -top-1 -right-1 w-4 h-4 bg-emerald-500 rounded-full border-2 border-black animate-pulse" />
                )}
              </div>

              <div>
                <div className="flex items-center gap-3 mb-1">
                  <h2 className="text-2xl font-bold text-white">{agent.name}</h2>
                  <span className={`px-2.5 py-0.5 text-xs font-medium rounded-full ${
                    agent.status === 'running'
                      ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                      : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                  }`}>
                    {agent.status === 'running' ? 'Live' : 'Stopped'}
                  </span>
                </div>
                <div className="flex items-center gap-4 text-sm text-white/50">
                  <a 
                    href={`https://${agent.domain}`}
                    target="_blank"
                    className="flex items-center gap-1.5 hover:text-white transition-colors"
                  >
                    <Globe className="w-4 h-4" />
                    {agent.domain}
                    <ExternalLink className="w-3 h-3" />
                  </a>
                  <span className="text-white/20">•</span>
                  <span>Uptime {agent.uptime}</span>
                </div>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="flex items-center gap-3">
              <button
                onClick={toggleAgent}
                disabled={isLoading}
                className={`flex items-center gap-2 px-6 py-3 rounded-xl font-medium transition-all ${
                  agent.status === 'running'
                    ? 'bg-white/5 hover:bg-white/10 text-white border border-white/10'
                    : 'bg-emerald-500 hover:bg-emerald-400 text-black'
                } disabled:opacity-50 disabled:cursor-not-allowed`}
              >
                {isLoading ? (
                  <div className="w-5 h-5 border-2 border-current border-t-transparent rounded-full animate-spin" />
                ) : agent.status === 'running' ? (
                  <>
                    <Pause className="w-5 h-5" />
                    Stop Agent
                  </>
                ) : (
                  <>
                    <Play className="w-5 h-5" />
                    Start Agent
                  </>
                )}
              </button>

              <Link
                href="/dashboard/settings"
                className="p-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-white/60 hover:text-white transition-all"
              >
                <Settings className="w-5 h-5" />
              </Link>
            </div>
          </div>

          {/* Stats Row */}
          <div className="relative mt-8 pt-8 border-t border-white/[0.06]">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
              <div>
                <p className="text-white/40 text-sm mb-1">Messages Today</p>
                <p className="text-3xl font-bold text-white">{agent.messagesToday}</p>
                <div className="flex items-center gap-1 mt-1 text-emerald-400 text-xs">
                  <TrendingUp className="w-3 h-3" />
                  <span>+12% from yesterday</span>
                </div>
              </div>

              <div>
                <p className="text-white/40 text-sm mb-1">Total Messages</p>
                <p className="text-3xl font-bold text-white">{agent.messagesTotal.toLocaleString()}</p>
                <p className="text-white/30 text-xs mt-1">All time</p>
              </div>

              <div>
                <p className="text-white/40 text-sm mb-1">Last Active</p>
                <p className="text-3xl font-bold text-white">{agent.lastActive}</p>
                <p className="text-white/30 text-xs mt-1">Processing requests</p>
              </div>

              <div>
                <p className="text-white/40 text-sm mb-1">Response Time</p>
                <p className="text-3xl font-bold text-white">1.2s</p>
                <p className="text-white/30 text-xs mt-1">Average</p>
              </div>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Bottom Grid */}
      <div className="grid lg:grid-cols-3 gap-6">
        {/* Recent Activity */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="lg:col-span-2"
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-white">Recent Activity</h3>
            <button className="text-sm text-white/40 hover:text-white transition-colors">
              View All
            </button>
          </div>

          <div className="space-y-3">
            {recentActivity.map((item, index) => (
              <motion.div
                key={item.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.1 + index * 0.05 }}
                className="flex items-center gap-4 p-4 bg-white/[0.03] hover:bg-white/[0.05] border border-white/[0.06] rounded-xl transition-colors group"
              >
                <div className={`w-10 h-10 rounded-xl flex items-center justify-center border ${getActivityColor(item.type)}`}>
                  {getActivityIcon(item.type)}
                </div>
                <div className="flex-1">
                  <p className="text-white font-medium">{item.action}</p>
                  <p className="text-white/40 text-sm">{item.timestamp}</p>
                </div>
                <button className="opacity-0 group-hover:opacity-100 p-2 text-white/40 hover:text-white transition-all">
                  <MoreHorizontal className="w-5 h-5" />
                </button>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Quick Actions */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          <h3 className="text-lg font-semibold text-white mb-4">Quick Actions</h3>
          
          <div className="space-y-3">
            <Link
              href="/dashboard/settings"
              className="flex items-center gap-4 p-4 bg-white/[0.03] hover:bg-white/[0.05] border border-white/[0.06] rounded-xl transition-all group"
            >
              <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center">
                <Settings className="w-5 h-5 text-amber-400" />
              </div>
              <div className="flex-1">
                <p className="text-white font-medium">Configure Agent</p>
                <p className="text-white/40 text-sm">Update settings & prompts</p>
              </div>
              <ChevronRight className="w-5 h-5 text-white/20 group-hover:text-white/60 transition-colors" />
            </Link>

            <a
              href={`https://${agent.domain}`}
              target="_blank"
              className="flex items-center gap-4 p-4 bg-white/[0.03] hover:bg-white/[0.05] border border-white/[0.06] rounded-xl transition-all group"
            >
              <div className="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center">
                <ExternalLink className="w-5 h-5 text-blue-400" />
              </div>
              <div className="flex-1">
                <p className="text-white font-medium">Open Agent UI</p>
                <p className="text-white/40 text-sm">View your public agent</p>
              </div>
              <ChevronRight className="w-5 h-5 text-white/20 group-hover:text-white/60 transition-colors" />
            </a>

            <Link
              href="/dashboard/agents"
              className="flex items-center gap-4 p-4 bg-white/[0.03] hover:bg-white/[0.05] border border-white/[0.06] rounded-xl transition-all group"
            >
              <div className="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center">
                <RotateCcw className="w-5 h-5 text-purple-400" />
              </div>
              <div className="flex-1">
                <p className="text-white font-medium">View Logs</p>
                <p className="text-white/40 text-sm">Check agent activity</p>
              </div>
              <ChevronRight className="w-5 h-5 text-white/20 group-hover:text-white/60 transition-colors" />
            </Link>

            <Link
              href="/dashboard/billing"
              className="flex items-center gap-4 p-4 bg-white/[0.03] hover:bg-white/[0.05] border border-white/[0.06] rounded-xl transition-all group"
            >
              <div className="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center">
                <FileText className="w-5 h-5 text-emerald-400" />
              </div>
              <div className="flex-1">
                <p className="text-white font-medium">Billing</p>
                <p className="text-white/40 text-sm">Manage subscription</p>
              </div>
              <ChevronRight className="w-5 h-5 text-white/20 group-hover:text-white/60 transition-colors" />
            </Link>
          </div>

          {/* System Status */}
          <div className="mt-6 p-4 bg-gradient-to-br from-emerald-500/10 to-transparent border border-emerald-500/20 rounded-xl">
            <div className="flex items-center gap-3 mb-2">
              <div className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
              <span className="text-emerald-400 font-medium">All Systems Operational</span>
            </div>
            <p className="text-white/40 text-sm">No incidents reported in the last 24 hours.</p>
          </div>
        </motion.div>
      </div>
    </div>
  )
}
