import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import { logout } from '../api/auth'
import { Avatar, ThemeSwitcher } from '../components'

interface NavItem {
    label: string
    path: string
    icon: string
}

const navItems: NavItem[] = [
    { label: 'Dashboard', path: '/', icon: '⊞' },
    { label: 'Books', path: '/books', icon: '📖' },
    { label: 'My Library', path: '/library', icon: '🗄️' },
]

export default function Navbar() {
    const { user, clearAuth } = useAuth()
    const location = useLocation()
    const navigate = useNavigate()
    const [menuOpen, setMenuOpen] = useState<boolean>(false)

    const logoutMutation = useMutation({
        mutationFn: logout,
        onSettled: () => {
            clearAuth()
            navigate('/login')
        },
    })

    function isActive(path: string): boolean {
        if (path === '/') return location.pathname === '/'
        return location.pathname.startsWith(path)
    }

    return (
        <nav
            style={{
                position: 'fixed',
                top: 0,
                left: 0,
                right: 0,
                zIndex: 100,
                height: '56px',
                display: 'flex',
                alignItems: 'center',
                padding: '0 24px',
                background: 'var(--glass-bg)',
                backdropFilter: 'var(--glass-blur)',
                borderBottom: '1px solid var(--color-border)',
                boxShadow: 'var(--shadow-sm)',
                gap: '8px',
            }}
        >
            {/* Logo */}
            <Link
                to="/"
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    textDecoration: 'none',
                    marginRight: '24px',
                    flexShrink: 0,
                }}
            >
                <span style={{ fontSize: '20px' }}>📚</span>
                <span
                    style={{
                        fontSize: '18px',
                        fontWeight: 700,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                        letterSpacing: '0.02em',
                    }}
                >
          Biblios
        </span>
            </Link>

            {/* Nav items — desktop */}
            <div
                className="hidden sm:flex"
                style={{ display: 'flex', gap: '4px', flex: 1 }}
            >
                {navItems.map(item => (
                    <Link
                        key={item.path}
                        to={item.path}
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '6px',
                            padding: '6px 12px',
                            borderRadius: 'var(--border-radius)',
                            textDecoration: 'none',
                            fontSize: '13px',
                            fontWeight: isActive(item.path) ? 600 : 400,
                            color: isActive(item.path) ? 'var(--color-primary)' : 'var(--color-text-muted)',
                            background: isActive(item.path) ? 'var(--color-surface-alt)' : 'transparent',
                            transition: 'var(--transition)',
                        }}
                        onMouseEnter={e => {
                            if (!isActive(item.path)) {
                                (e.currentTarget as HTMLAnchorElement).style.background = 'var(--color-surface-alt)'
                                ;(e.currentTarget as HTMLAnchorElement).style.color = 'var(--color-text)'
                            }
                        }}
                        onMouseLeave={e => {
                            if (!isActive(item.path)) {
                                (e.currentTarget as HTMLAnchorElement).style.background = 'transparent'
                                ;(e.currentTarget as HTMLAnchorElement).style.color = 'var(--color-text-muted)'
                            }
                        }}
                    >
                        <span>{item.icon}</span>
                        <span>{item.label}</span>
                    </Link>
                ))}
            </div>

            {/* Right side */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginLeft: 'auto' }}>
                <ThemeSwitcher />

                {/* User menu */}
                <div style={{ position: 'relative' }}>
                    <button
                        onClick={() => setMenuOpen(!menuOpen)}
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            background: 'none',
                            border: '1px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)',
                            padding: '4px 10px 4px 4px',
                            cursor: 'pointer',
                            transition: 'var(--transition)',
                        }}
                        onMouseEnter={e => {
                            (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-surface-alt)'
                        }}
                        onMouseLeave={e => {
                            (e.currentTarget as HTMLButtonElement).style.background = 'none'
                        }}
                    >
                        <Avatar name={user?.username ?? '?'} size="sm" />
                        <span style={{
                            fontSize: '13px',
                            fontWeight: 500,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-body)',
                        }}>
              {user?.username}
            </span>
                        <span style={{ color: 'var(--color-text-muted)', fontSize: '10px' }}>
              {menuOpen ? '▴' : '▾'}
            </span>
                    </button>

                    {/* Dropdown */}
                    {menuOpen && (
                        <>
                            <div
                                className="fixed inset-0 z-10"
                                onClick={() => setMenuOpen(false)}
                            />
                            <div
                                style={{
                                    position: 'absolute',
                                    top: 'calc(100% + 8px)',
                                    right: 0,
                                    minWidth: '200px',
                                    background: 'var(--color-surface)',
                                    border: '1px solid var(--color-border)',
                                    borderRadius: 'var(--border-radius)',
                                    boxShadow: 'var(--shadow-lg)',
                                    zIndex: 20,
                                    overflow: 'hidden',
                                }}
                            >
                                {/* User info */}
                                <div
                                    style={{
                                        padding: '12px 16px',
                                        borderBottom: '1px solid var(--color-border)',
                                    }}
                                >
                                    <p style={{
                                        margin: 0,
                                        fontSize: '13px',
                                        fontWeight: 600,
                                        color: 'var(--color-text)',
                                        fontFamily: 'var(--font-body)',
                                    }}>
                                        {user?.username}
                                    </p>
                                    <p style={{
                                        margin: '2px 0 0',
                                        fontSize: '11px',
                                        color: 'var(--color-text-muted)',
                                        fontFamily: 'var(--font-body)',
                                    }}>
                                        {user?.email}
                                    </p>
                                    {user?.role !== 'user' && (
                                        <span style={{
                                            display: 'inline-block',
                                            marginTop: '4px',
                                            padding: '1px 6px',
                                            background: 'var(--color-primary)',
                                            color: 'var(--color-primary-text)',
                                            borderRadius: '999px',
                                            fontSize: '10px',
                                            fontWeight: 600,
                                            textTransform: 'uppercase',
                                            fontFamily: 'var(--font-body)',
                                        }}>
                      {user?.role}
                    </span>
                                    )}
                                </div>

                                {/* Menu items */}
                                {[
                                    { label: 'My Library', path: '/library', icon: '🗄️' },
                                    { label: 'Settings', path: '/settings', icon: '⚙️' },
                                    ...(user?.role === 'admin' || user?.role === 'moderator'
                                            ? [{ label: 'Moderation', path: '/moderation', icon: '🛡️' }]
                                            : []
                                    ),
                                ].map(item => (
                                    <Link
                                        key={item.path}
                                        to={item.path}
                                        onClick={() => setMenuOpen(false)}
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '10px',
                                            padding: '10px 16px',
                                            textDecoration: 'none',
                                            fontSize: '13px',
                                            color: 'var(--color-text)',
                                            transition: 'var(--transition)',
                                            fontFamily: 'var(--font-body)',
                                        }}
                                        onMouseEnter={e => {
                                            (e.currentTarget as HTMLAnchorElement).style.background = 'var(--color-surface-alt)'
                                        }}
                                        onMouseLeave={e => {
                                            (e.currentTarget as HTMLAnchorElement).style.background = 'transparent'
                                        }}
                                    >
                                        <span>{item.icon}</span>
                                        <span>{item.label}</span>
                                    </Link>
                                ))}

                                {/* Divider + Logout */}
                                <div style={{ borderTop: '1px solid var(--color-border)' }}>
                                    <button
                                        onClick={() => {
                                            setMenuOpen(false)
                                            logoutMutation.mutate()
                                        }}
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '10px',
                                            padding: '10px 16px',
                                            width: '100%',
                                            background: 'none',
                                            border: 'none',
                                            cursor: 'pointer',
                                            fontSize: '13px',
                                            color: 'var(--color-error)',
                                            transition: 'var(--transition)',
                                            fontFamily: 'var(--font-body)',
                                            textAlign: 'left',
                                        }}
                                        onMouseEnter={e => {
                                            (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-surface-alt)'
                                        }}
                                        onMouseLeave={e => {
                                            (e.currentTarget as HTMLButtonElement).style.background = 'none'
                                        }}
                                    >
                                        <span>🚪</span>
                                        <span>{logoutMutation.isPending ? 'Logging out...' : 'Log out'}</span>
                                    </button>
                                </div>
                            </div>
                        </>
                    )}
                </div>
            </div>
        </nav>
    )
}