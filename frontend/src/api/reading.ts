import apiClient from './client'
import type {
    ReadingChallenge,
    ReadingSession,
    ChallengesResponse,
    SessionsResponse,
} from '../types'

export interface CreateChallengePayload {
    title: string
    start_date: string
    end_date: string
    goal_books: number
}

export interface CreateSessionPayload {
    copy_id: string
    date: string
    pages_read: number
    notes?: string
}

export async function getChallenges(): Promise<ChallengesResponse> {
    const res = await apiClient.get<ChallengesResponse>('/reading/challenges')
    return res.data
}

export async function createChallenge(data: CreateChallengePayload): Promise<ReadingChallenge> {
    const res = await apiClient.post<ReadingChallenge>('/reading/challenges', data)
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
