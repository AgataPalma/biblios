interface SpinnerProps {
    size?: 'sm' | 'md' | 'lg'
    label?: string
}

export default function Spinner({ size = 'md', label }: SpinnerProps) {
    const sizes = { sm: 16, md: 28, lg: 44 }
    const px = sizes[size]
    const stroke = size === 'sm' ? 2 : size === 'md' ? 3 : 4

    return (
        <div
            style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '12px',
            }}
        >
            <svg
                width={px}
                height={px}
                viewBox="0 0 24 24"
                fill="none"
                style={{ animation: 'spin 0.8s linear infinite' }}
            >
                <circle
                    cx="12" cy="12" r="10"
                    stroke="var(--color-border)"
                    strokeWidth={stroke}
                />
                <path
                    d="M12 2a10 10 0 0 1 10 10"
                    stroke="var(--color-primary)"
                    strokeWidth={stroke}
                    strokeLinecap="round"
                />
            </svg>
            {label && (
                <p style={{
                    margin: 0,
                    fontSize: '13px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                }}>
                    {label}
                </p>
            )}
            <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
        </div>
    )
}