import apiClient from './client'
import type { Notification, NotificationsResponse } from '../types'

export async function getNotifications(page = 1, limit = 20): Promise<NotificationsResponse> {
    const res = await apiClient.get<NotificationsResponse>('/notifications', {
        params: { page, limit },
    })
    return res.data
}

export async function markNotificationRead(id: string): Promise<void> {
    await apiClient.put(`/notifications/${id}/read`)
}

export async function markAllNotificationsRead(ids: string[]): Promise<void> {
    await Promise.all(ids.map(id => markNotificationRead(id)))
}

// Re-export Notification type for convenience
export type { Notification }
