interface ButtonProps {
    label: string
    onClick?: () => void
    type?: 'button' | 'submit'
    variant?: 'primary' | 'secondary' | 'danger'
    isLoading?: boolean
    disabled?: boolean
    fullWidth?: boolean
}

export default function Button({
                                   label,
                                   onClick,
                                   type = 'button',
                                   variant = 'primary',
                                   isLoading = false,
                                   disabled = false,
                                   fullWidth = false,
                               }: ButtonProps) {
    const styles: Record<string, React.CSSProperties> = {
        primary: {
            background: 'var(--color-primary)',
            color: 'var(--color-primary-text)',
        },
        secondary: {
            background: 'var(--color-surface-alt)',
            color: 'var(--color-text)',
            border: '1px solid var(--color-border)',
        },
        danger: {
            background: 'var(--color-error)',
            color: '#ffffff',
        },
    }

    return (
        <button
            type={type}
            onClick={onClick}
            disabled={disabled || isLoading}
            className="px-4 py-2 text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            style={{
                ...styles[variant],
                borderRadius: 'var(--border-radius)',
                transition: 'var(--transition)',
                width: fullWidth ? '100%' : 'auto',
                fontFamily: 'var(--font-body)',
            }}
        >
            {isLoading ? 'Loading...' : label}
        </button>
    )
}