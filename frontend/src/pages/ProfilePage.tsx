import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import Avatar from '../components/Avatar'
import Button from '../components/Button'
import Input from '../components/Input'
import Modal from '../components/Modal'
import Spinner from '../components/Spinner'
import { updateProfile, updateEmail, updatePassword } from '../api/users'
import type { UpdateProfilePayload, UpdateEmailPayload, UpdatePasswordPayload } from '../api/users'

// ─── extended user type ───────────────────────────────────────────────────────

interface UserProfile {
    id: string
    email: string
    username: string
    role: string
    theme: string
    bio?: string
    avatar_url?: string
    created_at: string
    updated_at: string
}

// ─── types ────────────────────────────────────────────────────────────────────

type Section = 'profile' | 'email' | 'password'

// ─── Dicebear config ──────────────────────────────────────────────────────────

const DICEBEAR_STYLES = [
    { id: 'avataaars',  label: 'Illustrated' },
    { id: 'bottts',     label: 'Robots' },
    { id: 'pixel-art',  label: 'Pixel Art' },
    { id: 'lorelei',    label: 'Lorelei' },
    { id: 'thumbs',     label: 'Thumbs' },
    { id: 'fun-emoji',  label: 'Emoji' },
]

const SEEDS = ['alpha', 'beta', 'gamma', 'delta', 'epsilon', 'zeta', 'eta', 'theta', 'iota', 'kappa', 'lambda', 'mu']

function dicebearUrl(style: string, seed: string) {
    return `https://api.dicebear.com/9.x/${style}/svg?seed=${encodeURIComponent(seed)}`
}

// ─── small helpers ────────────────────────────────────────────────────────────

function SectionTab({ label, id, active, onClick }: {
    label: string; id: Section; active: Section; onClick: (id: Section) => void
}) {
    const isActive = id === active
    return (
        <button
            onClick={() => onClick(id)}
            style={{
                background: 'none', border: 'none', cursor: 'pointer',
                padding: '10px 20px', fontSize: '14px',
                fontFamily: 'var(--font-body)',
                fontWeight: isActive ? 600 : 400,
                color: isActive ? 'var(--color-primary)' : 'var(--color-text-muted)',
                borderBottom: isActive ? '2px solid var(--color-primary)' : '2px solid transparent',
                transition: 'all 0.15s ease', whiteSpace: 'nowrap',
            }}
        >
            {label}
        </button>
    )
}

function SuccessBanner({ message }: { message: string }) {
    return (
        <div style={{
            padding: '12px 16px',
            background: 'color-mix(in srgb, var(--color-primary) 12%, transparent)',
            border: '1px solid color-mix(in srgb, var(--color-primary) 35%, transparent)',
            borderRadius: 'var(--border-radius)',
            color: 'var(--color-primary)', fontSize: '14px', fontFamily: 'var(--font-body)',
        }}>
            ✓ {message}
        </div>
    )
}

function ErrorBanner({ message }: { message: string }) {
    return (
        <div style={{
            padding: '12px 16px',
            background: 'color-mix(in srgb, #ef4444 10%, transparent)',
            border: '1px solid color-mix(in srgb, #ef4444 30%, transparent)',
            borderRadius: 'var(--border-radius)',
            color: '#ef4444', fontSize: '14px', fontFamily: 'var(--font-body)',
        }}>
            {message}
        </div>
    )
}

// ─── Avatar Picker Modal ──────────────────────────────────────────────────────

