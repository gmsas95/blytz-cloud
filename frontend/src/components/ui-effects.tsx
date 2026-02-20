'use client'

import { ReactNode } from 'react'
import { motion } from 'framer-motion'

interface MarqueeProps {
  children: ReactNode
  speed?: number
  direction?: 'left' | 'right'
}

interface GradientCardProps {
  children: ReactNode
  className?: string
  glowColor?: string
}

interface ShimmerButtonProps {
  children: ReactNode
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  className?: string
  onClick?: () => void
}

// Marquee component - infinite scrolling text/items
export function Marquee({ 
  children, 
  speed = 40, 
  direction = 'left' 
}: MarqueeProps) {
  return (
    <div className="overflow-hidden">
      <motion.div
        className="flex gap-4"
        animate={{
          x: direction === 'left' ? ['0%', '-50%'] : ['-50%', '0%']
        }}
        transition={{
          x: {
            repeat: Infinity,
            repeatType: 'loop',
            duration: speed,
            ease: 'linear'
          }
        }}
      >
        {children}
        {children}
      </motion.div>
    </div>
  )
}

// GradientCard component - card with animated gradient border glow
export function GradientCard({ 
  children, 
  className = '', 
  glowColor = 'from-amber-500 via-orange-500 to-yellow-500' 
}: GradientCardProps) {
  return (
    <div className={`relative group ${className}`}>
      {/* Glow effect */}
      <div 
        className={`absolute -inset-0.5 bg-gradient-to-r ${glowColor} rounded-2xl opacity-30 group-hover:opacity-50 blur transition duration-500`}
      />
      
      {/* Card content */}
      <div className="relative bg-gradient-to-b from-amber-950/80 to-black/80 backdrop-blur-sm rounded-2xl border border-amber-800/30 overflow-hidden">
        {children}
      </div>
    </div>
  )
}

// ShimmerButton component - button with animated shimmer effect
export function ShimmerButton({ 
  children, 
  type = 'button',
  disabled = false,
  className = '',
  onClick
}: ShimmerButtonProps) {
  return (
    <motion.button
      type={type}
      disabled={disabled}
      onClick={onClick}
      className={`
        relative overflow-hidden px-6 py-3 rounded-xl font-medium text-white
        transition-all duration-300
        disabled:opacity-50 disabled:cursor-not-allowed
        ${className}
      `}
      whileHover={{ scale: disabled ? 1 : 1.02 }}
      whileTap={{ scale: disabled ? 1 : 0.98 }}
    >
      {/* Shimmer effect */}
      <div className="absolute inset-0 overflow-hidden">
        <motion.div
          className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent"
          animate={{
            x: ['-100%', '100%']
          }}
          transition={{
            duration: 2,
            repeat: Infinity,
            ease: 'linear'
          }}
        />
      </div>
      
      {/* Button content */}
      <span className="relative z-10">{children}</span>
    </motion.button>
  )
}
