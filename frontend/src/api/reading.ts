import apiClient from './client'
import type {
    ReadingChallenge,
    ReadingSession,
    ChallengeProgress,
    SessionsResponse,
} from '../types'

export interface CreateChallengePayload {
    year: number
    goal: number
}

export interface CreateSessionPayload {
    copy_id: string
    logged_date: string
    pages_read?: number
    note?: string
}

export async function getChallenges(): Promise<ReadingChallenge[]> {
    const res = await apiClient.get<ReadingChallenge[]>('/reading/challenges')
    return res.data
}

export async function createChallenge(data: CreateChallengePayload): Promise<ReadingChallenge> {
    const res = await apiClient.post<ReadingChallenge>('/reading/challenges', data)
    return res.data
}

export async function getChallengeProgress(id: string): Promise<ChallengeProgress> {
    const res = await apiClient.get<ChallengeProgress>(`/reading/challenges/${id}/progress`)
    return res.data
}

export async function getSessions(page = 1, limit = 20): Promise<SessionsResponse> {
    const res = await apiClient.get<SessionsResponse>('/reading/sessions', {
        params: { page, limit },
    })
    return res.data
}

export async function createSession(data: CreateSessionPayload): Promise<ReadingSession> {
    const res = await apiClient.post<ReadingSession>('/reading/sessions', data)
    return res.data
}
