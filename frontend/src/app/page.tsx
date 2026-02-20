'use client'

import { useState, Suspense, lazy } from 'react'
import { useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import { 
  BoltIcon,
  GlobeAltIcon,
  ShieldCheckIcon,
  CpuChipIcon,
  ChatBubbleLeftRightIcon,
  LockClosedIcon,
  CheckIcon,
  ArrowRightIcon,
  XMarkIcon,
  ClockIcon,
  SparklesIcon,
  CommandLineIcon
} from '@heroicons/react/24/solid'
import { Marquee, GradientCard, ShimmerButton } from '@/components/ui-effects'

// Lazy load heavy animation components
const SparkleEffect = lazy(() => import('@/components/animations').then(mod => ({ default: mod.Sparkles })))
const FloatingOrbs = lazy(() => import('@/components/animations').then(mod => ({ default: mod.FloatingOrbs })))
const StarField = lazy(() => import('@/components/animations').then(mod => ({ default: mod.StarField })))

// Loading fallback for animations
const AnimationFallback = () => <div className="absolute inset-0" />

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.2,
    },
  },
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.5,
      ease: [0.4, 0, 0.2, 1] as [number, number, number, number],
    },
  },
}

const features = [
  {
    icon: BoltIcon,
    title: 'Lightning Fast',
    desc: 'Deploy in under 2 minutes with zero config',
    gradient: 'from-amber-400 via-orange-400 to-amber-500'
  },
  {
    icon: GlobeAltIcon,
    title: 'Custom Domain',
    desc: 'Your own subdomain like alice.blytz.cloud',
    gradient: 'from-orange-400 via-amber-400 to-yellow-400'
  },
  {
    icon: ShieldCheckIcon,
    title: 'Secure by Default',
    desc: 'Isolated containers with encrypted keys',
    gradient: 'from-yellow-400 via-amber-500 to-orange-500'
  },
  {
    icon: CpuChipIcon,
    title: 'Auto-Scaling',
    desc: 'Resources scale automatically with demand',
    gradient: 'from-amber-500 via-orange-400 to-amber-400'
  },
  {
    icon: ChatBubbleLeftRightIcon,
    title: 'Telegram Ready',
    desc: 'Connect your bot in one click',
    gradient: 'from-orange-500 via-amber-400 to-yellow-500'
  },
  {
    icon: LockClosedIcon,
    title: 'Private & Safe',
    desc: 'Your data never leaves your container',
    gradient: 'from-amber-400 via-yellow-400 to-orange-400'
  },
]

const comparison = [
  { traditional: 'Weeks of setup time', blytz: '2 minutes to deploy', icon: ClockIcon },
  { traditional: 'Complex infrastructure', blytz: 'Zero configuration needed', icon: CpuChipIcon },
  { traditional: 'Security headaches', blytz: 'Secure by default', icon: ShieldCheckIcon },
  { traditional: 'Expensive hosting', blytz: '$29/month flat rate', icon: BoltIcon },
]

const marqueeItems = [
  'AI Assistants',
  'Custom Domains',
  'Telegram Bots',
  'Docker Containers',
  'Auto-Scaling',
  'Secure APIs',
  'Instant Deploy',
  'Zero Config',
]

