import { createContext, useContext } from 'react'
import type { User } from '../types'

export interface AuthContextType {
    user: User | null
    token: string | null
    isLoading: boolean
    setAuth: (token: string, user: User) => void
    clearAuth: () => void
    isAuthenticated: boolean
}

export const AuthContext = createContext<AuthContextType | null>(null)

export function useAuth(): AuthContextType {
    const context = useContext(AuthContext)
    if (!context) throw new Error('useAuth must be used within AuthProvider')
    return context
}
