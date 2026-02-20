'use client'

import { Bot, Plus, ExternalLink, Settings } from 'lucide-react'

const agents = [
  {
    id: 1,
    name: 'My AI Assistant',
    domain: 'demo.blytz.cloud',
    status: 'active',
    framework: 'OpenClaw',
    lastDeployed: '2 minutes ago'
  }
]

export default function AgentsPage() {
  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">My Agents</h1>
          <p className="text-zinc-400">Manage your deployed AI assistants</p>
        </div>
        
        <button className="px-4 py-2 bg-blue-500 hover:bg-blue-600 rounded-lg font-medium flex items-center gap-2 transition-colors">
          <Plus className="w-5 h-5" />
          New Agent
        </button>
      </div>

      <div className="grid gap-4">
        {agents.map((agent) => (
          <div
            key={agent.id}
            className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6 flex items-center justify-between"
          >
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <Bot className="w-6 h-6 text-white" />
              </div>
              
              <div>
                <h3 className="font-semibold text-lg">{agent.name}</h3>
                <div className="flex items-center gap-4 text-sm text-zinc-400">
                  <span className="flex items-center gap-1">
                    <span className="w-2 h-2 bg-amber-500 rounded-full"></span>
                    {agent.status}
                  </span>
                  <span>{agent.framework}</span>
                  <span>Last deployed {agent.lastDeployed}</span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <a
                href={`https://${agent.domain}`}
                target="_blank"
                className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg text-sm font-medium flex items-center gap-2 transition-colors"
              >
                <ExternalLink className="w-4 h-4" />
                Visit
              </a>
              
              <button className="p-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg transition-colors">
                <Settings className="w-5 h-5" />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