function AvatarPickerModal({
                               isOpen,
                               onClose,
                               currentAvatarUrl,
                               onSelect,
                               isSaving,
                           }: {
    isOpen: boolean
    onClose: () => void
    currentAvatarUrl?: string
    onSelect: (url: string) => void
    isSaving: boolean
}) {
    const [activeStyle, setActiveStyle] = useState(DICEBEAR_STYLES[0].id)
    const [selected, setSelected] = useState<string | null>(currentAvatarUrl ?? null)
    const [imagesLoaded, setImagesLoaded] = useState<Record<string, boolean>>({})

    function handleImageLoad(key: string) {
        setImagesLoaded(prev => ({ ...prev, [key]: true }))
    }

    const avatars = SEEDS.map(seed => ({
        seed,
        url: dicebearUrl(activeStyle, seed),
        key: `${activeStyle}-${seed}`,
    }))

    return (
        <Modal
            isOpen={isOpen}
            onClose={onClose}
            title="Choose your avatar"
            size="lg"
            confirmLabel="Save avatar"
            onConfirm={() => selected && onSelect(selected)}
            isLoading={isSaving}
        >
            <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>

                {/* Style tabs */}
                <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                    {DICEBEAR_STYLES.map(style => (
                        <button
                            key={style.id}
                            onClick={() => {
                                setActiveStyle(style.id)
                                setImagesLoaded({})
                            }}
                            style={{
                                padding: '5px 12px', fontSize: '12px',
                                fontFamily: 'var(--font-body)',
                                fontWeight: activeStyle === style.id ? 600 : 400,
                                cursor: 'pointer', border: '1px solid',
                                borderColor: activeStyle === style.id ? 'var(--color-primary)' : 'var(--color-border)',
                                borderRadius: '999px',
                                background: activeStyle === style.id
                                    ? 'color-mix(in srgb, var(--color-primary) 12%, transparent)'
                                    : 'var(--color-surface)',
                                color: activeStyle === style.id ? 'var(--color-primary)' : 'var(--color-text-muted)',
                                transition: 'all 0.15s ease',
                            }}
                        >
                            {style.label}
                        </button>
                    ))}
                </div>

                {/* Avatar grid */}
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gap: '10px' }}>
                    {avatars.map(({ seed, url, key }) => {
                        const isSelected = selected === url
                        const isLoaded = imagesLoaded[key]
                        return (
                            <button
                                key={key}
                                onClick={() => setSelected(url)}
                                title={seed}
                                style={{
                                    padding: '4px', border: '2px solid',
                                    borderColor: isSelected ? 'var(--color-primary)' : 'var(--color-border)',
                                    borderRadius: '50%', cursor: 'pointer',
                                    background: isSelected
                                        ? 'color-mix(in srgb, var(--color-primary) 10%, transparent)'
                                        : 'var(--color-surface)',
                                    transition: 'all 0.15s ease',
                                    aspectRatio: '1',
                                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    position: 'relative',
                                    boxShadow: isSelected ? '0 0 0 2px var(--color-primary)' : 'none',
                                }}
                            >
                                {!isLoaded && (
                                    <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                        <Spinner size="sm" />
                                    </div>
                                )}
                                <img
                                    src={url}
                                    alt={seed}
                                    onLoad={() => handleImageLoad(key)}
                                    style={{
                                        width: '100%', height: '100%', borderRadius: '50%',
                                        opacity: isLoaded ? 1 : 0,
                                        transition: 'opacity 0.2s ease',
                                    }}
                                />
                                {isSelected && (
                                    <div style={{
                                        position: 'absolute', bottom: '-2px', right: '-2px',
                                        width: '18px', height: '18px', borderRadius: '50%',
                                        background: 'var(--color-primary)',
                                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                                        fontSize: '10px', color: 'white',
                                        border: '2px solid var(--color-surface)',
                                    }}>
                                        ✓
                                    </div>
                                )}
                            </button>
                        )
                    })}
                </div>

                {/* Preview of selected */}
                {selected && (
                    <div style={{
                        display: 'flex', alignItems: 'center', gap: '12px',
                        padding: '12px 16px',
                        background: 'color-mix(in srgb, var(--color-primary) 6%, transparent)',
                        border: '1px solid color-mix(in srgb, var(--color-primary) 20%, transparent)',
                        borderRadius: 'var(--border-radius)',
                    }}>
                        <img
                            src={selected}
                            alt="selected avatar"
                            style={{ width: 44, height: 44, borderRadius: '50%', border: '2px solid var(--color-primary)' }}
                        />
                        <div>
                            <p style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>
                                Selected avatar
                            </p>
                            <p style={{ margin: '2px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                This will replace your current avatar
                            </p>
                        </div>
                    </div>
                )}
            </div>
        </Modal>
    )
}

