import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { register } from '../api/auth'
import { useAuth } from '../context/AuthContext'
import { Input, Button } from '../components'
import type { AuthResponse } from '../types'

export default function RegisterPage() {
    const [email, setEmail] = useState<string>('')
    const [username, setUsername] = useState<string>('')
    const [password, setPassword] = useState<string>('')
    const [confirm, setConfirm] = useState<string>('')
    const [errors, setErrors] = useState<Record<string, string>>({})

    const { setAuth } = useAuth()
    const navigate = useNavigate()

    function validate(): boolean {
        const e: Record<string, string> = {}

        if (!email) {
            e.email = 'Email is required'
        } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
            e.email = 'Enter a valid email address'
        }

        if (!username) {
            e.username = 'Username is required'
        } else if (username.length < 3) {
            e.username = 'Username must be at least 3 characters'
        } else if (!/^[a-zA-Z0-9_]+$/.test(username)) {
            e.username = 'Only letters, numbers and underscores allowed'
        }

        if (!password) {
            e.password = 'Password is required'
        } else if (password.length < 8) {
            e.password = 'Password must be at least 8 characters'
        }

        if (!confirm) {
            e.confirm = 'Please confirm your password'
        } else if (password !== confirm) {
            e.confirm = 'Passwords do not match'
        }

        setErrors(e)
        return Object.keys(e).length === 0
    }

    const mutation = useMutation({
        mutationFn: () => register({ email, username, password }),
        onSuccess: (data: AuthResponse) => {
            setAuth(data.token, data.user)
            navigate('/')
        },
        onError: (err: unknown) => {
            const message = (err as { response?: { data?: { error?: string } } })
                ?.response?.data?.error ?? 'Registration failed'

            if (message.toLowerCase().includes('email')) {
                setErrors(prev => ({ ...prev, email: 'This email is already registered' }))
            } else if (message.toLowerCase().includes('username')) {
                setErrors(prev => ({ ...prev, username: 'This username is already taken' }))
            } else {
                setErrors(prev => ({ ...prev, general: message }))
            }
        },
    })

    function handleSubmit(): void {
        if (!validate()) return
        mutation.mutate()
    }

    return (
        <div
            className="min-h-screen flex items-center justify-center p-4"
        >

            <div
                className="w-full max-w-md"
                style={{
                    background: 'var(--glass-bg)',
                    backdropFilter: 'var(--glass-blur)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)',
                    boxShadow: 'var(--shadow-lg)',
                    padding: '32px',
                }}
            >
                {/* Header */}
                <div className="mb-8">
                    <h1
                        className="text-2xl font-bold"
                        style={{
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-heading)',
                        }}
                    >
                        Create your account
                    </h1>
                    <p className="text-sm mt-1" style={{ color: 'var(--color-text-muted)' }}>
                        Join Biblios and start building your library
                    </p>
                </div>

                {/* Form */}
                <div className="flex flex-col gap-4">
                    <Input
                        label="Email"
                        type="email"
                        value={email}
                        onChange={v => { setEmail(v); setErrors(p => ({ ...p, email: '' })) }}
                        placeholder="you@example.com"
                        error={errors.email}
                    />

                    <Input
                        label="Username"
                        value={username}
                        onChange={v => { setUsername(v); setErrors(p => ({ ...p, username: '' })) }}
                        placeholder="your_username"
                        error={errors.username}
                    />

                    <Input
                        label="Password"
                        type="password"
                        value={password}
                        onChange={v => { setPassword(v); setErrors(p => ({ ...p, password: '' })) }}
                        placeholder="At least 8 characters"
                        error={errors.password}
                    />

                    <Input
                        label="Confirm password"
                        type="password"
                        value={confirm}
                        onChange={v => { setConfirm(v); setErrors(p => ({ ...p, confirm: '' })) }}
                        placeholder="Repeat your password"
                        error={errors.confirm}
                    />

                    {errors.general && (
                        <p
                            className="text-sm px-3 py-2"
                            style={{
                                color: 'var(--color-error)',
                                background: 'var(--color-surface-alt)',
                                borderRadius: 'var(--border-radius)',
                            }}
                        >
                            {errors.general}
                        </p>
                    )}

                    {/* Password strength indicator */}
                    {password.length > 0 && (
                        <div className="flex flex-col gap-1">
                            <div className="flex gap-1">
                                {[1, 2, 3, 4].map(level => (
                                    <div
                                        key={level}
                                        style={{
                                            flex: 1,
                                            height: '3px',
                                            borderRadius: '2px',
                                            background: passwordStrength(password) >= level
                                                ? strengthColor(passwordStrength(password))
                                                : 'var(--color-border)',
                                            transition: 'var(--transition)',
                                        }}
                                    />
                                ))}
                            </div>
                            <p className="text-xs" style={{ color: 'var(--color-text-muted)' }}>
                                {strengthLabel(passwordStrength(password))}
                            </p>
                        </div>
                    )}

                    <Button
                        label="Create account"
                        onClick={handleSubmit}
                        isLoading={mutation.isPending}
                        fullWidth
                    />
                </div>

                {/* Footer */}
                <p className="text-sm text-center mt-6" style={{ color: 'var(--color-text-muted)' }}>
                    Already have an account?{' '}
                    <Link
                        to="/login"
                        style={{ color: 'var(--color-primary)' }}
                        className="font-medium hover:underline"
                    >
                        Sign in
                    </Link>
                </p>
            </div>
        </div>
    )
}

// ─── Password strength helpers ────────────────────────────────────────────────

function passwordStrength(password: string): number {
    let score = 0
    if (password.length >= 8) score++
    if (password.length >= 12) score++
    if (/[A-Z]/.test(password) && /[a-z]/.test(password)) score++
    if (/[0-9]/.test(password)) score++
    if (/[^A-Za-z0-9]/.test(password)) score++
    return Math.min(4, score)
}

function strengthColor(strength: number): string {
    const colors = ['', '#ef4444', '#f97316', '#eab308', '#22c55e']
    return colors[strength] ?? '#22c55e'
}

function strengthLabel(strength: number): string {
    const labels = ['', 'Weak', 'Fair', 'Good', 'Strong']
    return labels[strength] ?? 'Strong'
}