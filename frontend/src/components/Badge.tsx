type BadgeVariant = 'default' | 'success' | 'warning' | 'error' | 'info' | 'muted'

interface BadgeProps {
    label: string
    variant?: BadgeVariant
    size?: 'sm' | 'md'
}

export default function Badge({ label, variant = 'default', size = 'md' }: BadgeProps) {
    const colors: Record<BadgeVariant, { bg: string; text: string; border: string }> = {
        default: {
            bg: 'var(--color-surface-alt)',
            text: 'var(--color-text)',
            border: 'var(--color-border)',
        },
        success: {
            bg: 'rgba(34,197,94,0.15)',
            text: 'var(--color-success)',
            border: 'rgba(34,197,94,0.3)',
        },
        warning: {
            bg: 'rgba(234,179,8,0.15)',
            text: '#ca8a04',
            border: 'rgba(234,179,8,0.3)',
        },
        error: {
            bg: 'rgba(239,68,68,0.15)',
            text: 'var(--color-error)',
            border: 'rgba(239,68,68,0.3)',
        },
        info: {
            bg: 'rgba(59,130,246,0.15)',
            text: 'var(--color-primary)',
            border: 'rgba(59,130,246,0.3)',
        },
        muted: {
            bg: 'transparent',
            text: 'var(--color-text-muted)',
            border: 'transparent',
        },
    }

    const c = colors[variant]
    const fontSize = size === 'sm' ? '10px' : '11px'
    const padding = size === 'sm' ? '2px 6px' : '3px 8px'

    return (
        <span
            style={{
                display: 'inline-flex',
                alignItems: 'center',
                background: c.bg,
                color: c.text,
                border: `1px solid ${c.border}`,
                borderRadius: '999px',
                fontSize,
                fontWeight: 600,
                padding,
                letterSpacing: '0.03em',
                textTransform: 'uppercase',
                whiteSpace: 'nowrap',
                fontFamily: 'var(--font-body)',
            }}
        >
      {label}
    </span>
    )
}