export default function LandingPage() {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(false)
  const [formData, setFormData] = useState({
    email: ''
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    
    console.log('Signup:', formData)
    
    setTimeout(() => {
      setIsLoading(false)
      router.push('/dashboard')
    }, 1500)
  }

  return (
    <div className="min-h-screen bg-black text-white overflow-x-hidden noise-overlay">
      {/* Background Effects - Lazy loaded */}
      <Suspense fallback={<AnimationFallback />}>
        <FloatingOrbs />
      </Suspense>
      <Suspense fallback={<AnimationFallback />}>
        <StarField />
      </Suspense>
      
      {/* Navigation */}
      <motion.nav 
        initial={{ y: -20, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        className="relative z-50 flex items-center justify-between px-6 py-4 max-w-7xl mx-auto"
      >
        <div className="flex items-center gap-3">
          <div className="relative w-10 h-10">
            <div className="absolute inset-0 bg-gradient-to-br from-amber-500 to-orange-600 rounded-xl blur-sm animate-glow-pulse" />
            <div className="relative w-10 h-10 bg-gradient-to-br from-amber-500 to-orange-600 rounded-xl flex items-center justify-center">
              <CommandLineIcon className="w-6 h-6 text-white" />
            </div>
          </div>
          <span className="text-2xl font-bold bg-gradient-to-r from-white to-amber-200 bg-clip-text text-transparent">
            Blytz
          </span>
        </div>
        
        <div className="hidden md:flex items-center gap-8 text-sm text-amber-200/60">
          {['Features', 'Comparison', 'Pricing'].map((item) => (
            <a 
              key={item}
              href={`#${item.toLowerCase()}`} 
              className="hover:text-amber-200 transition-colors relative group"
            >
              {item}
              <span className="absolute -bottom-1 left-0 w-0 h-0.5 bg-gradient-to-r from-amber-500 to-orange-500 group-hover:w-full transition-all duration-300" />
            </a>
          ))}
        </div>
        
        <button className="px-5 py-2.5 text-sm font-medium bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/30 rounded-xl transition-all hover:border-amber-400/50 text-amber-300">
          Sign In
        </button>
      </motion.nav>

      {/* Hero Section */}
      <section className="relative z-10 max-w-7xl mx-auto px-6 pt-16 pb-32">
        <Suspense fallback={<AnimationFallback />}>
          <SparkleEffect />
        </Suspense>
        
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          {/* Left column - Copy */}
          <motion.div 
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="space-y-8"
          >
            <motion.div 
              variants={itemVariants}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-amber-500/10 to-orange-500/10 border border-amber-500/30 text-amber-300 text-sm"
            >
              <SparklesIcon className="w-4 h-4" />
              <span className="font-medium">Now with Agent Marketplace</span>
              <span className="px-2 py-0.5 text-xs bg-amber-500/30 rounded-full text-amber-200">New</span>
            </motion.div>
            
            <motion.h1 
              variants={itemVariants}
              className="text-5xl md:text-7xl font-bold leading-[1.1] tracking-tight"
            >
              Deploy your
              <br />
              <span className="animate-text-shimmer">AI Assistant</span>
              <br />
              in seconds
            </motion.h1>
            
            <motion.p 
              variants={itemVariants}
              className="text-xl text-amber-100/60 max-w-lg leading-relaxed"
            >
              The fastest way to deploy personalized AI assistants. Set up your OpenClaw 
              agent with custom configurations and your own Telegram bot.
            </motion.p>
            
            <motion.div 
              variants={itemVariants}
              className="flex flex-wrap gap-4"
            >
              {['Under 2 minutes setup', 'Custom subdomain', 'Telegram integration'].map((item, i) => (
                <div key={i} className="flex items-center gap-2 text-sm text-amber-100/50 bg-amber-500/5 px-4 py-2 rounded-full border border-amber-500/10">
                  <CheckIcon className="w-4 h-4 text-orange-400" />
                  <span>{item}</span>
                </div>
              ))}
            </motion.div>
          </motion.div>

          {/* Right column - Simplified Signup */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
          >
            <div className="glass-card max-w-md mx-auto p-8 md:p-10 text-center backdrop-blur-md bg-white/[0.03] border border-white/[0.08] rounded-2xl">
              {/* Header */}
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-white mb-2">Get Started in Seconds</h2>
                <p className="text-white/50 text-sm">
                  Connect your LLM provider and Telegram bot after signup
                </p>
              </div>

              {/* Google OAuth Button */}
              <button
                onClick={() => {
                  setIsLoading(true);
                  // TODO: Implement Google OAuth
                  setTimeout(() => {
                    setIsLoading(false);
                    router.push('/dashboard');
                  }, 1500);
                }}
                disabled={isLoading}
                className="w-full flex items-center justify-center gap-3 px-6 py-4 bg-white text-gray-900 rounded-xl font-medium transition-all hover:bg-gray-100 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed mb-4"
              >
                {isLoading ? (
                  <div className="w-5 h-5 border-2 border-gray-400/30 border-t-gray-900 rounded-full animate-spin" />
                ) : (
                  <>
                    <svg className="w-5 h-5" viewBox="0 0 24 24">
                      <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                      <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                      <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                      <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                    </svg>
                    Continue with Google
                  </>
                )}
              </button>

              {/* Divider */}
              <div className="relative my-6">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-white/10" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-4 bg-black/20 text-white/40 backdrop-blur-sm">or</span>
                </div>
              </div>

              {/* Email Signup Form */}
              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <input
                    type="email"
                    required
                    value={formData.email}
                    onChange={(e) => setFormData({...formData, email: e.target.value})}
                    className="w-full px-4 py-3.5 bg-white/[0.05] border border-white/[0.1] rounded-xl text-white placeholder-white/40 focus:border-amber-500/50 focus:ring-2 focus:ring-amber-500/20 transition-all outline-none text-center"
                    placeholder="Enter your email"
                  />
                </div>

                <ShimmerButton
                  type="submit"
                  disabled={isLoading}
                  className="w-full bg-gradient-to-r from-amber-500 via-orange-500 to-amber-500"
                >
                  {isLoading ? (
                    <span className="flex items-center justify-center gap-2">
                      <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                      Creating account...
                    </span>
                  ) : (
                    <span>Continue with Email</span>
                  )}
                </ShimmerButton>
              </form>

              {/* Benefits */}
              <div className="mt-8 pt-6 border-t border-white/[0.06]">
                <div className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2 text-xs text-white/40">
                  <span className="flex items-center gap-1.5">
                    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                    No credit card required
                  </span>
                  <span className="flex items-center gap-1.5">
                    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                    Connect LLM after signup
                  </span>
                </div>
              </div>

              <p className="mt-6 text-xs text-white/30">
                By signing up, you agree to our Terms and Privacy Policy
              </p>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Marquee Section */}
      <section className="relative z-10 py-12 border-y border-amber-900/30 bg-amber-950/20">
        <Marquee speed={40}>
          {marqueeItems.map((item, i) => (
            <div key={i} className="flex items-center gap-3 px-8">
              <div className="w-2 h-2 rounded-full bg-gradient-to-r from-amber-500 to-orange-500" />
              <span className="text-lg font-medium text-amber-200/70 whitespace-nowrap">{item}</span>
            </div>
          ))}
        </Marquee>
      </section>

      {/* Features Section */}
      <section id="features" className="relative z-10 py-32">
        <div className="max-w-7xl mx-auto px-6">
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="text-center mb-20"
          >
            <h2 className="text-4xl md:text-5xl font-bold mb-6 text-white">
              Everything you need to
              <br />
              <span className="animate-text-shimmer">deploy at scale</span>
            </h2>
            <p className="text-xl text-amber-100/50 max-w-2xl mx-auto">
              Powerful features to help you deploy and manage AI assistants without the complexity
            </p>
          </motion.div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.1 }}
              >
                <GradientCard 
                  className="h-full"
                  glowColor={feature.gradient}
                >
                  <div className="p-8 h-full">
                    <div className={`w-14 h-14 bg-gradient-to-br ${feature.gradient} rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-amber-500/20`}>
                      <feature.icon className="w-7 h-7 text-white" />
                    </div>
                    <h3 className="text-xl font-bold mb-3 text-white">{feature.title}</h3>
                    <p className="text-amber-100/50 leading-relaxed">{feature.desc}</p>
                  </div>
                </GradientCard>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Comparison Section */}
      <section id="comparison" className="relative z-10 py-32 bg-amber-950/10">
        <div className="max-w-5xl mx-auto px-6">
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="text-center mb-16"
          >
            <h2 className="text-4xl md:text-5xl font-bold mb-6 text-white">
              Why developers choose
              <span className="animate-text-shimmer ml-2">Blytz</span>
            </h2>
            <p className="text-xl text-amber-100/50">
              See how we compare to traditional deployment methods
            </p>
          </motion.div>

          <div className="space-y-4">
            {comparison.map((item, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, x: -20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.1 }}
                className="grid md:grid-cols-2 gap-4"
              >
                <div className="flex items-center gap-4 p-6 bg-amber-950/30 border border-amber-900/30 rounded-2xl">
                  <XMarkIcon className="w-6 h-6 text-red-400 flex-shrink-0" />
                  <span className="text-amber-100/40">{item.traditional}</span>
                </div>
                
                <div className="flex items-center gap-4 p-6 bg-gradient-to-r from-amber-500/10 to-orange-500/10 border border-amber-500/30 rounded-2xl">
                  <CheckIcon className="w-6 h-6 text-orange-400 flex-shrink-0" />
                  <span className="font-medium text-amber-100">{item.blytz}</span>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="relative z-10 py-32">
        <div className="max-w-4xl mx-auto px-6 text-center">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            className="relative"
          >
            <div className="absolute -inset-4 bg-gradient-to-r from-amber-500/20 via-orange-500/20 to-yellow-500/20 rounded-3xl blur-2xl" />
            
            <div className="relative bg-amber-950/30 border border-amber-800/50 rounded-3xl p-12">
              <Suspense fallback={<AnimationFallback />}>
                <SparkleEffect />
              </Suspense>
              
              <h2 className="text-4xl md:text-5xl font-bold mb-6 text-white">
                Ready to deploy your
                <br />
                <span className="animate-text-shimmer">AI assistant?</span>
              </h2>
              
              <p className="text-xl text-amber-100/50 mb-8 max-w-2xl mx-auto">
                Join thousands of developers who trust Blytz for their AI assistant deployment.
              </p>
              
              <ShimmerButton 
                onClick={() => router.push('#')}
                className="text-lg px-12 py-5 bg-gradient-to-r from-amber-600 via-orange-600 to-amber-600"
              >
                Get Started Free
              </ShimmerButton>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Footer */}
      <footer className="relative z-10 border-t border-amber-900/30 py-16">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex flex-col md:flex-row items-center justify-between gap-8">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-gradient-to-br from-amber-500 to-orange-600 rounded-lg flex items-center justify-center">
                <CommandLineIcon className="w-5 h-5 text-white" />
              </div>
              <span className="text-xl font-bold text-white">Blytz</span>
            </div>
            
            <div className="flex items-center gap-8 text-sm text-amber-600">
              <a href="#" className="hover:text-amber-300 transition-colors">Terms</a>
              <a href="#" className="hover:text-amber-300 transition-colors">Privacy</a>
              <a href="#" className="hover:text-amber-300 transition-colors">Docs</a>
            </div>
            
            <p className="text-amber-700 text-sm">
              © 2026 Blytz. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}
