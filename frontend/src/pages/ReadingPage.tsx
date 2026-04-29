import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getChallenges, createChallenge, getSessions, createSession } from '../api/reading'
import type { ReadingChallenge, ReadingSession } from '../types'
import { Button, Input, Modal, Card, Badge, Spinner, EmptyState } from '../components'

// ── Inline ProgressBar ────────────────────────────────────────────────────────
function ProgressBar({ value, max }: { value: number; max: number }) {
    const pct = max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0
    return (
        <div style={{ height: '6px', background: 'var(--color-surface-alt)', borderRadius: '3px', overflow: 'hidden' }}>
            <div style={{ height: '100%', width: `${pct}%`, background: 'var(--color-primary)', borderRadius: '3px', transition: 'width 0.3s ease' }} />
        </div>
    )
}

// ── Tab type ──────────────────────────────────────────────────────────────────
type Tab = 'challenges' | 'sessions'

// ── Challenges tab ────────────────────────────────────────────────────────────
function ChallengesTab() {
    const queryClient = useQueryClient()
    const [showModal, setShowModal] = useState(false)

    // Form state
    const [title, setTitle] = useState('')
    const [startDate, setStartDate] = useState('')
    const [endDate, setEndDate] = useState('')
    const [goalBooks, setGoalBooks] = useState('')

    const { data, isLoading, isError } = useQuery({
        queryKey: ['reading-challenges'],
        queryFn: getChallenges,
    })

    const createMutation = useMutation({
        mutationFn: () =>
            createChallenge({
                title,
                start_date: startDate,
                end_date: endDate,
                goal_books: Number(goalBooks),
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['reading-challenges'] })
            closeModal()
        },
    })

    function closeModal() {
        setShowModal(false)
        setTitle('')
        setStartDate('')
        setEndDate('')
        setGoalBooks('')
    }

    function statusBadge(challenge: ReadingChallenge) {
        if (challenge.status === 'completed') {
            return <Badge label="✅ completed" variant="success" />
        }
        if (challenge.status === 'failed') {
            return <Badge label="failed" variant="error" />
        }
        return <Badge label="active" variant="info" />
    }

    const challenges = data?.challenges ?? []

    return (
        <>
            {/* Tab header */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '20px' }}>
                <Button label="New Challenge" variant="primary" onClick={() => setShowModal(true)} />
            </div>

            {/* Loading */}
            {isLoading && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '48px 0' }}>
                    <Spinner />
                </div>
            )}

            {/* Error */}
            {isError && <EmptyState icon="⚠️" title="Failed to load challenges" />}

            {/* Challenge cards */}
            {!isLoading && !isError && challenges.length > 0 && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    {challenges.map((challenge) => {
                        const pct =
                            challenge.goal_books > 0
                                ? Math.min(100, Math.round((challenge.current_books / challenge.goal_books) * 100))
                                : 0
                        return (
                            <Card key={challenge.id}>
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                                    {/* Title + badge row */}
                                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '12px' }}>
                                        <span
                                            style={{
                                                fontSize: '16px',
                                                fontWeight: 700,
                                                color: 'var(--color-text)',
                                                fontFamily: 'var(--font-heading)',
                                            }}
                                        >
                                            {challenge.title}
                                        </span>
                                        {statusBadge(challenge)}
                                    </div>

                                    {/* Date range */}
                                    <span style={{ fontSize: '13px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                        {new Date(challenge.start_date).toLocaleDateString()} –{' '}
                                        {new Date(challenge.end_date).toLocaleDateString()}
                                    </span>

                                    {/* Progress bar */}
                                    <ProgressBar value={challenge.current_books} max={challenge.goal_books} />

                                    {/* Progress text */}
                                    <span style={{ fontSize: '13px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                        {challenge.current_books} / {challenge.goal_books} books ({pct}%)
                                    </span>
                                </div>
                            </Card>
                        )
                    })}
                </div>
            )}

            {/* Empty state */}
            {!isLoading && !isError && challenges.length === 0 && (
                <EmptyState
                    icon="🎯"
                    title="No challenges yet"
                    description="Set a reading goal to track your progress"
                    action={{ label: 'New Challenge', onClick: () => setShowModal(true) }}
                />
            )}

            {/* New Challenge modal */}
            <Modal
                isOpen={showModal}
                onClose={closeModal}
                title="New Challenge"
                confirmLabel="Create"
                onConfirm={() => createMutation.mutate()}
                confirmVariant="primary"
                isLoading={createMutation.isPending}
                size="md"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Input
                        label="Title"
                        value={title}
                        onChange={setTitle}
                        placeholder="e.g. 2026 Reading Challenge"
                    />
                    <Input
                        label="Start date"
                        type="date"
                        value={startDate}
                        onChange={setStartDate}
                    />
                    <Input
                        label="End date"
                        type="date"
                        value={endDate}
                        onChange={setEndDate}
                    />
                    <Input
                        label="Goal (books)"
                        type="number"
                        value={goalBooks}
                        onChange={setGoalBooks}
                        placeholder="e.g. 20"
                    />
                </div>
            </Modal>
        </>
    )
}

