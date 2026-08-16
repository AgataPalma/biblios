import apiClient from './client'
import type {
    Library,
    LibraryMember,
} from '../types'

export interface CreateLibraryPayload {
    name: string
    description?: string
    visibility: 'private' | 'semi_public' | 'public'
    is_cooperative?: boolean
}

export interface UpdateMemberPayload {
    can_view?: boolean
    can_add?: boolean
    can_remove?: boolean
    can_edit?: boolean
    can_invite?: boolean
    can_manage_members?: boolean
}

export interface UpdateLibraryPayload {
    name?: string
    description?: string
    visibility?: 'private' | 'semi_public' | 'public'
}

export async function getLibraries(): Promise<Library[]> {
    const res = await apiClient.get<Library[]>('/libraries')
    return res.data
}

export async function getLibrary(id: string): Promise<Library> {
    const res = await apiClient.get<Library>(`/libraries/${id}`)
    return res.data
}

export async function getLibraryMembers(id: string): Promise<LibraryMember[]> {
    const res = await apiClient.get<LibraryMember[]>(`/libraries/${id}/members`)
    return res.data
}

export async function createLibrary(data: CreateLibraryPayload): Promise<Library> {
    const res = await apiClient.post<Library>('/libraries', data)
    return res.data
}

export async function updateLibrary(
    id: string,
    data: UpdateLibraryPayload,
): Promise<Library> {
    const res = await apiClient.put<Library>(`/libraries/${id}`, data)
    return res.data
}

export async function inviteMember(libraryId: string, email: string): Promise<void> {
    await apiClient.post(`/libraries/${libraryId}/invite`, { email })
}

export async function acceptInvitation(token: string): Promise<void> {
    await apiClient.post(`/invitations/${token}/accept`)
}

export async function declineInvitation(token: string): Promise<void> {
    await apiClient.post(`/invitations/${token}/decline`)
}

export async function updateMember(
    libraryId: string,
    userId: string,
    permissions: UpdateMemberPayload,
): Promise<void> {
    await apiClient.put(
        `/libraries/${libraryId}/members/${userId}`,
        permissions,
    )
}

export async function removeMember(libraryId: string, userId: string): Promise<void> {
    await apiClient.delete(`/libraries/${libraryId}/members/${userId}`)
}
