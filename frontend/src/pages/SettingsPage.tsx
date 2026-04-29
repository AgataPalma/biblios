import { useState, useRef, useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import { useTheme } from '../context/ThemeContext'
import { themes } from '../themes/themes'
import type { ThemeId } from '../themes/themes'
import { updateProfile, updateEmail, updatePassword } from '../api/users'
import apiClient from '../api/client'
import { Button, Input, Spinner } from '../components'

type Tab = 'profile' | 'account' | 'theme' | 'import-export'

const TABS: { id: Tab; label: string }[] = [
    { id: 'profile', label: 'Profile' },
    { id: 'account', label: 'Account' },
    { id: 'theme', label: 'Theme' },
    { id: 'import-export', label: 'Import/Export' },
]

// ── Profile Tab ───────────────────────────────────────────────────────────────

function ProfileTab() {
    const { user, token, setAuth } = useAuth()
    const [username, setUsername] = useState(user?.username ?? '')
    const [bio, setBio] = useState(user?.bio ?? '')
    const [avatarUrl, setAvatarUrl] = useState(user?.avatar_url ?? '')
    const [successMsg, setSuccessMsg] = useState('')
    const [errorMsg, setErrorMsg] = useState('')

    const mutation = useMutation({
        mutationFn: () => updateProfile({ username, bio, avatar_url: avatarUrl }),
        onSuccess: (updatedUser) => {
            setAuth(token!, updatedUser)
            setSuccessMsg('Profile updated successfully.')
            setErrorMsg('')
        },
        onError: (err: unknown) => {
            const msg =
                err instanceof Error ? err.message : 'Failed to update profile.'
            setErrorMsg(msg)
            setSuccessMsg('')
        },
    })

    function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        setSuccessMsg('')
        setErrorMsg('')
        mutation.mutate()
    }

    return (
        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <Input
                label="Username"
                value={username}
                onChange={setUsername}
                placeholder="Your username"
            />

            {/* Bio textarea — custom because Input doesn't support textarea */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                <label
                    style={{
                        fontSize: '14px',
                        fontWeight: 500,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-body)',
                    }}
                >
                    Bio
                </label>
                <textarea
                    value={bio}
                    onChange={(e) => setBio(e.target.value)}
                    maxLength={500}
                    rows={4}
                    placeholder="Tell us about yourself…"
                    style={{
                        background: 'var(--input-bg)',
                        border: '1px solid var(--color-border)',
                        borderRadius: 'var(--border-radius)',
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-body)',
                        fontSize: '14px',
                        padding: '8px 12px',
                        resize: 'vertical',
                        outline: 'none',
                        transition: 'var(--transition)',
                        width: '100%',
                        boxSizing: 'border-box',
                    }}
                />
                <span
                    style={{
                        fontSize: '12px',
                        color: bio.length >= 480 ? 'var(--color-error)' : 'var(--color-text-muted)',
                        textAlign: 'right',
                        fontFamily: 'var(--font-body)',
                    }}
                >
                    {bio.length} / 500
                </span>
            </div>

            <Input
                label="Avatar URL"
                value={avatarUrl}
                onChange={setAvatarUrl}
                placeholder="https://example.com/avatar.png"
            />

            {successMsg && (
                <p style={{ color: 'var(--color-success)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                    {successMsg}
                </p>
            )}
            {errorMsg && (
                <p style={{ color: 'var(--color-error)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                    {errorMsg}
                </p>
            )}

            <div>
                <Button
                    label="Save Profile"
                    type="submit"
                    isLoading={mutation.isPending}
                />
            </div>
        </form>
    )
}

// ── Account Tab ───────────────────────────────────────────────────────────────

function AccountTab() {
    // Email form state
    const [newEmail, setNewEmail] = useState('')
    const [emailPassword, setEmailPassword] = useState('')
    const [emailSuccess, setEmailSuccess] = useState('')
    const [emailError, setEmailError] = useState('')

    // Password form state
    const [currentPassword, setCurrentPassword] = useState('')
    const [newPassword, setNewPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [passwordSuccess, setPasswordSuccess] = useState('')
    const [passwordError, setPasswordError] = useState('')

    const emailMutation = useMutation({
        mutationFn: () => updateEmail({ email: newEmail, current_password: emailPassword }),
        onSuccess: () => {
            setEmailSuccess('Email updated successfully.')
            setEmailError('')
            setNewEmail('')
            setEmailPassword('')
        },
        onError: (err: unknown) => {
            const msg = err instanceof Error ? err.message : 'Failed to update email.'
            setEmailError(msg)
            setEmailSuccess('')
        },
    })

    const passwordMutation = useMutation({
        mutationFn: () => updatePassword({ current_password: currentPassword, new_password: newPassword }),
        onSuccess: () => {
            setPasswordSuccess('Password updated successfully.')
            setPasswordError('')
            setCurrentPassword('')
            setNewPassword('')
            setConfirmPassword('')
        },
        onError: (err: unknown) => {
            const msg = err instanceof Error ? err.message : 'Failed to update password.'
            setPasswordError(msg)
            setPasswordSuccess('')
        },
    })

    function handleEmailSubmit(e: React.FormEvent) {
        e.preventDefault()
        setEmailSuccess('')
        setEmailError('')
        emailMutation.mutate()
    }

    function handlePasswordSubmit(e: React.FormEvent) {
        e.preventDefault()
        setPasswordSuccess('')
        setPasswordError('')
        if (newPassword !== confirmPassword) {
            setPasswordError('New passwords do not match.')
            return
        }
        passwordMutation.mutate()
    }

    const sectionStyle: React.CSSProperties = {
        background: 'var(--color-surface)',
        border: '1px solid var(--color-border)',
        borderRadius: 'var(--border-radius)',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
    }

    const sectionHeadingStyle: React.CSSProperties = {
        margin: 0,
        fontSize: '15px',
        fontWeight: 600,
        color: 'var(--color-text)',
        fontFamily: 'var(--font-heading)',
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
            {/* Email change */}
            <section style={sectionStyle}>
                <h2 style={sectionHeadingStyle}>Change Email</h2>
                <form onSubmit={handleEmailSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Input
                        label="New Email"
                        type="email"
                        value={newEmail}
                        onChange={setNewEmail}
                        placeholder="new@example.com"
                    />
                    <Input
                        label="Current Password"
                        type="password"
                        value={emailPassword}
                        onChange={setEmailPassword}
                        placeholder="Enter your current password"
                    />
                    {emailSuccess && (
                        <p style={{ color: 'var(--color-success)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                            {emailSuccess}
                        </p>
                    )}
                    {emailError && (
                        <p style={{ color: 'var(--color-error)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                            {emailError}
                        </p>
                    )}
                    <div>
                        <Button
                            label="Update Email"
                            type="submit"
                            isLoading={emailMutation.isPending}
                        />
                    </div>
                </form>
            </section>

            {/* Password change */}
            <section style={sectionStyle}>
                <h2 style={sectionHeadingStyle}>Change Password</h2>
                <form onSubmit={handlePasswordSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Input
                        label="Current Password"
                        type="password"
                        value={currentPassword}
                        onChange={setCurrentPassword}
                        placeholder="Enter your current password"
                    />
                    <Input
                        label="New Password"
                        type="password"
                        value={newPassword}
                        onChange={setNewPassword}
                        placeholder="Enter a new password"
                    />
                    <Input
                        label="Confirm New Password"
                        type="password"
                        value={confirmPassword}
                        onChange={setConfirmPassword}
                        placeholder="Repeat your new password"
                    />
                    {passwordSuccess && (
                        <p style={{ color: 'var(--color-success)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                            {passwordSuccess}
                        </p>
                    )}
                    {passwordError && (
                        <p style={{ color: 'var(--color-error)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                            {passwordError}
                        </p>
                    )}
                    <div>
                        <Button
                            label="Update Password"
                            type="submit"
                            isLoading={passwordMutation.isPending}
                        />
                    </div>
                </form>
            </section>
        </div>
    )
}

// ── Theme Tab ─────────────────────────────────────────────────────────────────

function ThemeTab() {
    const { themeId, setTheme } = useTheme()

    async function handleThemeSelect(id: ThemeId) {
        setTheme(id)
        try {
            await apiClient.put('/users/me/theme', { theme: id })
        } catch {
            // Non-critical — theme is already applied locally
        }
    }

    const themeList = Object.values(themes)

    return (
        <div>
            <p style={{
                margin: '0 0 20px',
                fontSize: '14px',
                color: 'var(--color-text-muted)',
                fontFamily: 'var(--font-body)',
            }}>
                Choose a theme for your Biblios experience.
            </p>
            <div
                style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))',
                    gap: '12px',
                }}
            >
                {themeList.map((t) => {
                    const isActive = t.id === themeId
                    return (
                        <button
                            key={t.id}
                            onClick={() => handleThemeSelect(t.id)}
                            style={{
                                background: 'var(--color-surface)',
                                border: isActive
                                    ? '2px solid var(--color-primary)'
                                    : '2px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)',
                                padding: '16px 12px',
                                cursor: 'pointer',
                                display: 'flex',
                                flexDirection: 'column',
                                alignItems: 'center',
                                gap: '8px',
                                transition: 'var(--transition)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            <span style={{ fontSize: '28px', lineHeight: 1 }}>{t.emoji}</span>
                            <span
                                style={{
                                    fontSize: '13px',
                                    fontWeight: isActive ? 600 : 400,
                                    color: isActive ? 'var(--color-primary)' : 'var(--color-text)',
                                }}
                            >
                                {t.name}
                            </span>
                            <span
                                style={{
                                    fontSize: '11px',
                                    color: 'var(--color-text-muted)',
                                    textAlign: 'center',
                                    lineHeight: 1.3,
                                }}
                            >
                                {t.description}
                            </span>
                        </button>
                    )
                })}
            </div>
        </div>
    )
}

// ── Import/Export Tab ─────────────────────────────────────────────────────────

type ImportStatus = 'idle' | 'pending' | 'processing' | 'completed' | 'failed'

function ImportExportTab() {
    const fileInputRef = useRef<HTMLInputElement>(null)
    const [importStatus, setImportStatus] = useState<ImportStatus>('idle')
    const [importError, setImportError] = useState('')
    const [exportError, setExportError] = useState('')
    const [isExporting, setIsExporting] = useState(false)
    const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

    // Clean up polling on unmount
    useEffect(() => {
        return () => {
            if (pollRef.current) clearInterval(pollRef.current)
        }
    }, [])

    async function handleImport() {
        const file = fileInputRef.current?.files?.[0]
        if (!file) {
            setImportError('Please select a CSV file first.')
            return
        }

        setImportError('')
        setImportStatus('pending')

        try {
            const formData = new FormData()
            formData.append('file', file)

            const response = await apiClient.post<{ id: string; status: string }>(
                '/import/goodreads',
                formData,
                { headers: { 'Content-Type': 'multipart/form-data' } }
            )

            const jobId = response.data.id
            setImportStatus('processing')

            // Poll every 3 seconds
            pollRef.current = setInterval(async () => {
                try {
                    const jobRes = await apiClient.get<{ id: string; status: string }>(
                        `/import/jobs/${jobId}`
                    )
                    const status = jobRes.data.status

                    if (status === 'completed') {
                        setImportStatus('completed')
                        if (pollRef.current) clearInterval(pollRef.current)
                    } else if (status === 'failed') {
                        setImportStatus('failed')
                        setImportError('Import job failed. Please try again.')
                        if (pollRef.current) clearInterval(pollRef.current)
                    }
                    // 'pending' | 'processing' — keep polling
                } catch {
                    setImportStatus('failed')
                    setImportError('Failed to check import status.')
                    if (pollRef.current) clearInterval(pollRef.current)
                }
            }, 3000)
        } catch (err: unknown) {
            const msg = err instanceof Error ? err.message : 'Import failed.'
            setImportError(msg)
            setImportStatus('failed')
        }
    }

    async function handleExport() {
        setExportError('')
        setIsExporting(true)
        try {
            const response = await apiClient.get('/export/library', { responseType: 'blob' })
            const url = window.URL.createObjectURL(new Blob([response.data]))
            const link = document.createElement('a')
            link.href = url
            link.setAttribute('download', 'biblios-library.csv')
            document.body.appendChild(link)
            link.click()
            link.remove()
            window.URL.revokeObjectURL(url)
        } catch (err: unknown) {
            const msg = err instanceof Error ? err.message : 'Export failed.'
            setExportError(msg)
        } finally {
            setIsExporting(false)
        }
    }

    function resetImport() {
        setImportStatus('idle')
        setImportError('')
        if (fileInputRef.current) fileInputRef.current.value = ''
        if (pollRef.current) clearInterval(pollRef.current)
    }

    const statusLabels: Record<ImportStatus, string> = {
        idle: '',
        pending: 'Upload received, waiting to start…',
        processing: 'Processing your library…',
        completed: 'Import completed successfully!',
        failed: '',
    }

    const sectionStyle: React.CSSProperties = {
        background: 'var(--color-surface)',
        border: '1px solid var(--color-border)',
        borderRadius: 'var(--border-radius)',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
    }

    const sectionHeadingStyle: React.CSSProperties = {
        margin: 0,
        fontSize: '15px',
        fontWeight: 600,
        color: 'var(--color-text)',
        fontFamily: 'var(--font-heading)',
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
            {/* Import */}
            <section style={sectionStyle}>
                <h2 style={sectionHeadingStyle}>Import from Goodreads</h2>
                <p style={{ margin: 0, fontSize: '14px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                    Export your Goodreads library as a CSV and upload it here.
                </p>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    <input
                        ref={fileInputRef}
                        type="file"
                        accept=".csv"
                        disabled={importStatus === 'pending' || importStatus === 'processing'}
                        style={{
                            fontSize: '14px',
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-body)',
                            cursor: 'pointer',
                        }}
                    />

                    {importStatus === 'idle' || importStatus === 'failed' ? (
                        <div>
                            <Button
                                label="Import from Goodreads"
                                onClick={handleImport}
                            />
                        </div>
                    ) : importStatus === 'completed' ? (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <span style={{ color: 'var(--color-success)', fontSize: '14px', fontFamily: 'var(--font-body)' }}>
                                ✓ {statusLabels.completed}
                            </span>
                            <Button label="Import another" variant="secondary" onClick={resetImport} />
                        </div>
                    ) : (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <Spinner size="sm" />
                            <span style={{ fontSize: '14px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                {statusLabels[importStatus]}
                            </span>
                        </div>
                    )}

                    {importError && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <p style={{ color: 'var(--color-error)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                                {importError}
                            </p>
                            {importStatus === 'failed' && (
                                <Button label="Try again" variant="secondary" onClick={resetImport} />
                            )}
                        </div>
                    )}
                </div>
            </section>

            {/* Export */}
            <section style={sectionStyle}>
                <h2 style={sectionHeadingStyle}>Export Library</h2>
                <p style={{ margin: 0, fontSize: '14px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                    Download your entire library as a CSV file.
                </p>

                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <Button
                        label={isExporting ? 'Exporting…' : 'Export as CSV'}
                        onClick={handleExport}
                        isLoading={isExporting}
                    />
                </div>

                {exportError && (
                    <p style={{ color: 'var(--color-error)', fontSize: '14px', margin: 0, fontFamily: 'var(--font-body)' }}>
                        {exportError}
                    </p>
                )}
            </section>
        </div>
    )
}

// ── Settings Page ─────────────────────────────────────────────────────────────

export default function SettingsPage() {
    const [activeTab, setActiveTab] = useState<Tab>('profile')

    return (
        <div
            style={{
                maxWidth: '640px',
                margin: '0 auto',
                padding: '32px 16px',
                fontFamily: 'var(--font-body)',
            }}
        >
            {/* Page header */}
            <h1
                style={{
                    margin: '0 0 28px',
                    fontSize: '28px',
                    fontWeight: 700,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-heading)',
                }}
            >
                Settings
            </h1>

            {/* Tab navigation */}
            <nav
                style={{
                    display: 'flex',
                    borderBottom: '1px solid var(--color-border)',
                    marginBottom: '28px',
                    gap: '0',
                }}
                role="tablist"
            >
                {TABS.map((tab) => {
                    const isActive = tab.id === activeTab
                    return (
                        <button
                            key={tab.id}
                            role="tab"
                            aria-selected={isActive}
                            onClick={() => setActiveTab(tab.id)}
                            style={{
                                background: 'none',
                                border: 'none',
                                borderBottom: isActive
                                    ? '2px solid var(--color-primary)'
                                    : '2px solid transparent',
                                padding: '10px 16px',
                                cursor: 'pointer',
                                fontSize: '14px',
                                fontWeight: isActive ? 600 : 400,
                                color: isActive ? 'var(--color-primary)' : 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                                transition: 'var(--transition)',
                                marginBottom: '-1px',
                            }}
                        >
                            {tab.label}
                        </button>
                    )
                })}
            </nav>

            {/* Tab content */}
            <div role="tabpanel">
                {activeTab === 'profile' && <ProfileTab />}
                {activeTab === 'account' && <AccountTab />}
                {activeTab === 'theme' && <ThemeTab />}
                {activeTab === 'import-export' && <ImportExportTab />}
            </div>
        </div>
    )
}
