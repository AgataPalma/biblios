import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import { useTheme } from '../context/ThemeContext'
import { logout } from '../api/auth'
import { Avatar } from '../components'
import { themes } from '../themes/themes'
import type { ThemeId } from '../themes/themes'
import apiClient from '../api/client'

// ── Collapse state is stored in localStorage so it persists across pages ──────
const COLLAPSED_KEY = 'sidebar_collapsed'

function getInitialCollapsed(): boolean {
    return localStorage.getItem(COLLAPSED_KEY) === 'true'
}

// ── The sidebar width is exposed as a CSS variable on <html> so every page
//    can use `padding-left: var(--sidebar-w)` without importing this component.
function setSidebarCssVar(collapsed: boolean) {
    document.documentElement.style.setProperty('--sidebar-w', collapsed ? '56px' : '220px')
}

interface NavItem {
    label: string
    path: string
    icon: string
    modOnly?: boolean
}

const NAV_ITEMS: NavItem[] = [
    { label: 'Dashboard',   path: '/',            icon: '⊞' },
    { label: 'Books',       path: '/books',        icon: '📖' },
    { label: 'My Library',  path: '/library',      icon: '🗄️' },
    { label: 'Submit Book', path: '/books/add',    icon: '➕' },
    { label: 'Moderation',  path: '/moderation',   icon: '🛡️', modOnly: true },
]

