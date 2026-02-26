import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import type { User } from '../types'
import { getMe } from '../api/auth'

interface AuthContextType {
    user: User | null
    token: string | null
    isLoading: boolean
    setAuth: (token: string, user: User) => void
    clearAuth: () => void
    isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType | null>(null)

interface AuthProviderProps {
    children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [user, setUser] = useState<User | null>(null)
    const [token, setToken] = useState<string | null>(localStorage.getItem('token'))
    const [isLoading, setIsLoading] = useState<boolean>(true)

    useEffect(() => {
        async function loadUser() {
            const storedToken: string | null = localStorage.getItem('token')
            if (storedToken) {
                try {
                    const me: User = await getMe()
                    setUser(me)
                    setToken(storedToken)
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
    }

    function clearAuth(): void {
        localStorage.removeItem('token')
        setToken(null)
        setUser(null)
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
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider')
    }
    return context
}