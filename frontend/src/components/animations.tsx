'use client'

import { useEffect, useState, useMemo } from 'react'
import { motion, useReducedMotion } from 'framer-motion'

interface Star {
  id: number
  x: number
  y: number
  size: number
  duration: number
  delay: number
}

interface Orb {
  id: number
  x: number
  y: number
  size: number
  duration: number
  color: string
}

// Reduced motion hook wrapper
function useAnimationsEnabled() {
  const shouldReduceMotion = useReducedMotion()
  return !shouldReduceMotion
}

// Sparkle effect component - CSS-based for better performance
export function Sparkles() {
  const animationsEnabled = useAnimationsEnabled()
  const [sparkles, setSparkles] = useState<Array<{ id: number; x: number; y: number; delay: number; color: string }>>([])
  
  const sparkleColors = [
    'rgba(251, 191, 36, 0.9)',   // amber-400
    'rgba(251, 146, 60, 0.9)',   // orange-400
    'rgba(245, 158, 11, 0.9)',   // amber-500
    'rgba(234, 179, 8, 0.9)',    // yellow-500
    'rgba(249, 115, 22, 0.9)',   // orange-500
  ]
  
  useEffect(() => {
    if (!animationsEnabled) return
    
    const generateSparkles = () => {
      const newSparkles = Array.from({ length: 12 }, (_, i) => ({
        id: i,
        x: Math.random() * 100,
        y: Math.random() * 100,
        delay: Math.random() * 3,
        color: sparkleColors[Math.floor(Math.random() * sparkleColors.length)]
      }))
      setSparkles(newSparkles)
    }
    
    generateSparkles()
    const interval = setInterval(generateSparkles, 4000)
    return () => clearInterval(interval)
  }, [animationsEnabled])
  
  if (!animationsEnabled) return null
  
  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {sparkles.map((sparkle) => (
        <div
          key={sparkle.id}
          className="absolute w-1.5 h-1.5 rounded-full animate-twinkle"
          style={{
            left: `${sparkle.x}%`,
            top: `${sparkle.y}%`,
            backgroundColor: sparkle.color,
            boxShadow: `0 0 6px ${sparkle.color}`,
            animationDelay: `${sparkle.delay}s`,
            willChange: 'opacity, transform',
            transform: 'translateZ(0)',
          }}
        />
      ))}
    </div>
  )
}

// Floating orbs - optimized with CSS transforms
export function FloatingOrbs() {
  const animationsEnabled = useAnimationsEnabled()
  
  const orbs = useMemo<Orb[]>(() => [
    { id: 1, x: 10, y: 20, size: 300, duration: 20, color: 'rgba(251, 191, 36, 0.3)' },   // amber-400
    { id: 2, x: 70, y: 60, size: 400, duration: 25, color: 'rgba(249, 115, 22, 0.25)' },  // orange-500
    { id: 3, x: 50, y: 10, size: 250, duration: 18, color: 'rgba(245, 158, 11, 0.35)' },  // amber-500
    { id: 4, x: 80, y: 30, size: 350, duration: 22, color: 'rgba(234, 179, 8, 0.2)' },   // yellow-500
  ], [])
  
  if (!animationsEnabled) {
    // Static version for reduced motion
    return (
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        {orbs.map((orb) => (
          <div
            key={orb.id}
            className="absolute rounded-full blur-2xl"
            style={{
              width: orb.size,
              height: orb.size,
              background: `radial-gradient(circle, ${orb.color} 0%, transparent 70%)`,
              left: `${orb.x}%`,
              top: `${orb.y}%`,
            }}
          />
        ))}
      </div>
    )
  }
  
  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {orbs.map((orb) => (
        <motion.div
          key={orb.id}
          className="absolute rounded-full blur-2xl"
          style={{
            width: orb.size,
            height: orb.size,
            background: `radial-gradient(circle, ${orb.color} 0%, transparent 70%)`,
            left: `${orb.x}%`,
            top: `${orb.y}%`,
            willChange: 'transform',
            transform: 'translateZ(0)',
          }}
          animate={{
            x: [0, 30, -20, 0],
            y: [0, -20, 30, 0],
            scale: [1, 1.1, 0.95, 1],
          }}
          transition={{
            duration: orb.duration,
            repeat: Infinity,
            ease: "easeInOut"
          }}
        />
      ))}
    </div>
  )
}

// Star field - reduced count and CSS-based animations
export function StarField() {
  const animationsEnabled = useAnimationsEnabled()
  const [stars, setStars] = useState<Star[]>([])
  const [mounted, setMounted] = useState(false)
  
  useEffect(() => {
    setMounted(true)
    // Reduced from 50 to 25 stars for better performance
    const newStars = Array.from({ length: 25 }, (_, i) => ({
      id: i,
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 1.5 + 0.5,
      duration: Math.random() * 2 + 2,
      delay: Math.random() * 5
    }))
    setStars(newStars)
  }, [])
  
  if (!mounted) return null
  
  if (!animationsEnabled) {
    // Static stars for reduced motion
    return (
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        {stars.map((star) => (
          <div
            key={star.id}
            className="absolute rounded-full opacity-60"
            style={{
              left: `${star.x}%`,
              top: `${star.y}%`,
              width: star.size,
              height: star.size,
              backgroundColor: 'rgba(251, 191, 36, 0.8)',
              boxShadow: '0 0 4px rgba(251, 191, 36, 0.6)',
            }}
          />
        ))}
      </div>
    )
  }

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      {stars.map((star) => (
        <div
          key={star.id}
          className="absolute rounded-full animate-twinkle"
          style={{
            left: `${star.x}%`,
            top: `${star.y}%`,
            width: star.size,
            height: star.size,
            backgroundColor: 'rgba(251, 191, 36, 0.9)',
            boxShadow: '0 0 4px rgba(251, 191, 36, 0.6)',
            animationDuration: `${star.duration}s`,
            animationDelay: `${star.delay}s`,
            willChange: 'opacity, transform',
            transform: 'translateZ(0)',
          }}
        />
      ))}
    </div>
  )
}
