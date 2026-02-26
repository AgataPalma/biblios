import type { ReactNode } from 'react'

interface CardProps {
    children: ReactNode
    onClick?: () => void
    hover?: boolean
    padding?: 'sm' | 'md' | 'lg'
    className?: string
}

export default function Card({ children, onClick, hover = false, padding = 'md', className = '' }: CardProps) {
    const paddings = {
        sm: '12px',
        md: '20px',
        lg: '32px',
    }

    return (
        <div
            onClick={onClick}
            className={className}
            style={{
                background: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--border-radius)',
                boxShadow: 'var(--shadow-sm)',
                padding: paddings[padding],
                cursor: onClick ? 'pointer' : 'default',
                transition: 'var(--transition)',
                ...(hover && onClick ? {
                    ['--hover-shadow' as string]: 'var(--shadow-md)',
                } : {}),
            }}
            onMouseEnter={e => {
                if (hover && onClick) {
                    (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-md)'
                    ;(e.currentTarget as HTMLDivElement).style.transform = 'translateY(-2px)'
                }
            }}
            onMouseLeave={e => {
                if (hover && onClick) {
                    (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-sm)'
                    ;(e.currentTarget as HTMLDivElement).style.transform = 'translateY(0)'
                }
            }}
        >
            {children}
        </div>
    )
}