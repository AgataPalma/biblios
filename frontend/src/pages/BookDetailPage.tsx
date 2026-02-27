import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getBook } from '../api/books'
import { Badge, Card, Spinner } from '../components'

export default function BookDetailPage() {
    const { id } = useParams<{ id: string }>()
    const navigate = useNavigate()

    const { data: book, isLoading, isError } = useQuery({
        queryKey: ['book', id],
        queryFn: () => getBook(id!),
        enabled: !!id,
    })

    if (isLoading) {
        return (
            <div style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                minHeight: 'calc(100vh - 56px)',
            }}>
                <Spinner size="lg" label="Loading book..." />
            </div>
        )
    }

    if (isError || !book) {
        return (
            <div style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                minHeight: 'calc(100vh - 56px)',
                gap: '16px',
            }}>
                <span style={{ fontSize: '48px' }}>📭</span>
                <h2 style={{
                    margin: 0,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-heading)',
                }}>
                    Book not found
                </h2>
                <button
                    onClick={() => navigate('/books')}
                    style={{
                        background: 'var(--color-primary)',
                        color: 'var(--color-primary-text)',
                        border: 'none',
                        borderRadius: 'var(--border-radius)',
                        padding: '10px 20px',
                        cursor: 'pointer',
                        fontFamily: 'var(--font-body)',
                        fontSize: '14px',
                    }}
                >
                    Back to catalogue
                </button>
            </div>
        )
    }

    // Deterministic cover colour fallback
    const colors = [
        '#2563eb', '#16a34a', '#dc2626', '#9333ea',
        '#ea580c', '#0891b2', '#d97706', '#4f46e5',
        '#0d9488', '#be185d',
    ]
    const colorIndex = book.title
        .split('')
        .reduce((a, c) => a + c.charCodeAt(0), 0) % colors.length
    const coverColor = colors[colorIndex]

    // Find primary edition
    const primaryEdition = book.editions?.[0]

    return (
        <div style={{
            minHeight: 'calc(100vh - 56px)',
            padding: '32px 24px',
            maxWidth: '900px',
            margin: '0 auto',
        }}>
            {/* Back button */}
            <button
                onClick={() => navigate('/books')}
                style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--color-text-muted)',
                    fontSize: '13px',
                    fontFamily: 'var(--font-body)',
                    padding: 0,
                    marginBottom: '24px',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                }}
            >
                ← Back to catalogue
            </button>

            <div style={{
                display: 'grid',
                gridTemplateColumns: '220px 1fr',
                gap: '32px',
                alignItems: 'start',
            }}>
                {/* Left column — cover + edition info */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    {/* Cover */}
                    <div style={{
                        width: '220px',
                        height: '300px',
                        borderRadius: '6px',
                        overflow: 'hidden',
                        boxShadow: 'var(--shadow-lg)',
                        background: book.cover_url
                            ? `url(${book.cover_url}) center/cover no-repeat`
                            : `linear-gradient(135deg, ${coverColor}dd, ${coverColor}88)`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0,
                        border: '1px solid var(--color-border)',
                    }}>
                        {!book.cover_url && (
                            <span style={{ fontSize: '64px', opacity: 0.5 }}>📖</span>
                        )}
                    </div>

                    {/* Edition details */}
                    {primaryEdition && (
                        <Card padding="sm">
                            <div style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: '8px',
                            }}>
                                <p style={{
                                    margin: 0,
                                    fontSize: '11px',
                                    fontWeight: 700,
                                    color: 'var(--color-text-muted)',
                                    textTransform: 'uppercase',
                                    letterSpacing: '0.08em',
                                    fontFamily: 'var(--font-body)',
                                }}>
                                    Edition Details
                                </p>
                                {[
                                    { label: 'Format', value: primaryEdition.format },
                                    { label: 'Language', value: primaryEdition.language },
                                    { label: 'Publisher', value: primaryEdition.publisher },
                                    { label: 'Published', value: primaryEdition.published_at },
                                    { label: 'Pages', value: primaryEdition.page_count },
                                    { label: 'ISBN', value: primaryEdition.isbn },
                                ].filter(f => f.value).map(field => (
                                    <div key={field.label} style={{
                                        display: 'flex',
                                        justifyContent: 'space-between',
                                        gap: '8px',
                                    }}>
                    <span style={{
                        fontSize: '11px',
                        color: 'var(--color-text-muted)',
                        fontFamily: 'var(--font-body)',
                        flexShrink: 0,
                    }}>
                      {field.label}
                    </span>
                                        <span style={{
                                            fontSize: '11px',
                                            color: 'var(--color-text)',
                                            fontFamily: 'var(--font-body)',
                                            textAlign: 'right',
                                            textTransform: 'capitalize',
                                        }}>
                      {String(field.value)}
                    </span>
                                    </div>
                                ))}
                            </div>
                        </Card>
                    )}
                </div>

                {/* Right column — main info */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    {/* Title + authors */}
                    <div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap', marginBottom: '8px' }}>
                            <Badge
                                label={book.status}
                                variant={book.status === 'approved' ? 'success' : 'warning'}
                                size="sm"
                            />
                            {book.genres?.map(g => (
                                <Badge key={g.id} label={g.name} variant="info" size="sm" />
                            ))}
                        </div>

                        <h1 style={{
                            margin: '0 0 8px',
                            fontSize: '28px',
                            fontWeight: 700,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-heading)',
                            lineHeight: 1.2,
                        }}>
                            {book.title}
                        </h1>

                        <p style={{
                            margin: 0,
                            fontSize: '15px',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                        }}>
                            {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                        </p>
                    </div>

                    {/* Description */}
                    {book.description && (
                        <Card padding="md">
                            <p style={{
                                margin: 0,
                                fontSize: '14px',
                                color: 'var(--color-text)',
                                lineHeight: '1.7',
                                fontFamily: 'var(--font-body)',
                            }}>
                                {book.description}
                            </p>
                        </Card>
                    )}

                    {/* All editions */}
                    {book.editions && book.editions.length > 1 && (
                        <div>
                            <h3 style={{
                                margin: '0 0 12px',
                                fontSize: '14px',
                                fontWeight: 600,
                                color: 'var(--color-text)',
                                fontFamily: 'var(--font-heading)',
                            }}>
                                All Editions ({book.editions.length})
                            </h3>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                {book.editions.map(edition => (
                                    <Card key={edition.id} padding="sm">
                                        <div style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: '12px',
                                            flexWrap: 'wrap',
                                        }}>
                                            <Badge
                                                label={edition.format}
                                                variant="default"
                                                size="sm"
                                            />
                                            {edition.language && (
                                                <span style={{
                                                    fontSize: '12px',
                                                    color: 'var(--color-text-muted)',
                                                    fontFamily: 'var(--font-body)',
                                                }}>
                          {edition.language.toUpperCase()}
                        </span>
                                            )}
                                            {edition.publisher && (
                                                <span style={{
                                                    fontSize: '12px',
                                                    color: 'var(--color-text-muted)',
                                                    fontFamily: 'var(--font-body)',
                                                }}>
                          {edition.publisher}
                        </span>
                                            )}
                                            {edition.isbn && (
                                                <span style={{
                                                    fontSize: '11px',
                                                    color: 'var(--color-text-muted)',
                                                    fontFamily: 'var(--font-body)',
                                                    marginLeft: 'auto',
                                                }}>
                          ISBN: {edition.isbn}
                        </span>
                                            )}
                                        </div>
                                    </Card>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Action buttons */}
                    <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
                        <button
                            onClick={() => navigate('/books/add')}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '8px',
                                padding: '10px 20px',
                                background: 'var(--color-primary)',
                                color: 'var(--color-primary-text)',
                                border: 'none',
                                borderRadius: 'var(--border-radius)',
                                fontSize: '13px',
                                fontWeight: 600,
                                cursor: 'pointer',
                                transition: 'var(--transition)',
                                fontFamily: 'var(--font-body)',
                            }}
                            onMouseEnter={e => {
                                (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary-hover)'
                            }}
                            onMouseLeave={e => {
                                (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary)'
                            }}
                        >
                            ➕ Add to my library
                        </button>

                        <button
                            onClick={() => navigate('/books')}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '8px',
                                padding: '10px 20px',
                                background: 'var(--color-surface-alt)',
                                color: 'var(--color-text)',
                                border: '1px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)',
                                fontSize: '13px',
                                fontWeight: 600,
                                cursor: 'pointer',
                                transition: 'var(--transition)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            ← Back
                        </button>
                    </div>
                </div>
            </div>
        </div>
    )
}