import { createContext, useContext, useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { themes } from '../themes/themes'
import type { ThemeId } from '../themes/themes'

interface ThemeContextType {
    themeId: ThemeId
    setTheme: (id: ThemeId) => void
}

const ThemeContext = createContext<ThemeContextType | null>(null)

function applyTheme(id: ThemeId): void {
    const theme = themes[id]
    const root = document.documentElement

    root.style.setProperty('--color-background', theme.colors.background)
    root.style.setProperty('--color-surface', theme.colors.surface)
    root.style.setProperty('--color-surface-alt', theme.colors.surfaceAlt)
    root.style.setProperty('--color-border', theme.colors.border)
    root.style.setProperty('--color-text', theme.colors.text)
    root.style.setProperty('--color-text-muted', theme.colors.textMuted)
    root.style.setProperty('--color-primary', theme.colors.primary)
    root.style.setProperty('--color-primary-hover', theme.colors.primaryHover)
    root.style.setProperty('--color-primary-text', theme.colors.primaryText)
    root.style.setProperty('--color-accent', theme.colors.accent)
    root.style.setProperty('--color-error', theme.colors.error)
    root.style.setProperty('--color-success', theme.colors.success)
    root.style.setProperty('--color-shadow', theme.colors.shadow)
    root.style.setProperty('--border-radius', theme.effects.borderRadius)
    root.style.setProperty('--shadow-sm', theme.effects.shadowSm)
    root.style.setProperty('--shadow-md', theme.effects.shadowMd)
    root.style.setProperty('--shadow-lg', theme.effects.shadowLg)
    root.style.setProperty('--transition', theme.effects.transition)
    root.style.setProperty('--input-bg', theme.effects.inputBg)
    root.style.setProperty('--glass-bg', theme.effects.glassBg)
    root.style.setProperty('--glass-blur', theme.effects.glassBlur)
    root.style.setProperty('--font-heading', theme.fonts.heading)
    root.style.setProperty('--font-body', theme.fonts.body)

    // Load Google Fonts dynamically
    const fontId = `theme-font-${id}`
    if (!document.getElementById(fontId)) {
        const link = document.createElement('link')
        link.id = fontId
        link.rel = 'stylesheet'
        const fonts = [theme.fonts.heading, theme.fonts.body]
            .map(f => f.replace(/'/g, '').split(',')[0].trim())
            .filter((v, i, a) => a.indexOf(v) === i)
            .map(f => f.replace(/ /g, '+'))
            .join('&family=')
        link.href = `https://fonts.googleapis.com/css2?family=${fonts}&display=swap`
        document.head.appendChild(link)
    }
    root.setAttribute('data-theme', id)
}

export function ThemeProvider({ children }: { children: ReactNode }) {
    // Default to woody for unauthenticated pages
    // When user logs in, AuthContext will call setTheme with their saved theme
    const stored = localStorage.getItem('theme') as ThemeId | null
    const [themeId, setThemeId] = useState<ThemeId>(stored ?? 'woody')

    useEffect(() => {
        applyTheme(themeId)
    }, [themeId])

    function setTheme(id: ThemeId): void {
        setThemeId(id)
        localStorage.setItem('theme', id)
    }

    return (
        <ThemeContext.Provider value={{ themeId, setTheme }}>
            {children}
        </ThemeContext.Provider>
    )
}

export function useTheme(): ThemeContextType {
    const context = useContext(ThemeContext)
    if (!context) throw new Error('useTheme must be used within ThemeProvider')
    return context
}