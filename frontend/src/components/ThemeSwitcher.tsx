import { useState } from 'react'
import { useTheme } from '../context/ThemeContext'
import { themes } from '../themes/themes'
import type { ThemeId } from '../themes/themes'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import apiClient from '../api/client'

export default function ThemeSwitcher() {
    const { themeId, setTheme } = useTheme()
    const { isAuthenticated } = useAuth()
    const [isOpen, setIsOpen] = useState<boolean>(false)

    const mutation = useMutation({
        mutationFn: (id: ThemeId) => apiClient.put('/users/me/theme', { theme: id }),
    })

    function handleThemeChange(id: ThemeId): void {
        setTheme(id)
        // Only save to backend if logged in
        if (isAuthenticated) {
            mutation.mutate(id)
        }
        setIsOpen(false)
    }

    return (
        <div className="relative">
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 px-3 py-2 rounded text-sm"
                style={{
                    background: 'var(--color-surface)',
                    border: '1px solid var(--color-border)',
                    color: 'var(--color-text)',
                    borderRadius: 'var(--border-radius)',
                    boxShadow: 'var(--shadow-sm)',
                }}
            >
                <span>{themes[themeId].emoji}</span>
                <span>{themes[themeId].name}</span>
                <span style={{ color: 'var(--color-text-muted)' }}>▾</span>
            </button>

            {isOpen && (
                <>
                    <div
                        className="fixed inset-0 z-10"
                        onClick={() => setIsOpen(false)}
                    />
                    <div
                        className="absolute right-0 mt-2 w-64 z-20 overflow-hidden"
                        style={{
                            background: 'var(--color-surface)',
                            border: '1px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)',
                            boxShadow: 'var(--shadow-lg)',
                        }}
                    >
                        <div className="p-2">
                            <p
                                className="text-xs font-medium px-2 py-1 mb-1"
                                style={{ color: 'var(--color-text-muted)' }}
                            >
                                Choose your theme
                            </p>
                            {Object.values(themes).map((theme) => (
                                <button
                                    key={theme.id}
                                    onClick={() => handleThemeChange(theme.id)}
                                    className="w-full flex items-center gap-3 px-3 py-2 text-sm text-left"
                                    style={{
                                        borderRadius: 'var(--border-radius)',
                                        background: themeId === theme.id ? 'var(--color-surface-alt)' : 'transparent',
                                        color: 'var(--color-text)',
                                    }}
                                >
                                    <span className="text-lg">{theme.emoji}</span>
                                    <div>
                                        <p className="font-medium">{theme.name}</p>
                                        <p className="text-xs" style={{ color: 'var(--color-text-muted)' }}>
                                            {theme.description}
                                        </p>
                                    </div>
                                    {themeId === theme.id && (
                                        <span className="ml-auto" style={{ color: 'var(--color-primary)' }}>✓</span>
                                    )}
                                </button>
                            ))}
                        </div>
                    </div>
                </>
            )}
        </div>
    )
}