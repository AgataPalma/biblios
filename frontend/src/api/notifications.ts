import apiClient from './client'
import type { Notification, NotificationsResponse } from '../types'

interface NotificationFilters {
    type?: string
    read?: boolean
}

export async function getNotifications(
    page = 1,
    limit = 20,
    filters: NotificationFilters = {},
): Promise<NotificationsResponse> {
    const res = await apiClient.get<NotificationsResponse>('/notifications', {
        params: { page, limit, type: filters.type, read: filters.read },
    })
    return res.data
}

export async function markNotificationRead(id: string): Promise<void> {
    await apiClient.put(`/notifications/${id}/read`)
}

export async function markAllNotificationsRead(): Promise<void> {
    await apiClient.put('/notifications/read-all')
}

// Re-export Notification type for convenience
export type { Notification }
