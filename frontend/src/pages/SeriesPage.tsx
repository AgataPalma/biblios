import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getSeries } from '../api/series'
import type { Series } from '../types'
import { Card, Spinner, EmptyState } from '../components'

export default function SeriesPage() {
    const navigate = useNavigate()
    const [search, setSearch] = useState('')

    const { data, isLoading, isError } = useQuery({
        queryKey: ['series'],
        queryFn: () => getSeries(1, 100),
    })

    if (isLoading) {
        return (
            <div style={{ display: 'flex', justifyContent: 'center', padding: '64px 0' }}>
                <Spinner size="lg" />
            </div>
        )
    }

    if (isError) {
        return (
            <div style={{ maxWidth: 900, margin: '0 auto', padding: '32px' }}>
                <EmptyState icon="⚠️" title="Failed to load series" />
            </div>
        )
    }

    const allSeries: Series[] = data?.series ?? []
    const filtered = allSeries.filter(s =>
        s.name.toLowerCase().includes(search.toLowerCase())
    )

    return (
        <div style={{ maxWidth: 900, margin: '0 auto', padding: '32px' }}>
            {/* Page header */}
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: '24px',
                    gap: '16px',
                    flexWrap: 'wrap',
                }}
            >
                <h1
                    style={{
                        margin: 0,
                        fontSize: '28px',
                        fontWeight: 700,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                    }}
                >
                    Series
                </h1>
                <input
                    type="text"
                    placeholder="Search series…"
                    value={search}
                    onChange={e => setSearch(e.target.value)}
                    style={{
                        padding: '8px 12px',
                        background: 'var(--input-bg)',
                        border: '1px solid var(--color-border)',
                        borderRadius: 'var(--border-radius)',
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-body)',
                        fontSize: '14px',
                        outline: 'none',
                        width: '220px',
                    }}
                />
            </div>

            {/* Series grid */}
            {filtered.length === 0 ? (
                <EmptyState
                    icon="📚"
                    title="No series found"
                    description={
                        search
                            ? 'Try a different search term'
                            : 'No series have been added yet'
                    }
                />
            ) : (
                <div
                    style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
                        gap: '16px',
                    }}
                >
                    {filtered.map(series => (
                        <Card
                            key={series.id}
                            hover
                            onClick={() => navigate(`/series/${series.id}`)}
                        >
                            <p
                                style={{
                                    margin: '0 0 8px',
                                    fontSize: '16px',
                                    fontWeight: 700,
                                    color: 'var(--color-text)',
                                    fontFamily: 'var(--font-heading)',
                                }}
                            >
                                {series.name}
                            </p>
                            {series.description && (
                                <p
                                    style={{
                                        margin: '0 0 12px',
                                        fontSize: '13px',
                                        color: 'var(--color-text-muted)',
                                        fontFamily: 'var(--font-body)',
                                        lineHeight: 1.5,
                                        display: '-webkit-box',
                                        WebkitLineClamp: 2,
                                        WebkitBoxOrient: 'vertical',
                                        overflow: 'hidden',
                                    }}
                                >
                                    {series.description}
                                </p>
                            )}
                            <p
                                style={{
                                    margin: 0,
                                    fontSize: '12px',
                                    color: 'var(--color-text-muted)',
                                    fontFamily: 'var(--font-body)',
                                }}
                            >
                                📚 {series.book_count ?? 0} books
                            </p>
                        </Card>
                    ))}
                </div>
            )}
        </div>
    )
}