// ── Sessions tab ──────────────────────────────────────────────────────────────
function SessionsTab() {
    const queryClient = useQueryClient()
    const [showModal, setShowModal] = useState(false)

    // Form state
    const [copyId, setCopyId] = useState('')
    const [date, setDate] = useState(new Date().toISOString().slice(0, 10))
    const [pagesRead, setPagesRead] = useState('')
    const [notes, setNotes] = useState('')

    const { data, isLoading, isError } = useQuery({
        queryKey: ['reading-sessions'],
        queryFn: () => getSessions(1, 50),
    })

    const createMutation = useMutation({
        mutationFn: () =>
            createSession({
                copy_id: copyId,
                date,
                pages_read: Number(pagesRead),
                notes: notes || undefined,
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['reading-sessions'] })
            closeModal()
        },
    })

    function closeModal() {
        setShowModal(false)
        setCopyId('')
        setDate(new Date().toISOString().slice(0, 10))
        setPagesRead('')
        setNotes('')
    }

    const sessions = data?.sessions ?? []

    return (
        <>
            {/* Tab header */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '20px' }}>
                <Button label="Log Session" variant="primary" onClick={() => setShowModal(true)} />
            </div>

            {/* Loading */}
            {isLoading && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '48px 0' }}>
                    <Spinner />
                </div>
            )}

            {/* Error */}
            {isError && <EmptyState icon="⚠️" title="Failed to load sessions" />}

            {/* Session list */}
            {!isLoading && !isError && sessions.length > 0 && (
                <div
                    style={{
                        background: 'var(--color-surface)',
                        border: '1px solid var(--color-border)',
                        borderRadius: 'var(--border-radius)',
                        overflow: 'hidden',
                    }}
                >
                    {sessions.map((session: ReadingSession, idx: number) => (
                        <div
                            key={session.id}
                            style={{
                                padding: '16px 20px',
                                borderBottom: idx < sessions.length - 1 ? '1px solid var(--color-border)' : 'none',
                                display: 'flex',
                                flexDirection: 'column',
                                gap: '6px',
                            }}
                        >
                            {/* Book title + date row */}
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '12px' }}>
                                <span
                                    style={{
                                        fontSize: '14px',
                                        fontWeight: 700,
                                        color: 'var(--color-text)',
                                        fontFamily: 'var(--font-heading)',
                                    }}
                                >
                                    {session.book_title ?? 'Unknown book'}
                                </span>
                                <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', whiteSpace: 'nowrap', fontFamily: 'var(--font-body)' }}>
                                    {new Date(session.date).toLocaleDateString()}
                                </span>
                            </div>

                            {/* Pages read */}
                            <span style={{ fontSize: '13px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                📄 {session.pages_read} pages
                            </span>

                            {/* Notes (truncated) */}
                            {session.notes && (
                                <span
                                    style={{
                                        fontSize: '12px',
                                        color: 'var(--color-text-muted)',
                                        fontFamily: 'var(--font-body)',
                                        overflow: 'hidden',
                                        textOverflow: 'ellipsis',
                                        whiteSpace: 'nowrap',
                                        maxWidth: '600px',
                                    }}
                                >
                                    {session.notes}
                                </span>
                            )}
                        </div>
                    ))}
                </div>
            )}

            {/* Empty state */}
            {!isLoading && !isError && sessions.length === 0 && (
                <EmptyState
                    icon="📖"
                    title="No sessions logged"
                    description="Track your reading sessions"
                    action={{ label: 'Log Session', onClick: () => setShowModal(true) }}
                />
            )}

            {/* Log Session modal */}
            <Modal
                isOpen={showModal}
                onClose={closeModal}
                title="Log Session"
                confirmLabel="Log"
                onConfirm={() => createMutation.mutate()}
                confirmVariant="primary"
                isLoading={createMutation.isPending}
                size="md"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Input
                        label="Copy ID"
                        value={copyId}
                        onChange={setCopyId}
                        placeholder="Enter your copy ID"
                    />
                    <Input
                        label="Date"
                        type="date"
                        value={date}
                        onChange={setDate}
                    />
                    <Input
                        label="Pages read"
                        type="number"
                        value={pagesRead}
                        onChange={setPagesRead}
                        placeholder="e.g. 30"
                    />
                    {/* Notes textarea */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                        <label
                            style={{
                                fontSize: '14px',
                                fontWeight: 500,
                                color: 'var(--color-text)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            Notes
                        </label>
                        <textarea
                            value={notes}
                            onChange={(e) => setNotes(e.target.value)}
                            placeholder="Optional notes about this session…"
                            rows={3}
                            style={{
                                background: 'var(--input-bg)',
                                border: '1px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)',
                                color: 'var(--color-text)',
                                fontFamily: 'var(--font-body)',
                                fontSize: '14px',
                                padding: '8px 12px',
                                resize: 'vertical',
                                outline: 'none',
                                width: '100%',
                                boxSizing: 'border-box',
                            }}
                        />
                    </div>
                </div>
            </Modal>
        </>
    )
}

// ── ReadingPage ───────────────────────────────────────────────────────────────
export default function ReadingPage() {
    const [activeTab, setActiveTab] = useState<Tab>('challenges')

    const tabs: { key: Tab; label: string }[] = [
        { key: 'challenges', label: 'Challenges' },
        { key: 'sessions', label: 'Sessions' },
    ]

    return (
        <div
            style={{
                maxWidth: '900px',
                margin: '0 auto',
                padding: '32px',
                fontFamily: 'var(--font-body)',
                color: 'var(--color-text)',
            }}
        >
            {/* Page header */}
            <h1
                style={{
                    margin: '0 0 24px 0',
                    fontSize: '24px',
                    fontWeight: 700,
                    fontFamily: 'var(--font-heading)',
                    color: 'var(--color-text)',
                }}
            >
                Reading
            </h1>

            {/* Tab navigation */}
            <div
                style={{
                    display: 'flex',
                    gap: '0',
                    borderBottom: '2px solid var(--color-border)',
                    marginBottom: '24px',
                }}
            >
                {tabs.map((tab) => (
                    <button
                        key={tab.key}
                        onClick={() => setActiveTab(tab.key)}
                        style={{
                            background: 'none',
                            border: 'none',
                            borderBottom: activeTab === tab.key
                                ? '2px solid var(--color-primary)'
                                : '2px solid transparent',
                            marginBottom: '-2px',
                            padding: '10px 20px',
                            fontSize: '14px',
                            fontWeight: activeTab === tab.key ? 600 : 400,
                            color: activeTab === tab.key ? 'var(--color-primary)' : 'var(--color-text-muted)',
                            cursor: 'pointer',
                            transition: 'var(--transition)',
                            fontFamily: 'var(--font-body)',
                        }}
                    >
                        {tab.label}
                    </button>
                ))}
            </div>

            {/* Tab content */}
            {activeTab === 'challenges' && <ChallengesTab />}
            {activeTab === 'sessions' && <SessionsTab />}
        </div>
    )
}
