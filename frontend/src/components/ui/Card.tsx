import { ReactNode } from 'react'

interface CardProps {
  title?: string
  children: ReactNode
  className?: string
}

export function Card({ title, children, className = '' }: CardProps) {
  return (
    <div className={`bg-white/[0.03] ring-1 ring-white/[0.06] rounded-lg p-5 ${className}`}>
      {title && <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-4">{title}</h3>}
      {children}
    </div>
  )
}
