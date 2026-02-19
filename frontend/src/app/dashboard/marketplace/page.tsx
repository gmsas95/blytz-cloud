'use client'

import { Bot, ArrowRight, Sparkles, Code, MessageSquare, Palette } from 'lucide-react'

const agents = [
  {
    id: 1,
    name: 'Personal Assistant',
    description: 'General purpose AI assistant for daily tasks, scheduling, and research',
    icon: Bot,
    color: 'blue',
    features: ['Task Management', 'Scheduling', 'Research', 'Email Drafting'],
    popular: true
  },
  {
    id: 2,
    name: 'Code Assistant',
    description: 'Specialized in coding, debugging, and technical documentation',
    icon: Code,
    color: 'purple',
    features: ['Code Review', 'Debugging', 'Documentation', 'Best Practices'],
    popular: false
  },
  {
    id: 3,
    name: 'Content Creator',
    description: 'Helps with writing, editing, and content strategy',
    icon: MessageSquare,
    color: 'pink',
    features: ['Writing', 'Editing', 'SEO', 'Strategy'],
    popular: false
  },
  {
    id: 4,
    name: 'Design Assistant',
    description: 'Assists with UI/UX design, feedback, and creative direction',
    icon: Palette,
    color: 'orange',
    features: ['UI/UX Feedback', 'Design Systems', 'Typography', 'Color Theory'],
    popular: false
  }
]

export default function MarketplacePage() {
  return (
    <div className="p-8">
      <div className="mb-8">
        <div className="flex items-center gap-2 mb-2">
          <Sparkles className="w-5 h-5 text-yellow-500" />
          <span className="text-yellow-500 font-medium">Coming Soon</span>
        </div>
        <h1 className="text-3xl font-bold mb-2">Agent Marketplace</h1>
        <p className="text-zinc-400">Choose from pre-configured AI assistants for your specific needs</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {agents.map((agent) => {
          const Icon = agent.icon
          const colorClasses = {
            blue: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
            purple: 'bg-purple-500/10 text-purple-500 border-purple-500/20',
            pink: 'bg-pink-500/10 text-pink-500 border-pink-500/20',
            orange: 'bg-orange-500/10 text-orange-500 border-orange-500/20'
          }

          return (
            <div
              key={agent.id}
              className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6 card-hover"
            >
              <div className="flex items-start justify-between mb-4">
                <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${colorClasses[agent.color as keyof typeof colorClasses]}`}>
                  <Icon className="w-6 h-6" />
                </div>
                
                {agent.popular && (
                  <span className="px-2 py-1 bg-yellow-500/10 text-yellow-500 text-xs rounded-full">
                    Popular
                  </span>
                )}
              </div>

              <h3 className="text-xl font-semibold mb-2">{agent.name}</h3>
              <p className="text-zinc-400 mb-4">{agent.description}</p>

              <div className="flex flex-wrap gap-2 mb-6">
                {agent.features.map((feature) => (
                  <span
                    key={feature}
                    className="px-2 py-1 bg-zinc-800 text-zinc-300 text-xs rounded-md"
                  >
                    {feature}
                  </span>
                ))}
              </div>

              <button className="w-full px-4 py-3 bg-blue-500 hover:bg-blue-600 rounded-lg font-medium flex items-center justify-center gap-2 transition-colors">
                Deploy Agent
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          )
        })}
      </div>

      <div className="mt-12 p-8 bg-gradient-to-r from-blue-500/10 to-purple-500/10 border border-blue-500/20 rounded-xl text-center">
        <h2 className="text-2xl font-bold mb-3">Want a custom agent?</h2>
        <p className="text-zinc-400 mb-6 max-w-lg mx-auto">
          We can help you build a specialized AI assistant tailored to your specific workflow and requirements.
        </p>
        <button className="px-6 py-3 bg-white text-black rounded-lg font-medium hover:bg-zinc-200 transition-colors">
          Contact Sales
        </button>
      </div>
    </div>
  )
}
