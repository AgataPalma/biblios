import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import {
    getPendingSubmissions,
    approveSubmission,
    rejectSubmission,
} from '../api/books'
import { Card, Badge, Spinner, EmptyState, Modal } from '../components'
import type { Submission } from '../types'

export default function ModerationPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()
    const [page, setPage] = useState(1)
    const [rejectTarget, setRejectTarget] = useState<Submission | null>(null)
    const [rejectReason, setRejectReason] = useState('')

    const { data, isLoading } = useQuery({
        queryKey: ['moderation', page],
        queryFn: () => getPendingSubmissions(page, 20),
    })

    const approveMutation = useMutation({
        mutationFn: (id: string) => approveSubmission(id),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['moderation'] }),
    })

    const rejectMutation = useMutation({
        mutationFn: ({ id, reason }: { id: string; reason: string }) =>
            rejectSubmission(id, reason),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['moderation'] })
            setRejectTarget(null)
            setRejectReason('')
        },
    })

    const submissions: Submission[] = data?.submissions ?? []

    return (
        <div style={{
            minHeight: 'calc(100vh - 56px)',
            padding: '32px 24px',
            maxWidth: '900px',
            margin: '0 auto',
        }}>
            {/* Header */}
            <div style={{ marginBottom: '24px' }}>
                <h1 style={{
                    margin: 0,
                    fontSize: '24px',
                    fontWeight: 700,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-heading)',
                }}>
                    🛡️ Moderation Queue
                </h1>
                <p style={{
                    margin: '4px 0 0',
                    fontSize: '13px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                }}>
                    {data?.total ?? 0} pending submission{data?.total !== 1 ? 's' : ''} awaiting review
                </p>
            </div>

            {/* Content */}
            {isLoading ? (
                <div style={{ display: 'flex', justifyContent: 'center', paddingTop: '80px' }}>
                    <Spinner size="lg" label="Loading submissions..." />
                </div>
            ) : submissions.length === 0 ? (
                <Card padding="lg">
                    <EmptyState
                        icon="✅"
                        title="Queue is clear!"
                        description="No pending submissions to review right now."
                    />
                </Card>
            ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    {submissions.map(submission => (
                        <SubmissionCard
                            key={submission.id}
                            submission={submission}
                            onView={() => navigate(`/books/${submission.book_id}`)}
                            onApprove={() => approveMutation.mutate(submission.id)}
                            onReject={() => setRejectTarget(submission)}
                            isApproving={approveMutation.isPending && approveMutation.variables === submission.id}
                        />
                    ))}

                    {/* Pagination */}
                    {data?.total > 20 && (
                        <div style={{ display: 'flex', justifyContent: 'center', gap: '8px', marginTop: '16px' }}>
                            <button
                                disabled={page === 1}
                                onClick={() => setPage(p => p - 1)}
                                style={{
                                    padding: '8px 16px',
                                    background: 'var(--color-surface)',
                                    border: '1px solid var(--color-border)',
                                    borderRadius: 'var(--border-radius)',
                                    cursor: page === 1 ? 'not-allowed' : 'pointer',
                                    opacity: page === 1 ? 0.4 : 1,
                                    color: 'var(--color-text)',
                                    fontFamily: 'var(--font-body)',
                                    fontSize: '13px',
                                }}
                            >
                                ← Prev
                            </button>
                            <span style={{
                                padding: '8px 16px',
                                color: 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                                fontSize: '13px',
                            }}>
                Page {page}
              </span>
                            <button
                                disabled={submissions.length < 20}
                                onClick={() => setPage(p => p + 1)}
                                style={{
                                    padding: '8px 16px',
                                    background: 'var(--color-surface)',
                                    border: '1px solid var(--color-border)',
                                    borderRadius: 'var(--border-radius)',
                                    cursor: submissions.length < 20 ? 'not-allowed' : 'pointer',
                                    opacity: submissions.length < 20 ? 0.4 : 1,
                                    color: 'var(--color-text)',
                                    fontFamily: 'var(--font-body)',
                                    fontSize: '13px',
                                }}
                            >
                                Next →
                            </button>
                        </div>
                    )}
                </div>
            )}

            {/* Reject modal */}
            <Modal
                isOpen={!!rejectTarget}
                onClose={() => { setRejectTarget(null); setRejectReason('') }}
                title="Reject submission"
                confirmLabel="Reject"
                confirmVariant="danger"
                onConfirm={() => rejectTarget && rejectMutation.mutate({
                    id: rejectTarget.id,
                    reason: rejectReason,
                })}
                isLoading={rejectMutation.isPending}
                size="sm"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    <p style={{
                        margin: 0,
                        fontSize: '13px',
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-body)',
                    }}>
                        Provide a reason for rejecting this submission:
                    </p>
                    <textarea
                        value={rejectReason}
                        onChange={e => setRejectReason(e.target.value)}
                        placeholder="e.g. Duplicate entry, incorrect information..."
                        rows={3}
                        style={{
                            padding: '10px 12px',
                            background: 'var(--input-bg)',
                            border: '1px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)',
                            color: 'var(--color-text)',
                            fontSize: '13px',
                            fontFamily: 'var(--font-body)',
                            resize: 'vertical',
                            outline: 'none',
                            width: '100%',
                            boxSizing: 'border-box',
                        }}
                    />
                </div>
            </Modal>
        </div>
    )
}

