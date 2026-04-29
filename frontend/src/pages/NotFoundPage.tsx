import { useNavigate } from 'react-router-dom'
import { Button } from '../components'

export default function NotFoundPage() {
    const navigate = useNavigate()

    return (
        <div
            style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                minHeight: '100vh',
                gap: '1rem',
                background: 'var(--color-background)',
                fontFamily: 'var(--font-body)',
                textAlign: 'center',
                padding: '2rem',
            }}
        >
            <span style={{ fontSize: '4rem', lineHeight: 1 }}>📚</span>
            <h1
                style={{
                    fontFamily: 'var(--font-heading)',
                    fontSize: '2rem',
                    fontWeight: 700,
                    color: 'var(--color-text)',
                    margin: 0,
                }}
            >
                Page Not Found
            </h1>
            <p
                style={{
                    color: 'var(--color-text-muted)',
                    fontSize: '1rem',
                    margin: 0,
                    maxWidth: '360px',
                }}
            >
                The page you're looking for doesn't exist.
            </p>
            <Button label="Go home" onClick={() => navigate('/')} variant="primary" />
        </div>
    )
}