// ─── section: profile ─────────────────────────────────────────────────────────

function ProfileSection() {
    const { user: rawUser, setAuth, token } = useAuth()
    const user = rawUser as UserProfile | null
    const [username, setUsername] = useState(user?.username ?? '')
    const [bio, setBio] = useState(user?.bio ?? '')
    const [avatarPickerOpen, setAvatarPickerOpen] = useState(false)
    const [success, setSuccess] = useState('')
    const [error, setError] = useState('')

    const mutation = useMutation({
        mutationFn: (payload: UpdateProfilePayload) => updateProfile(payload),
        onSuccess: (updatedUser) => {
            if (token) setAuth(token, updatedUser)
            setSuccess('Profile updated successfully.')
            setError('')
        },
        onError: (err: unknown) => {
            const msg = err instanceof Error ? err.message : 'Failed to update profile.'
            if (msg.toLowerCase().includes('unique') || msg.toLowerCase().includes('username')) {
                setError('That username is already taken.')
            } else {
                setError(msg)
            }
            setSuccess('')
        },
    })

    const avatarMutation = useMutation({
        mutationFn: (avatarUrl: string) => updateProfile({ avatar_url: avatarUrl }),
        onSuccess: (updatedUser) => {
            if (token) setAuth(token, updatedUser)
            setAvatarPickerOpen(false)
            setSuccess('Avatar updated successfully.')
            setError('')
        },
        onError: () => {
            setError('Failed to update avatar.')
            setSuccess('')
        },
    })

    function handleSave() {
        setSuccess('')
        setError('')
        if (!username.trim()) { setError('Username cannot be empty.'); return }
        if (bio.length > 500) { setError('Bio must be 500 characters or fewer.'); return }
        mutation.mutate({ username: username.trim(), bio: bio.trim() || undefined })
    }

    const bioOver = bio.length > 500

    return (
        <>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>

                {/* Avatar row */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
                    <Avatar name={user?.username ?? '?'} avatarUrl={user?.avatar_url} size="lg" />
                    <div>
                        <p style={{ margin: '0 0 8px', fontSize: '13px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                            Choose from illustrated avatar styles
                        </p>
                        <button
                            onClick={() => setAvatarPickerOpen(true)}
                            style={{
                                padding: '6px 14px', fontSize: '13px',
                                fontFamily: 'var(--font-body)', fontWeight: 500,
                                cursor: 'pointer',
                                background: 'var(--color-surface)',
                                border: '1px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)',
                                color: 'var(--color-text)',
                                transition: 'background 0.15s',
                            }}
                            onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-border)')}
                            onMouseLeave={e => (e.currentTarget.style.background = 'var(--color-surface)')}
                        >
                            Change avatar
                        </button>
                    </div>
                </div>

                {/* Username */}
                <Input label="Username" type="text" value={username} onChange={setUsername} placeholder="your_username" />

                {/* Bio */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                    <label style={{ fontSize: '14px', fontWeight: 500, color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>
                        Bio
                    </label>
                    <textarea
                        value={bio}
                        onChange={e => setBio(e.target.value)}
                        placeholder="Tell other readers a bit about yourself…"
                        rows={4}
                        style={{
                            width: '100%', padding: '10px 12px', fontSize: '14px',
                            fontFamily: 'var(--font-body)', color: 'var(--color-text)',
                            background: 'var(--color-surface)',
                            border: `1px solid ${bioOver ? '#ef4444' : 'var(--color-border)'}`,
                            borderRadius: 'var(--border-radius)',
                            resize: 'vertical', outline: 'none', boxSizing: 'border-box',
                            lineHeight: 1.5, transition: 'border-color 0.15s',
                        }}
                    />
                    <p style={{
                        margin: 0, fontSize: '12px', fontFamily: 'var(--font-body)',
                        color: bioOver ? '#ef4444' : 'var(--color-text-muted)', textAlign: 'right',
                    }}>
                        {bio.length} / 500
                    </p>
                </div>

                {success && <SuccessBanner message={success} />}
                {error && <ErrorBanner message={error} />}

                <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <Button
                        label="Save changes"
                        onClick={handleSave}
                        disabled={mutation.isPending || bioOver}
                        isLoading={mutation.isPending}
                    />
                </div>
            </div>

            <AvatarPickerModal
                isOpen={avatarPickerOpen}
                onClose={() => setAvatarPickerOpen(false)}
                currentAvatarUrl={user?.avatar_url}
                onSelect={(url) => avatarMutation.mutate(url)}
                isSaving={avatarMutation.isPending}
            />
        </>
    )
}

// ─── section: email ───────────────────────────────────────────────────────────

function EmailSection() {
    const { user, setAuth, token } = useAuth()
    const [email, setEmail] = useState(user?.email ?? '')
    const [currentPassword, setCurrentPassword] = useState('')
    const [success, setSuccess] = useState('')
    const [error, setError] = useState('')

    const mutation = useMutation({
        mutationFn: (payload: UpdateEmailPayload) => updateEmail(payload),
        onSuccess: (updatedUser) => {
            if (token) setAuth(token, updatedUser)
            setCurrentPassword('')
            setSuccess('Email address updated.')
            setError('')
        },
        onError: () => {
            setError('Failed to update email. Check your password and try again.')
            setSuccess('')
        },
    })

    function handleSave() {
        setSuccess('')
        setError('')
        if (!email.trim()) { setError('Email cannot be empty.'); return }
        if (!currentPassword) { setError('Enter your current password to confirm this change.'); return }
        mutation.mutate({ email: email.trim(), current_password: currentPassword })
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
            <Input label="New email address" type="email" value={email} onChange={setEmail} placeholder="you@example.com" />
            <Input label="Current password" type="password" value={currentPassword} onChange={setCurrentPassword} placeholder="Confirm with your password" />

            {success && <SuccessBanner message={success} />}
            {error && <ErrorBanner message={error} />}

            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                <Button label="Update email" onClick={handleSave} disabled={mutation.isPending} isLoading={mutation.isPending} />
            </div>
        </div>
    )
}
// ─── section: password ────────────────────────────────────────────────────────

function PasswordSection() {
    const [current, setCurrent] = useState('')
    const [next, setNext] = useState('')
    const [confirm, setConfirm] = useState('')
    const [success, setSuccess] = useState('')
    const [error, setError] = useState('')

    const mutation = useMutation({
        mutationFn: (payload: UpdatePasswordPayload) => updatePassword(payload),
        onSuccess: () => {
            setCurrent(''); setNext(''); setConfirm('')
            setSuccess('Password changed successfully.')
            setError('')
        },
        onError: () => {
            setError('Failed to change password. Check your current password and try again.')
            setSuccess('')
        },
    })

    function strength(pw: string): { level: number; label: string; color: string } {
        if (!pw) return { level: 0, label: '', color: 'transparent' }
        let score = 0
        if (pw.length >= 8) score++
        if (pw.length >= 12) score++
        if (/[A-Z]/.test(pw)) score++
        if (/[0-9]/.test(pw)) score++
        if (/[^A-Za-z0-9]/.test(pw)) score++
        if (score <= 1) return { level: 1, label: 'Weak', color: '#ef4444' }
        if (score <= 3) return { level: 2, label: 'Fair', color: '#f59e0b' }
        return { level: 3, label: 'Strong', color: '#22c55e' }
    }

    const pw = strength(next)

    function handleSave() {
        setSuccess(''); setError('')
        if (!current) { setError('Enter your current password.'); return }
        if (next.length < 8) { setError('New password must be at least 8 characters.'); return }
        if (next !== confirm) { setError('Passwords do not match.'); return }
        mutation.mutate({ current_password: current, new_password: next })
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
            <Input label="Current password" type="password" value={current} onChange={setCurrent} placeholder="Your existing password" />
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <Input label="New password" type="password" value={next} onChange={setNext} placeholder="At least 8 characters" />
                {next && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                        <div style={{ flex: 1, height: 4, borderRadius: 2, background: 'var(--color-border)', overflow: 'hidden' }}>
                            <div style={{
                                height: '100%', width: `${(pw.level / 3) * 100}%`,
                                background: pw.color, borderRadius: 2,
                                transition: 'width 0.3s ease, background 0.3s ease',
                            }} />
                        </div>
                        <span style={{ fontSize: '12px', fontFamily: 'var(--font-body)', color: pw.color, minWidth: 40 }}>
                            {pw.label}
                        </span>
                    </div>
                )}
            </div>
            <Input label="Confirm new password" type="password" value={confirm} onChange={setConfirm} placeholder="Repeat new password" />

            {success && <SuccessBanner message={success} />}
            {error && <ErrorBanner message={error} />}

            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                <Button label="Change password" onClick={handleSave} disabled={mutation.isPending} isLoading={mutation.isPending} />
            </div>
        </div>
    )
}

// ─── main page ────────────────────────────────────────────────────────────────

export default function ProfilePage() {
    const { user: rawUser } = useAuth()
    const user = rawUser as UserProfile | null
    const [section, setSection] = useState<Section>('profile')

    if (!user) return null

    return (
        <div className="min-h-screen" style={{ padding: '40px 16px' }}>
            <div style={{ maxWidth: 640, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: '32px' }}>

                {/* Page header */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
                    <Avatar name={user.username} avatarUrl={user.avatar_url} size="lg" />
                    <div>
                        <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 700, fontFamily: 'var(--font-heading)', color: 'var(--color-text)' }}>
                            {user.username}
                        </h1>
                        <p style={{ margin: '4px 0 0', fontSize: '14px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                            {user.email}
                        </p>
                        {user.role !== 'user' && (
                            <span style={{
                                display: 'inline-block', marginTop: '6px', padding: '2px 8px',
                                background: 'var(--color-primary)', color: 'var(--color-primary-text)',
                                borderRadius: '999px', fontSize: '11px', fontWeight: 600,
                                textTransform: 'uppercase', fontFamily: 'var(--font-body)', letterSpacing: '0.05em',
                            }}>
                                {user.role}
                            </span>
                        )}
                    </div>
                </div>

                {/* Card */}
                <div style={{
                    background: 'var(--color-surface)', border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)', boxShadow: 'var(--shadow-lg)', overflow: 'hidden',
                }}>
                    <div style={{ display: 'flex', borderBottom: '1px solid var(--color-border)', padding: '0 8px', overflowX: 'auto' }}>
                        <SectionTab id="profile"  label="Profile"  active={section} onClick={setSection} />
                        <SectionTab id="email"    label="Email"    active={section} onClick={setSection} />
                        <SectionTab id="password" label="Password" active={section} onClick={setSection} />
                    </div>
                    <div style={{ padding: '28px 28px 32px' }}>
                        {section === 'profile'  && <ProfileSection />}
                        {section === 'email'    && <EmailSection />}
                        {section === 'password' && <PasswordSection />}
                    </div>
                </div>

                <p style={{ margin: 0, fontSize: '12px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', textAlign: 'center' }}>
                    Member since {new Date(user.created_at).toLocaleDateString('en-GB', { year: 'numeric', month: 'long', day: 'numeric' })}
                </p>
            </div>
        </div>
    )
}