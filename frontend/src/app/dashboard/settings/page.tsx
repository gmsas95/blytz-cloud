'use client'

import { useState } from 'react'
import { Save, Bot, Globe, Key } from 'lucide-react'

export default function SettingsPage() {
  const [formData, setFormData] = useState({
    assistantName: 'My AI Assistant',
    telegramToken: '',
    instructions: 'I am a freelance developer. I need help with client proposals, research, and scheduling...'
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    console.log('Saving settings:', formData)
    // TODO: Call API
  }

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Settings</h1>
        <p className="text-zinc-400">Manage your agent configuration</p>
      </div>

      <form onSubmit={handleSubmit} className="max-w-2xl space-y-6">
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 bg-blue-500/10 rounded-lg flex items-center justify-center">
              <Bot className="w-5 h-5 text-blue-500" />
            </div>
            <h2 className="text-xl font-semibold">Agent Configuration</h2>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">Assistant Name</label>
              <input
                type="text"
                value={formData.assistantName}
                onChange={(e) => setFormData({...formData, assistantName: e.target.value})}
                className="input-dark w-full px-4 py-3 rounded-lg text-white"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">
                Telegram Bot Token
                <a href="https://t.me/botfather" target="_blank" className="text-blue-400 hover:text-blue-300 ml-2 text-xs">
                  Get from @BotFather →
                </a>
              </label>
              <input
                type="password"
                value={formData.telegramToken}
                onChange={(e) => setFormData({...formData, telegramToken: e.target.value})}
                className="input-dark w-full px-4 py-3 rounded-lg text-white font-mono text-sm"
                placeholder="••••••••••••••••"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Custom Instructions</label>
              <textarea
                rows={6}
                value={formData.instructions}
                onChange={(e) => setFormData({...formData, instructions: e.target.value})}
                className="input-dark w-full px-4 py-3 rounded-lg text-white resize-none"
              />
              <p className="mt-2 text-sm text-zinc-400">
                These instructions help your assistant understand your needs and preferences.
              </p>
            </div>
          </div>
        </div>

        <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 bg-purple-500/10 rounded-lg flex items-center justify-center">
              <Globe className="w-5 h-5 text-purple-500" />
            </div>
            <h2 className="text-xl font-semibold">Domain Settings</h2>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-zinc-800/50 rounded-lg">
              <div>
                <p className="font-medium">Current Domain</p>
                <p className="text-zinc-400">demo.blytz.cloud</p>
              </div>
              <span className="px-3 py-1 bg-amber-500/10 text-amber-400 text-sm rounded-full">
                Active
              </span>
            </div>
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            className="btn-primary px-6 py-3 rounded-lg font-medium flex items-center gap-2"
          >
            <Save className="w-5 h-5" />
            Save Changes
          </button>
        </div>
      </form>
    </div>
  )
}
