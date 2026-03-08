import apiClient from './client'
import type { User } from '../types'

export interface UpdateProfilePayload {
    username?: string
    bio?: string
    avatar_url?: string
}

export interface UpdateEmailPayload {
    email: string
    current_password: string
}

export interface UpdatePasswordPayload {
    current_password: string
    new_password: string
}

export async function updateProfile(data: UpdateProfilePayload): Promise<User> {
    const response = await apiClient.put<User>('/users/me', data)
    return response.data
}

export async function updateEmail(data: UpdateEmailPayload): Promise<User> {
    const response = await apiClient.put<User>('/users/me', data)
    return response.data
}

export async function updatePassword(data: UpdatePasswordPayload): Promise<void> {
    await apiClient.put('/users/me/password', data)
}