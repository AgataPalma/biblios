interface EmptyStateProps {
    icon?: string
    title: string
    description?: string
    action?: {
        label: string
        onClick: () => void
    }
}

export default function EmptyState({ icon = '📭', title, description, action }: EmptyStateProps) {
    return (
        <div
            style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                padding: '48px 24px',
                textAlign: 'center',
                gap: '12px',
            }}
        >
            <span style={{ fontSize: '48px', lineHeight: 1 }}>{icon}</span>
            <h3
                style={{
                    margin: 0,
                    fontSize: '16px',
                    fontWeight: 600,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-heading)',
                }}
            >
                {title}
            </h3>
            {description && (
                <p
                    style={{
                        margin: 0,
                        fontSize: '13px',
                        color: 'var(--color-text-muted)',
                        maxWidth: '320px',
                        lineHeight: 1.5,
                        fontFamily: 'var(--font-body)',
                    }}
                >
                    {description}
                </p>
            )}
            {action && (
                <button
                    onClick={action.onClick}
                    style={{
                        marginTop: '8px',
                        padding: '8px 20px',
                        background: 'var(--color-primary)',
                        color: 'var(--color-primary-text)',
                        border: 'none',
                        borderRadius: 'var(--border-radius)',
                        fontSize: '13px',
                        fontWeight: 600,
                        cursor: 'pointer',
                        transition: 'var(--transition)',
                        fontFamily: 'var(--font-body)',
                    }}
                >
                    {action.label}
                </button>
            )}
        </div>
    )
}