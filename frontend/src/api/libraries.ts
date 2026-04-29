import apiClient from './client'
import type {
    CooperativeLibrary,
    LibraryMember,
    LibrariesResponse,
    LibraryDetailResponse,
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

export async function getLibraries(): Promise<LibrariesResponse> {
    const res = await apiClient.get<LibrariesResponse>('/libraries')
    return res.data
}

export async function getLibrary(id: string): Promise<LibraryDetailResponse> {
    const res = await apiClient.get<LibraryDetailResponse>(`/libraries/${id}`)
    return res.data
}

export async function createLibrary(data: CreateLibraryPayload): Promise<CooperativeLibrary> {
    const res = await apiClient.post<CooperativeLibrary>('/libraries', data)
    return res.data
}

export async function updateLibrary(
    id: string,
    data: Partial<CreateLibraryPayload>,
): Promise<CooperativeLibrary> {
    const res = await apiClient.put<CooperativeLibrary>(`/libraries/${id}`, data)
    return res.data
}

export async function inviteMember(libraryId: string, email: string): Promise<void> {
    await apiClient.post(`/libraries/${libraryId}/invite`, { email })
}

export async function acceptInvitation(token: string): Promise<void> {
    await apiClient.put(`/libraries/invitations/${token}/accept`)
}

export async function declineInvitation(token: string): Promise<void> {
    await apiClient.put(`/libraries/invitations/${token}/decline`)
}

export async function updateMember(
    libraryId: string,
    userId: string,
    permissions: UpdateMemberPayload,
): Promise<LibraryMember> {
    const res = await apiClient.put<LibraryMember>(
        `/libraries/${libraryId}/members/${userId}`,
        permissions,
    )
    return res.data
}

export async function removeMember(libraryId: string, userId: string): Promise<void> {
    await apiClient.delete(`/libraries/${libraryId}/members/${userId}`)
}
