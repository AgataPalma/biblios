import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import type { User } from '../types'
import { getMe } from '../api/auth'
import { useTheme } from './ThemeContext'
import type { ThemeId } from '../themes/themes'

interface AuthContextType {
    user: User | null
    token: string | null
    isLoading: boolean
    setAuth: (token: string, user: User) => void
    clearAuth: () => void
    isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [token, setToken] = useState<string | null>(localStorage.getItem('token'))
    const [isLoading, setIsLoading] = useState<boolean>(true)
    const { setTheme } = useTheme()

    useEffect(() => {
        async function loadUser() {
            const storedToken = localStorage.getItem('token')
            if (storedToken) {
                try {
                    const me: User = await getMe()
                    setUser(me)
                    setToken(storedToken)
                    // Apply user's saved theme
                    if (me.theme) {
                        setTheme(me.theme as ThemeId)
                        localStorage.setItem('theme', me.theme)
                    }
                } catch {
                    localStorage.removeItem('token')
                    setToken(null)
                    setUser(null)
                }
            }
            setIsLoading(false)
        }
        loadUser()
    }, [])

    function setAuth(newToken: string, newUser: User): void {
        localStorage.setItem('token', newToken)
        setToken(newToken)
        setUser(newUser)
        // Apply user's saved theme on login
        if (newUser.theme) {
            setTheme(newUser.theme as ThemeId)
            localStorage.setItem('theme', newUser.theme)
        }
    }

    function clearAuth(): void {
        localStorage.removeItem('token')
        localStorage.removeItem('theme')
        setToken(null)
        setUser(null)
        // Reset to woody on logout
        setTheme('woody')
    }

    return (
        <AuthContext.Provider value={{
            user,
            token,
            isLoading,
            setAuth,
            clearAuth,
            isAuthenticated: !!token && !!user,
        }}>
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth(): AuthContextType {
    const context = useContext(AuthContext)
    if (!context) throw new Error('useAuth must be used within AuthProvider')
    return context
}