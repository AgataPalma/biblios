import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { acceptInvitation } from '../api/libraries'
import { Button, Spinner } from '../components'

type Status = 'loading' | 'success' | 'error'

export default function InvitationAcceptPage() {
    const { token } = useParams<{ token: string }>()
    const navigate = useNavigate()
    const [status, setStatus] = useState<Status>('loading')
    const [errorMessage, setErrorMessage] = useState<string>('')

    useEffect(() => {
        if (!token) {
            setErrorMessage('This invitation link is invalid or has expired.')
            setStatus('error')
            return
        }

        acceptInvitation(token)
            .then(() => {
                setStatus('success')
            })
            .catch((err) => {
                const message =
                    err?.response?.data?.message ||
                    err?.message ||
                    'This invitation link is invalid or has expired.'
                setErrorMessage(message)
                setStatus('error')
            })
    }, [token])

    const containerStyle: React.CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        padding: '32px',
    }

    const cardStyle: React.CSSProperties = {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '16px',
        maxWidth: '400px',
        width: '100%',
        textAlign: 'center',
    }

    const emojiStyle: React.CSSProperties = {
        fontSize: '64px',
        lineHeight: 1,
    }

    const headingStyle: React.CSSProperties = {
        margin: 0,
        fontSize: '24px',
        fontWeight: 600,
        color: 'var(--color-text)',
        fontFamily: 'var(--font-heading)',
    }

    const descStyle: React.CSSProperties = {
        margin: 0,
        fontSize: '15px',
        color: 'var(--color-text-muted)',
        fontFamily: 'var(--font-body)',
    }

    if (status === 'loading') {
        return (
            <div style={containerStyle}>
                <Spinner size="lg" label="Accepting invitation…" />
            </div>
        )
    }

    if (status === 'success') {
        return (
            <div style={containerStyle}>
                <div style={cardStyle}>
                    <span style={emojiStyle}>✅</span>
                    <h1 style={headingStyle}>Invitation accepted!</h1>
                    <p style={descStyle}>You've been added to the library.</p>
                    <Button label="View Libraries" onClick={() => navigate('/libraries')} />
                </div>
            </div>
        )
    }

    return (
        <div style={containerStyle}>
            <div style={cardStyle}>
                <span style={emojiStyle}>❌</span>
                <h1 style={headingStyle}>Invitation failed</h1>
                <p style={descStyle}>{errorMessage}</p>
                <Button label="Go home" onClick={() => navigate('/')} variant="secondary" />
            </div>
        </div>
    )
}