export default function Sidebar() {
    const { user, clearAuth, isAuthenticated } = useAuth()
    const { themeId, setTheme } = useTheme()
    const location = useLocation()
    const navigate = useNavigate()

    const [collapsed, setCollapsed] = useState<boolean>(() => {
        const initial = getInitialCollapsed()
        setSidebarCssVar(initial)
        return initial
    })
    const [userOpen, setUserOpen]   = useState(false)
    const [themeOpen, setThemeOpen] = useState(false)

    const logoutMutation = useMutation({
        mutationFn: logout,
        onSettled: () => {
            clearAuth()
            navigate('/login')
        },
    })

    const themeMutation = useMutation({
        mutationFn: (id: ThemeId) => apiClient.put('/users/me/theme', { theme: id }),
    })

    function handleCollapse() {
        const next = !collapsed
        setCollapsed(next)
        setSidebarCssVar(next)
        localStorage.setItem(COLLAPSED_KEY, String(next))
        // Close submenus when collapsing
        if (next) { setUserOpen(false); setThemeOpen(false) }
    }

    function handleThemeChange(id: ThemeId) {
        setTheme(id)
        if (isAuthenticated) themeMutation.mutate(id)
        setThemeOpen(false)
    }

    function isActive(path: string): boolean {
        if (path === '/') return location.pathname === '/'
        return location.pathname.startsWith(path)
    }

    const currentTheme = themes[themeId]
    const isMod = user?.role === 'moderator' || user?.role === 'admin'

    return (
        <aside
            style={{
                width: collapsed ? '56px' : '220px',
                flexShrink: 0,
                height: '100vh',
                position: 'fixed',
                top: 0,
                left: 0,
                zIndex: 100,
                background: 'var(--glass-bg)',
                backdropFilter: 'var(--glass-blur)',
                borderRight: '1px solid var(--color-border)',
                boxShadow: 'var(--shadow-sm)',
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
                transition: 'width 0.22s ease',
            }}
        >
            {/* ── Logo + collapse toggle ──────────────────────────────── */}
            <div
                style={{
                    padding: collapsed ? '16px 0' : '16px 16px',
                    borderBottom: '1px solid var(--color-border)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: collapsed ? 'center' : 'space-between',
                    gap: '8px',
                    flexShrink: 0,
                    minHeight: '56px',
                }}
            >
                {!collapsed && (
                    <Link
                        to="/"
                        style={{
                            display: 'flex', alignItems: 'center', gap: '8px',
                            textDecoration: 'none', overflow: 'hidden',
                        }}
                    >
                        <span style={{ fontSize: '20px', flexShrink: 0 }}>📚</span>
                        <span style={{
                            fontSize: '17px', fontWeight: 700,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-heading)',
                            letterSpacing: '0.02em',
                            whiteSpace: 'nowrap',
                        }}>
                            Biblios
                        </span>
                    </Link>
                )}

                {collapsed && (
                    <Link to="/" style={{ textDecoration: 'none', fontSize: '20px' }}>📚</Link>
                )}

                {/* Collapse/expand button */}
                <button
                    onClick={handleCollapse}
                    title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                    style={{
                        width: '24px',
                        height: '24px',
                        borderRadius: '4px',
                        border: '1px solid var(--color-border)',
                        background: 'none',
                        cursor: 'pointer',
                        color: 'var(--color-text-muted)',
                        fontSize: '10px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0,
                        transition: 'var(--transition)',
                    }}
                    onMouseEnter={e => {
                        ;(e.currentTarget as HTMLButtonElement).style.borderColor = 'var(--color-primary)'
                        ;(e.currentTarget as HTMLButtonElement).style.color = 'var(--color-primary)'
                    }}
                    onMouseLeave={e => {
                        ;(e.currentTarget as HTMLButtonElement).style.borderColor = 'var(--color-border)'
                        ;(e.currentTarget as HTMLButtonElement).style.color = 'var(--color-text-muted)'
                    }}
                >
                    {collapsed ? '▶' : '◀'}
                </button>
            </div>

            {/* ── Nav items ───────────────────────────────────────────── */}
            <nav
                style={{
                    padding: collapsed ? '10px 6px' : '10px 8px',
                    flex: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '2px',
                    overflowY: 'auto',
                }}
            >
                {!collapsed && (
                    <p style={{
                        margin: '0 0 4px 10px',
                        fontSize: '9px',
                        color: 'var(--color-text-muted)',
                        fontFamily: 'var(--font-body)',
                        letterSpacing: '0.1em',
                        textTransform: 'uppercase',
                        opacity: 0.6,
                    }}>
                        Navigation
                    </p>
                )}

                {NAV_ITEMS.filter(item => !item.modOnly || isMod).map(item => {
                    const active = isActive(item.path)
                    return (
                        <Link
                            key={item.path}
                            to={item.path}
                            title={collapsed ? item.label : ''}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: collapsed ? 0 : '10px',
                                justifyContent: collapsed ? 'center' : 'flex-start',
                                padding: collapsed ? '10px 0' : '9px 10px',
                                borderRadius: 'var(--border-radius)',
                                textDecoration: 'none',
                                fontSize: '13px',
                                fontFamily: 'var(--font-body)',
                                fontWeight: active ? 600 : 400,
                                color: active ? 'var(--color-primary)' : 'var(--color-text-muted)',
                                background: active ? 'color-mix(in srgb, var(--color-primary) 10%, transparent)' : 'transparent',
                                borderLeft: !collapsed && active
                                    ? '2px solid var(--color-primary)'
                                    : '2px solid transparent',
                                transition: 'var(--transition)',
                            }}
                            onMouseEnter={e => {
                                if (!active) {
                                    ;(e.currentTarget as HTMLAnchorElement).style.background = 'var(--color-surface-alt)'
                                    ;(e.currentTarget as HTMLAnchorElement).style.color = 'var(--color-text)'
                                }
                            }}
                            onMouseLeave={e => {
                                if (!active) {
                                    ;(e.currentTarget as HTMLAnchorElement).style.background = 'transparent'
                                    ;(e.currentTarget as HTMLAnchorElement).style.color = 'var(--color-text-muted)'
                                }
                            }}
                        >
                            <span style={{ fontSize: '15px', width: collapsed ? 'auto' : '18px', textAlign: 'center', flexShrink: 0 }}>
                                {item.icon}
                            </span>
                            {!collapsed && <span style={{ flex: 1 }}>{item.label}</span>}
                            {!collapsed && item.modOnly && (
                                <span style={{
                                    fontSize: '9px',
                                    color: 'var(--color-primary)',
                                    fontFamily: 'var(--font-body)',
                                    padding: '1px 5px',
                                    border: '1px solid color-mix(in srgb, var(--color-primary) 40%, transparent)',
                                    borderRadius: '3px',
                                }}>
                                    MOD
                                </span>
                            )}
                        </Link>
                    )
                })}
            </nav>

            {/* ── Divider ─────────────────────────────────────────────── */}
            <div style={{ height: '1px', background: 'var(--color-border)', margin: '0 10px', flexShrink: 0 }} />

            {/* ── User section ────────────────────────────────────────── */}
            <div style={{ padding: collapsed ? '8px 6px' : '8px', flexShrink: 0 }}>
                <button
                    onClick={() => {
                        if (!collapsed) {
                            setUserOpen(!userOpen)
                            setThemeOpen(false)
                        }
                    }}
                    title={collapsed ? user?.username ?? '' : ''}
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: collapsed ? 0 : '10px',
                        justifyContent: collapsed ? 'center' : 'flex-start',
                        padding: collapsed ? '8px 0' : '8px 10px',
                        borderRadius: 'var(--border-radius)',
                        border: 'none',
                        cursor: 'pointer',
                        width: '100%',
                        background: userOpen ? 'var(--color-surface-alt)' : 'none',
                        transition: 'var(--transition)',
                    }}
                    onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-surface-alt)')}
                    onMouseLeave={e => (e.currentTarget.style.background = userOpen ? 'var(--color-surface-alt)' : 'none')}
                >
                    <Avatar name={user?.username ?? '?'} avatarUrl={user?.avatar_url} size="sm" />
                    {!collapsed && (
                        <>
                            <div style={{ flex: 1, minWidth: 0, textAlign: 'left' }}>
                                <p style={{
                                    margin: 0, fontSize: '12px', fontWeight: 600,
                                    color: 'var(--color-text)', fontFamily: 'var(--font-body)',
                                    whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis',
                                }}>
                                    {user?.username}
                                </p>
                                <p style={{
                                    margin: 0, fontSize: '10px',
                                    color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                                    textTransform: 'capitalize',
                                }}>
                                    {user?.role}
                                </p>
                            </div>
                            <span style={{ fontSize: '10px', color: 'var(--color-text-muted)', flexShrink: 0 }}>
                                {userOpen ? '▴' : '▾'}
                            </span>
                        </>
                    )}
                </button>

                {/* User dropdown */}
                {userOpen && !collapsed && (
                    <div style={{
                        marginTop: '4px',
                        background: 'var(--color-surface-alt)',
                        borderRadius: 'var(--border-radius)',
                        border: '1px solid var(--color-border)',
                        overflow: 'hidden',
                    }}>
                        {/* Theme sub-row */}
                        <button
                            onClick={() => setThemeOpen(!themeOpen)}
                            style={{
                                display: 'flex', alignItems: 'center', gap: '8px',
                                padding: '9px 12px', width: '100%', border: 'none',
                                background: themeOpen ? 'var(--color-surface-alt)' : 'transparent',
                                color: 'var(--color-text-muted)',
                                fontSize: '12px', fontFamily: 'var(--font-body)',
                                cursor: 'pointer', textAlign: 'left', transition: 'var(--transition)',
                            }}
                            onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-surface-alt)')}
                            onMouseLeave={e => (e.currentTarget.style.background = themeOpen ? 'var(--color-surface-alt)' : 'transparent')}
                        >
                            <span>{currentTheme.emoji}</span>
                            <span style={{ flex: 1 }}>Theme</span>
                            <span style={{ fontSize: '10px', color: 'var(--color-text-muted)' }}>
                                {currentTheme.name} {themeOpen ? '▴' : '▾'}
                            </span>
                        </button>

                        {themeOpen && (
                            <div style={{
                                borderTop: '1px solid var(--color-border)',
                                maxHeight: '200px',
                                overflowY: 'auto',
                            }}>
                                {Object.values(themes).map(theme => (
                                    <button
                                        key={theme.id}
                                        onClick={() => handleThemeChange(theme.id as ThemeId)}
                                        style={{
                                            display: 'flex', alignItems: 'center', gap: '8px',
                                            padding: '7px 12px 7px 22px', width: '100%', border: 'none',
                                            background: themeId === theme.id ? 'color-mix(in srgb, var(--color-primary) 10%, transparent)' : 'transparent',
                                            color: themeId === theme.id ? 'var(--color-primary)' : 'var(--color-text-muted)',
                                            fontSize: '12px', fontFamily: 'var(--font-body)',
                                            cursor: 'pointer', textAlign: 'left', transition: 'var(--transition)',
                                        }}
                                        onMouseEnter={e => {
                                            if (themeId !== theme.id) (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-surface-alt)'
                                        }}
                                        onMouseLeave={e => {
                                            if (themeId !== theme.id) (e.currentTarget as HTMLButtonElement).style.background = 'transparent'
                                        }}
                                    >
                                        <span>{theme.emoji}</span>
                                        <span style={{ flex: 1 }}>{theme.name}</span>
                                        {themeId === theme.id && (
                                            <span style={{ fontSize: '10px', color: 'var(--color-primary)' }}>✓</span>
                                        )}
                                    </button>
                                ))}
                            </div>
                        )}

                        <div style={{ height: '1px', background: 'var(--color-border)' }} />

                        {/* Profile + Settings links */}
                        {[
                            { label: 'Profile',  path: '/profile',  icon: '👤' },
                            { label: 'Settings', path: '/settings', icon: '⚙️' },
                        ].map(item => (
                            <Link
                                key={item.path}
                                to={item.path}
                                onClick={() => setUserOpen(false)}
                                style={{
                                    display: 'flex', alignItems: 'center', gap: '8px',
                                    padding: '9px 12px', textDecoration: 'none',
                                    color: 'var(--color-text-muted)',
                                    fontSize: '12px', fontFamily: 'var(--font-body)',
                                    transition: 'var(--transition)',
                                }}
                                onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-surface-alt)')}
                                onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                            >
                                <span>{item.icon}</span>
                                <span>{item.label}</span>
                            </Link>
                        ))}

                        <div style={{ height: '1px', background: 'var(--color-border)' }} />

                        {/* Logout */}
                        <button
                            onClick={() => { setUserOpen(false); logoutMutation.mutate() }}
                            style={{
                                display: 'flex', alignItems: 'center', gap: '8px',
                                padding: '9px 12px', width: '100%', border: 'none',
                                background: 'transparent', cursor: 'pointer',
                                color: 'var(--color-error)',
                                fontSize: '12px', fontFamily: 'var(--font-body)',
                                textAlign: 'left', transition: 'var(--transition)',
                            }}
                            onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-surface-alt)')}
                            onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                        >
                            <span>🚪</span>
                            <span>{logoutMutation.isPending ? 'Logging out…' : 'Log out'}</span>
                        </button>
                    </div>
                )}
            </div>
        </aside>
    )
}