// ─── Submission Card ──────────────────────────────────────────────────────────

function SubmissionCard({
                            submission,
                            onView,
                            onApprove,
                            onReject,
                            isApproving,
                        }: {
    submission: Submission
    onView: () => void
    onApprove: () => void
    onReject: () => void
    isApproving: boolean
}) {
    const date = new Date(submission.created_at).toLocaleDateString('en-GB', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
    })

    return (
        <Card padding="md">
            <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '16px',
                flexWrap: 'wrap',
            }}>
                {/* Info */}
                <div style={{ flex: 1, minWidth: '200px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                        <Badge label="pending" variant="warning" size="sm" />
                        <span style={{
                            fontSize: '11px',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                        }}>
              Submitted {date}
            </span>
                    </div>
                    <p style={{
                        margin: 0,
                        fontSize: '12px',
                        color: 'var(--color-text-muted)',
                        fontFamily: 'var(--font-body)',
                        fontStyle: 'italic',
                    }}>
                        Book ID: {submission.book_id}
                    </p>
                </div>

                {/* Actions */}
                <div style={{ display: 'flex', gap: '8px' }}>
                    <button
                        onClick={onView}
                        style={{
                            padding: '8px 14px',
                            background: 'var(--color-surface-alt)',
                            border: '1px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)',
                            cursor: 'pointer',
                            fontSize: '12px',
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-body)',
                            transition: 'var(--transition)',
                        }}
                    >
                        👁 View
                    </button>
                    <button
                        onClick={onApprove}
                        disabled={isApproving}
                        style={{
                            padding: '8px 14px',
                            background: 'rgba(34,197,94,0.15)',
                            border: '1px solid rgba(34,197,94,0.3)',
                            borderRadius: 'var(--border-radius)',
                            cursor: isApproving ? 'not-allowed' : 'pointer',
                            fontSize: '12px',
                            color: 'var(--color-success)',
                            fontFamily: 'var(--font-body)',
                            fontWeight: 600,
                            transition: 'var(--transition)',
                        }}
                    >
                        ✅ Approve
                    </button>
                    <button
                        onClick={onReject}
                        style={{
                            padding: '8px 14px',
                            background: 'rgba(239,68,68,0.1)',
                            border: '1px solid rgba(239,68,68,0.2)',
                            borderRadius: 'var(--border-radius)',
                            cursor: 'pointer',
                            fontSize: '12px',
                            color: 'var(--color-error)',
                            fontFamily: 'var(--font-body)',
                            fontWeight: 600,
                            transition: 'var(--transition)',
                        }}
                    >
                        ❌ Reject
                    </button>
                </div>
            </div>
        </Card>
    )
}