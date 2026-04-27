import type { ButtonHTMLAttributes } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost'
  loading?: boolean
}

export function Button({ variant = 'primary', loading = false, className = '', children, disabled, ...props }: ButtonProps) {
  const base = 'inline-flex items-center justify-center rounded-lg px-6 py-3 font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed'
  const variants = {
    primary: 'bg-[#162447] text-white hover:bg-[#1f3a6e] focus:ring-[#162447]',
    secondary: 'border border-[#162447] text-[#162447] hover:bg-[#162447]/5 focus:ring-[#162447]',
    ghost: 'text-[#162447] hover:underline focus:ring-[#162447]',
  }

  return (
    <button
      className={`${base} ${variants[variant]} ${className}`}
      disabled={disabled || loading}
      {...props}
    >
      {loading ? 'Carregando…' : children}
    </button>
  )
}
