import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { login } from '../api/auth'
import { useAuth } from '../context/AuthContext'
import Input from '../components/Input'
import Button from '../components/Button'
import ThemeSwitcher from '../components/ThemeSwitcher'
import type { AuthResponse } from '../types'

export default function LoginPage() {
    const [email, setEmail] = useState<string>('')
    const [password, setPassword] = useState<string>('')
    const [error, setError] = useState<string>('')

    const { setAuth } = useAuth()
    const navigate = useNavigate()

    const mutation = useMutation({
        mutationFn: () => login({ email, password }),
        onSuccess: (data: AuthResponse) => {
            setAuth(data.token, data.user)
            navigate('/')
        },
        onError: () => {
            setError('Invalid email or password')
        },
    })

    function handleSubmit(): void {
        setError('')
        if (!email || !password) {
            setError('Please fill in all fields')
            return
        }
        mutation.mutate()
    }

    return (
        <div
            className="min-h-screen flex items-center justify-center p-4"

        >
            {/* Theme switcher in top right */}
            <div className="fixed top-4 right-4">
                <ThemeSwitcher />
            </div>

            <div
                className="p-8 w-full max-w-md"
                style={{
                    background: 'var(--color-surface)',
                    borderRadius: 'var(--border-radius)',
                    boxShadow: 'var(--shadow-lg)',
                    border: '1px solid var(--color-border)',
                }}
            >
                <div className="mb-8">
                    <h1
                        className="text-2xl font-bold"
                        style={{ color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}
                    >
                        Welcome back
                    </h1>
                    <p className="text-sm mt-1" style={{ color: 'var(--color-text-muted)' }}>
                        Sign in to your Biblios account
                    </p>
                </div>

                <div className="flex flex-col gap-4">
                    <Input
                        label="Email"
                        type="email"
                        value={email}
                        onChange={setEmail}
                        placeholder="you@example.com"
                    />
                    <Input
                        label="Password"
                        type="password"
                        value={password}
                        onChange={setPassword}
                        placeholder="••••••••"
                    />

                    {error && (
                        <p
                            className="text-sm px-3 py-2 rounded"
                            style={{
                                color: 'var(--color-error)',
                                background: 'var(--color-surface-alt)',
                                borderRadius: 'var(--border-radius)',
                            }}
                        >
                            {error}
                        </p>
                    )}

                    <Button
                        label="Sign in"
                        onClick={handleSubmit}
                        isLoading={mutation.isPending}
                        fullWidth
                    />
                </div>

                <p className="text-sm text-center mt-6" style={{ color: 'var(--color-text-muted)' }}>
                    Don't have an account?{' '}
                    <Link
                        to="/register"
                        style={{ color: 'var(--color-primary)' }}
                        className="font-medium hover:underline"
                    >
                        Sign up
                    </Link>
                </p>
            </div>
        </div>
    )
}