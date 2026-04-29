"use client"

import React, { useRef, useState } from "react"

export interface MagicCardProps extends React.HTMLAttributes<HTMLDivElement> {
  gradientSize?: number
  gradientColor?: string
  gradientOpacity?: number
}

export function MagicCard({
  children,
  className,
  gradientSize = 200,
  gradientColor = "#262626",
  gradientOpacity = 0.8,
  ...props
}: MagicCardProps) {
  const cardRef = useRef<HTMLDivElement>(null)
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 })
  const [isHovered, setIsHovered] = useState(false)

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!cardRef.current) return
    const rect = cardRef.current.getBoundingClientRect()
    setMousePosition({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    })
  }

  return (
    <div
      ref={cardRef}
      onMouseMove={handleMouseMove}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className={`relative overflow-hidden ${className || ""}`}
      {...props}
    >
      <div
        className="pointer-events-none absolute -inset-px rounded-xl opacity-0 transition duration-300"
        style={{
          opacity: isHovered ? gradientOpacity : 0,
          background: `radial-gradient(${gradientSize}px circle at ${mousePosition.x}px ${mousePosition.y}px, ${gradientColor}, transparent 100%)`,
        }}
      />
      <div className="relative z-10 h-full w-full">{children}</div>
    </div>
  )
}

