interface AvatarProps {
    name: string
    avatarUrl?: string
    size?: 'sm' | 'md' | 'lg'
}

export default function Avatar({ name, avatarUrl, size = 'md' }: AvatarProps) {
    const sizes = { sm: 28, md: 36, lg: 52 }
    const fontSizes = { sm: '11px', md: '14px', lg: '20px' }
    const px = sizes[size]

    // Generate a consistent colour from name
    const colors = [
        '#2563eb', '#16a34a', '#dc2626', '#9333ea',
        '#ea580c', '#0891b2', '#d97706', '#4f46e5',
    ]
    const colorIndex = name.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % colors.length
    const bg = colors[colorIndex]

    const initials = name
        .split(' ')
        .map(w => w[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)

    if (avatarUrl) {
        return (
            <img
                src={avatarUrl}
                alt={name}
                style={{
                    width: px,
                    height: px,
                    borderRadius: '50%',
                    objectFit: 'cover',
                    border: '2px solid var(--color-border)',
                }}
            />
        )
    }

    return (
        <div
            style={{
                width: px,
                height: px,
                borderRadius: '50%',
                background: bg,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#ffffff',
                fontSize: fontSizes[size],
                fontWeight: 700,
                fontFamily: 'var(--font-body)',
                border: '2px solid var(--color-border)',
                flexShrink: 0,
            }}
        >
            {initials}
        </div>
    )
}