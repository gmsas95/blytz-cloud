'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { 
  Bot, 
  Zap, 
  Shield, 
  Globe, 
  ChevronRight,
  Check,
  Loader2
} from 'lucide-react'

export default function LandingPage() {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    email: '',
    assistantName: '',
    telegramToken: '',
    instructions: ''
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    
    // TODO: Call API
    console.log('Signup:', formData)
    
    // Simulate API call
    setTimeout(() => {
      setIsLoading(false)
      router.push('/dashboard')
    }, 1500)
  }

  return (
    <div className="min-h-screen bg-black text-white overflow-hidden">
      {/* Background gradient */}
      <div className="fixed inset-0 bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(120,119,198,0.3),rgba(255,255,255,0))] pointer-events-none" />
      
      {/* Navigation */}
      <nav className="relative z-50 flex items-center justify-between px-6 py-4 max-w-7xl mx-auto">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <Bot className="w-5 h-5 text-white" />
          </div>
          <span className="text-xl font-bold">Blytz</span>
        </div>
        
        <div className="hidden md:flex items-center gap-8 text-sm text-zinc-400">
          <a href="#features" className="hover:text-white transition-colors">Features</a>
          <a href="#pricing" className="hover:text-white transition-colors">Pricing</a>
          <a href="#docs" className="hover:text-white transition-colors">Docs</a>
        </div>
        
        <button className="px-4 py-2 text-sm bg-white/10 hover:bg-white/20 rounded-lg transition-colors">
          Sign In
        </button>
      </nav>

      {/* Hero Section */}
      <main className="relative z-10 max-w-7xl mx-auto px-6 pt-20 pb-32">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          {/* Left column - Copy */}
          <div className="space-y-8">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-sm">
              <Zap className="w-4 h-4" />
              <span>Now with Agent Marketplace</span>
            </div>
            
            <h1 className="text-5xl md:text-7xl font-bold leading-tight">
              Deploy your
              <br />
              <span className="gradient-text">AI Assistant</span>
              <br />
              in seconds
            </h1>
            
            <p className="text-xl text-zinc-400 max-w-lg">
              The fastest way to deploy personalized AI assistants. Set up your OpenClaw 
              agent with custom configurations, your own Telegram bot, and a custom subdomain.
            </p>
            
            <div className="flex flex-wrap gap-4">
              <div className="flex items-center gap-2 text-sm text-zinc-400">
                <Check className="w-4 h-4 text-green-500" />
                <span>Under 2 minutes setup</span>
              </div>
              <div className="flex items-center gap-2 text-sm text-zinc-400">
                <Check className="w-4 h-4 text-green-500" />
                <span>Custom subdomain</span>
              </div>
              <div className="flex items-center gap-2 text-sm text-zinc-400">
                <Check className="w-4 h-4 text-green-500" />
                <span>Telegram integration</span>
              </div>
            </div>
          </div>

          {/* Right column - Signup Form */}
          <div className="relative">
            <div className="absolute -inset-1 bg-gradient-to-r from-blue-500 to-purple-600 rounded-2xl blur opacity-20" />
            
            <div className="relative bg-zinc-900/50 backdrop-blur-xl border border-zinc-800 rounded-2xl p-8">
              <div className="mb-6">
                <h2 className="text-2xl font-bold mb-2">Get Started</h2>
                <p className="text-zinc-400">
                  $29/month · Cancel anytime
                </p>
              </div>

              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">Email</label>
                  <input
                    type="email"
                    required
                    value={formData.email}
                    onChange={(e) => setFormData({...formData, email: e.target.value})}
                    className="input-dark w-full px-4 py-3 rounded-lg text-white placeholder-zinc-500"
                    placeholder="you@example.com"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Assistant Name</label>
                  <input
                    type="text"
                    required
                    value={formData.assistantName}
                    onChange={(e) => setFormData({...formData, assistantName: e.target.value})}
                    className="input-dark w-full px-4 py-3 rounded-lg text-white placeholder-zinc-500"
                    placeholder="My AI Assistant"
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
                    type="text"
                    required
                    value={formData.telegramToken}
                    onChange={(e) => setFormData({...formData, telegramToken: e.target.value})}
                    className="input-dark w-full px-4 py-3 rounded-lg text-white placeholder-zinc-500 font-mono text-sm"
                    placeholder="123456:ABC-DEF1234..."
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Custom Instructions</label>
                  <textarea
                    required
                    rows={3}
                    value={formData.instructions}
                    onChange={(e) => setFormData({...formData, instructions: e.target.value})}
                    className="input-dark w-full px-4 py-3 rounded-lg text-white placeholder-zinc-500 resize-none"
                    placeholder="I'm a freelance developer. I need help with client proposals, research, and scheduling..."
                  />
                </div>

                <button
                  type="submit"
                  disabled={isLoading}
                  className="btn-primary w-full py-4 rounded-lg font-semibold text-white flex items-center justify-center gap-2"
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      Setting up...
                    </>
                  ) : (
                    <>
                      Deploy Now
                      <ChevronRight className="w-5 h-5" />
                    </>
                  )}
                </button>
              </form>

              <p className="mt-4 text-xs text-center text-zinc-500">
                By signing up, you agree to our Terms and Privacy Policy
              </p>
            </div>
          </div>
        </div>
      </main>

      {/* Features Section */}
      <section id="features" className="relative z-10 border-t border-zinc-800 bg-zinc-900/30">
        <div className="max-w-7xl mx-auto px-6 py-24">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">Everything you need</h2>
            <p className="text-zinc-400 max-w-2xl mx-auto">
              Powerful features to help you deploy and manage AI assistants at scale
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            <div className="card-hover bg-zinc-900/50 border border-zinc-800 rounded-xl p-8">
              <div className="w-12 h-12 bg-blue-500/10 rounded-lg flex items-center justify-center mb-6">
                <Zap className="w-6 h-6 text-blue-500" />
              </div>
              <h3 className="text-xl font-semibold mb-3">Lightning Fast</h3>
              <p className="text-zinc-400">
                Deploy your assistant in under 2 minutes. Automated provisioning with zero configuration.
              </p>
            </div>

            <div className="card-hover bg-zinc-900/50 border border-zinc-800 rounded-xl p-8">
              <div className="w-12 h-12 bg-purple-500/10 rounded-lg flex items-center justify-center mb-6">
                <Globe className="w-6 h-6 text-purple-500" />
              </div>
              <h3 className="text-xl font-semibold mb-3">Custom Domain</h3>
              <p className="text-zinc-400">
                Get your own subdomain like alice.blytz.cloud. Access your assistant from anywhere.
              </p>
            </div>

            <div className="card-hover bg-zinc-900/50 border border-zinc-800 rounded-xl p-8">
              <div className="w-12 h-12 bg-green-500/10 rounded-lg flex items-center justify-center mb-6">
                <Shield className="w-6 h-6 text-green-500" />
              </div>
              <h3 className="text-xl font-semibold mb-3">Secure by Default</h3>
              <p className="text-zinc-400">
                Isolated Docker containers, encrypted API keys, and automatic security updates.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-zinc-800 py-12">
        <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 bg-gradient-to-br from-blue-500 to-purple-600 rounded-md flex items-center justify-center">
              <Bot className="w-4 h-4 text-white" />
            </div>
            <span className="font-semibold">Blytz</span>
          </div>
          
          <p className="text-zinc-500 text-sm">
            © 2026 Blytz. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  )
}
