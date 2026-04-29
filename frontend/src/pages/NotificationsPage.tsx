import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getNotifications, markNotificationRead, markAllNotificationsRead } from '../api/notifications'
import { Spinner, EmptyState, Badge, Button } from '../components'
import type { Notification } from '../types'

// ── Relative timestamp ────────────────────────────────────────────────────────

function relativeDate(iso: string): string {
    const diff = Date.now() - new Date(iso).getTime()
    const minutes = Math.floor(diff / 60_000)
    if (minutes < 1) return 'just now'
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ago`
    const days = Math.floor(hours / 24)
    if (days < 7) return `${days}d ago`
    return new Date(iso).toLocaleDateString()
}

// ── Badge mapping ─────────────────────────────────────────────────────────────

type BadgeVariant = 'info' | 'success' | 'warning' | 'default'

function typeBadge(type: string): { variant: BadgeVariant; label: string } {
    switch (type) {
        case 'invitation':
            return { variant: 'info', label: 'Invitation' }
        case 'review_like':
            return { variant: 'success', label: 'Review Like' }
        case 'library_activity':
            return { variant: 'warning', label: 'Library' }
        default:
            return { variant: 'default', label: type }
    }
}

// ── Filter tabs ───────────────────────────────────────────────────────────────

type FilterTab = 'all' | 'invitation' | 'review_like' | 'library_activity'

const TABS: { key: FilterTab; label: string }[] = [
    { key: 'all', label: 'All' },
    { key: 'invitation', label: 'Invitations' },
    { key: 'review_like', label: 'Review Likes' },
    { key: 'library_activity', label: 'Library Activity' },
]

// ── Notification item ─────────────────────────────────────────────────────────

interface NotificationItemProps {
    notification: Notification
    onMarkRead: (id: string) => void
    isPending: boolean
}

function NotificationItem({ notification, onMarkRead, isPending }: NotificationItemProps) {
    const { variant, label } = typeBadge(notification.type)

    return (
        <div
            onClick={() => !notification.is_read && !isPending && onMarkRead(notification.id)}
            style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: '12px',
                padding: '14px 16px',
                background: notification.is_read
                    ? 'var(--color-surface)'
                    : 'rgba(var(--color-primary-rgb, 0, 0, 0), 0.08)',
                borderBottom: '1px solid var(--color-border)',
                cursor: notification.is_read ? 'default' : 'pointer',
                transition: 'var(--transition)',
            }}
        >
            {/* Type badge */}
            <div style={{ flexShrink: 0, paddingTop: '2px' }}>
                <Badge variant={variant} label={label} size="sm" />
            </div>

            {/* Title + body */}
            <div style={{ flex: 1, minWidth: 0 }}>
                <p
                    style={{
                        margin: 0,
                        fontSize: '14px',
                        fontWeight: notification.is_read ? 400 : 600,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-body)',
                        lineHeight: 1.4,
                    }}
                >
                    {notification.title}
                </p>
                {notification.body && (
                    <p
                        style={{
                            margin: '2px 0 0',
                            fontSize: '13px',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                            lineHeight: 1.5,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        {notification.body}
                    </p>
                )}
            </div>

            {/* Timestamp */}
            <span
                style={{
                    flexShrink: 0,
                    fontSize: '12px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                    paddingTop: '2px',
                    whiteSpace: 'nowrap',
                }}
            >
                {relativeDate(notification.created_at)}
            </span>
        </div>
    )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function NotificationsPage() {
    const [activeTab, setActiveTab] = useState<FilterTab>('all')
    const queryClient = useQueryClient()

    const { data, isLoading, isError } = useQuery({
        queryKey: ['notifications'],
        queryFn: () => getNotifications(1, 50),
    })

    const markReadMutation = useMutation({
        mutationFn: (id: string) => markNotificationRead(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['notifications'] })
        },
    })

    const markAllReadMutation = useMutation({
        mutationFn: (ids: string[]) => markAllNotificationsRead(ids),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['notifications'] })
        },
    })

    const notifications = data?.notifications ?? []
    const unreadCount = data?.unread_count ?? 0

    const filtered =
        activeTab === 'all'
            ? notifications
            : notifications.filter(n => n.type === activeTab)

    const unreadIds = notifications.filter(n => !n.is_read).map(n => n.id)

    function handleMarkAllRead() {
        if (unreadIds.length === 0) return
        markAllReadMutation.mutate(unreadIds)
    }

    return (
        <div
            style={{
                maxWidth: '700px',
                margin: '0 auto',
                padding: '32px',
            }}
        >
            {/* Page header */}
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: '24px',
                    gap: '16px',
                }}
            >
                <h1
                    style={{
                        margin: 0,
                        fontSize: '24px',
                        fontWeight: 700,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                    }}
                >
                    Notifications
                </h1>
                <Button
                    label="Mark all as read"
                    variant="secondary"
                    onClick={handleMarkAllRead}
                    disabled={unreadCount === 0 || markAllReadMutation.isPending}
                    isLoading={markAllReadMutation.isPending}
                />
            </div>

            {/* Filter tabs */}
            <div
                style={{
                    display: 'flex',
                    gap: '0',
                    borderBottom: '1px solid var(--color-border)',
                    marginBottom: '0',
                }}
            >
                {TABS.map(tab => (
                    <button
                        key={tab.key}
                        onClick={() => setActiveTab(tab.key)}
                        style={{
                            background: 'none',
                            border: 'none',
                            borderBottom: activeTab === tab.key
                                ? '2px solid var(--color-primary)'
                                : '2px solid transparent',
                            padding: '10px 16px',
                            fontSize: '14px',
                            fontWeight: activeTab === tab.key ? 600 : 400,
                            color: activeTab === tab.key
                                ? 'var(--color-primary)'
                                : 'var(--color-text-muted)',
                            cursor: 'pointer',
                            transition: 'var(--transition)',
                            fontFamily: 'var(--font-body)',
                            marginBottom: '-1px',
                        }}
                    >
                        {tab.label}
                    </button>
                ))}
            </div>

            {/* Content */}
            <div
                style={{
                    background: 'var(--color-surface)',
                    borderRadius: 'var(--border-radius)',
                    boxShadow: 'var(--shadow-sm)',
                    overflow: 'hidden',
                    border: '1px solid var(--color-border)',
                    borderTop: 'none',
                    borderTopLeftRadius: 0,
                    borderTopRightRadius: 0,
                }}
            >
                {isLoading ? (
                    <div style={{ padding: '48px', display: 'flex', justifyContent: 'center' }}>
                        <Spinner />
                    </div>
                ) : isError ? (
                    <EmptyState
                        icon="⚠️"
                        title="Failed to load notifications"
                    />
                ) : filtered.length === 0 ? (
                    <EmptyState
                        icon="🔔"
                        title="No notifications"
                        description="You're all caught up!"
                    />
                ) : (
                    filtered.map(notification => (
                        <NotificationItem
                            key={notification.id}
                            notification={notification}
                            onMarkRead={id => markReadMutation.mutate(id)}
                            isPending={markReadMutation.isPending}
                        />
                    ))
                )}
            </div>
        </div>
    )
}
