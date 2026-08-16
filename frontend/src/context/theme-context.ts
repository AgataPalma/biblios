import { createContext, useContext } from 'react'
import type { ThemeId } from '../themes/themes'

export interface ThemeContextType {
    themeId: ThemeId
    setTheme: (id: ThemeId) => void
}

export const ThemeContext = createContext<ThemeContextType | null>(null)

export function useTheme(): ThemeContextType {
    const context = useContext(ThemeContext)
    if (!context) throw new Error('useTheme must be used within ThemeProvider')
    return context
}
