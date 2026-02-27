import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import { Card, Badge, Spinner, EmptyState } from '../components'
import { getMyBooks, listBooks } from '../api/books'

export default function DashboardPage() {
    const { user } = useAuth()
    const navigate = useNavigate()

    const { data: myBooks, isLoading: loadingMyBooks } = useQuery({
        queryKey: ['my-books'],
        queryFn: () => getMyBooks(1, 4),
    })

    const { data: allBooks, isLoading: loadingAll } = useQuery({
        queryKey: ['books', 1],
        queryFn: () => listBooks(1, 6),
    })

    const hour = new Date().getHours()
    const greeting = hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening'

    return (
        <div
            style={{
                minHeight: 'calc(100vh - 56px)',
                padding: '32px 24px',
                maxWidth: '1100px',
                margin: '0 auto',
            }}
        >
            {/* Welcome header */}
            <div style={{ marginBottom: '32px' }}>
                <h1
                    style={{
                        margin: 0,
                        fontSize: '28px',
                        fontWeight: 700,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                    }}
                >
                    {greeting}, {user?.username} 👋
                </h1>
                <p style={{
                    margin: '6px 0 0',
                    fontSize: '14px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                }}>
                    Here's what's happening in your library today.
                </p>
            </div>

            {/* Stats row */}
            <div
                style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
                    gap: '16px',
                    marginBottom: '40px',
                }}
            >
                {[
                    {
                        icon: '📚',
                        label: 'My Books',
                        value: loadingMyBooks ? '…' : String(myBooks?.total ?? 0),
                        sub: 'in your library',
                        action: () => navigate('/library'),
                    },
                    {
                        icon: '🌐',
                        label: 'Catalogue',
                        value: loadingAll ? '…' : String(allBooks?.total ?? 0),
                        sub: 'approved books',
                        action: () => navigate('/books'),
                    },
                    {
                        icon: '🎭',
                        label: 'Role',
                        value: user?.role ?? '',
                        sub: 'your access level',
                        action: undefined,
                    },
                    {
                        icon: '🎨',
                        label: 'Theme',
                        value: user?.theme ?? 'woody',
                        sub: 'current look',
                        action: undefined,
                    },
                ].map(stat => (
                    <Card
                        key={stat.label}
                        hover={!!stat.action}
                        onClick={stat.action}
                        padding="md"
                    >
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                            <span style={{ fontSize: '24px' }}>{stat.icon}</span>
                            <div>
                                <p style={{
                                    margin: 0,
                                    fontSize: '22px',
                                    fontWeight: 700,
                                    color: 'var(--color-text)',
                                    fontFamily: 'var(--font-heading)',
                                    textTransform: 'capitalize',
                                }}>
                                    {stat.value}
                                </p>
                                <p style={{
                                    margin: '2px 0 0',
                                    fontSize: '12px',
                                    color: 'var(--color-text-muted)',
                                    fontFamily: 'var(--font-body)',
                                }}>
                                    {stat.label} · {stat.sub}
                                </p>
                            </div>
                        </div>
                    </Card>
                ))}
            </div>

            {/* Two column layout */}
            <div
                style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
                    gap: '24px',
                }}
            >
                {/* My recent books */}
                <Card padding="md">
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        marginBottom: '16px',
                    }}>
                        <h2 style={{
                            margin: 0,
                            fontSize: '15px',
                            fontWeight: 600,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-heading)',
                        }}>
                            My Library
                        </h2>
                        <button
                            onClick={() => navigate('/library')}
                            style={{
                                background: 'none',
                                border: 'none',
                                cursor: 'pointer',
                                fontSize: '12px',
                                color: 'var(--color-primary)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            View all →
                        </button>
                    </div>

                    {loadingMyBooks ? (
                        <div style={{ padding: '24px 0' }}>
                            <Spinner size="sm" />
                        </div>
                    ) : !myBooks?.books?.length ? (
                        <EmptyState
                            icon="📭"
                            title="No books yet"
                            description="Start building your library by adding your first book."
                            action={{ label: 'Browse catalogue', onClick: () => navigate('/books') }}
                        />
                    ) : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                            {myBooks.books.map(book => (
                                <div
                                    key={book.id}
                                    onClick={() => navigate(`/books/${book.id}`)}
                                    style={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: '12px',
                                        padding: '10px',
                                        borderRadius: 'var(--border-radius)',
                                        cursor: 'pointer',
                                        transition: 'var(--transition)',
                                    }}
                                    onMouseEnter={e => {
                                        (e.currentTarget as HTMLDivElement).style.background = 'var(--color-surface-alt)'
                                    }}
                                    onMouseLeave={e => {
                                        (e.currentTarget as HTMLDivElement).style.background = 'transparent'
                                    }}
                                >
                                    {/* Cover placeholder */}
                                    <div style={{
                                        width: '36px',
                                        height: '48px',
                                        borderRadius: '3px',
                                        background: 'var(--color-primary)',
                                        flexShrink: 0,
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        fontSize: '18px',
                                        opacity: 0.8,
                                    }}>
                                        📖
                                    </div>
                                    <div style={{ flex: 1, minWidth: 0 }}>
                                        <p style={{
                                            margin: 0,
                                            fontSize: '13px',
                                            fontWeight: 600,
                                            color: 'var(--color-text)',
                                            fontFamily: 'var(--font-body)',
                                            overflow: 'hidden',
                                            textOverflow: 'ellipsis',
                                            whiteSpace: 'nowrap',
                                        }}>
                                            {book.title}
                                        </p>
                                        <p style={{
                                            margin: '2px 0 0',
                                            fontSize: '11px',
                                            color: 'var(--color-text-muted)',
                                            fontFamily: 'var(--font-body)',
                                        }}>
                                            {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                                        </p>
                                    </div>
                                    <Badge label={book.status} variant="success" size="sm" />
                                </div>
                            ))}
                        </div>
                    )}
                </Card>

                {/* Recently added to catalogue */}
                <Card padding="md">
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        marginBottom: '16px',
                    }}>
                        <h2 style={{
                            margin: 0,
                            fontSize: '15px',
                            fontWeight: 600,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-heading)',
                        }}>
                            Recently Added
                        </h2>
                        <button
                            onClick={() => navigate('/books')}
                            style={{
                                background: 'none',
                                border: 'none',
                                cursor: 'pointer',
                                fontSize: '12px',
                                color: 'var(--color-primary)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            Browse all →
                        </button>
                    </div>

                    {loadingAll ? (
                        <div style={{ padding: '24px 0' }}>
                            <Spinner size="sm" />
                        </div>
                    ) : !allBooks?.books?.length ? (
                        <EmptyState
                            icon="🔍"
                            title="No books in catalogue"
                            description="Be the first to add a book to the Biblios catalogue."
                            action={{ label: 'Add a book', onClick: () => navigate('/books/submit') }}
                        />
                    ) : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                            {allBooks.books.map(book => (
                                <div
                                    key={book.id}
                                    onClick={() => navigate(`/books/${book.id}`)}
                                    style={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: '12px',
                                        padding: '10px',
                                        borderRadius: 'var(--border-radius)',
                                        cursor: 'pointer',
                                        transition: 'var(--transition)',
                                    }}
                                    onMouseEnter={e => {
                                        (e.currentTarget as HTMLDivElement).style.background = 'var(--color-surface-alt)'
                                    }}
                                    onMouseLeave={e => {
                                        (e.currentTarget as HTMLDivElement).style.background = 'transparent'
                                    }}
                                >
                                    <div style={{
                                        width: '36px',
                                        height: '48px',
                                        borderRadius: '3px',
                                        background: 'var(--color-accent)',
                                        flexShrink: 0,
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        fontSize: '18px',
                                        opacity: 0.8,
                                    }}>
                                        📚
                                    </div>
                                    <div style={{ flex: 1, minWidth: 0 }}>
                                        <p style={{
                                            margin: 0,
                                            fontSize: '13px',
                                            fontWeight: 600,
                                            color: 'var(--color-text)',
                                            fontFamily: 'var(--font-body)',
                                            overflow: 'hidden',
                                            textOverflow: 'ellipsis',
                                            whiteSpace: 'nowrap',
                                        }}>
                                            {book.title}
                                        </p>
                                        <p style={{
                                            margin: '2px 0 0',
                                            fontSize: '11px',
                                            color: 'var(--color-text-muted)',
                                            fontFamily: 'var(--font-body)',
                                        }}>
                                            {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                                        </p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </Card>

                {/* Quick actions */}
                <Card padding="md">
                    <h2 style={{
                        margin: '0 0 16px',
                        fontSize: '15px',
                        fontWeight: 600,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                    }}>
                        Quick Actions
                    </h2>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                        {[
                            { icon: '➕', label: 'Submit a new book', path: '/books/submit' },
                            { icon: '🔍', label: 'Browse the catalogue', path: '/books' },
                            { icon: '🗄️', label: 'View my library', path: '/library' },
                            ...(user?.role === 'moderator' || user?.role === 'admin'
                                    ? [{ icon: '🛡️', label: 'Moderation queue', path: '/moderation' }]
                                    : []
                            ),
                        ].map(action => (
                            <button
                                key={action.path}
                                onClick={() => navigate(action.path)}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '12px',
                                    padding: '12px',
                                    background: 'var(--color-surface-alt)',
                                    border: '1px solid var(--color-border)',
                                    borderRadius: 'var(--border-radius)',
                                    cursor: 'pointer',
                                    transition: 'var(--transition)',
                                    textAlign: 'left',
                                    width: '100%',
                                }}
                                onMouseEnter={e => {
                                    (e.currentTarget as HTMLButtonElement).style.borderColor = 'var(--color-primary)'
                                    ;(e.currentTarget as HTMLButtonElement).style.color = 'var(--color-primary)'
                                }}
                                onMouseLeave={e => {
                                    (e.currentTarget as HTMLButtonElement).style.borderColor = 'var(--color-border)'
                                    ;(e.currentTarget as HTMLButtonElement).style.color = 'var(--color-text)'
                                }}
                            >
                                <span style={{ fontSize: '18px' }}>{action.icon}</span>
                                <span style={{
                                    fontSize: '13px',
                                    fontWeight: 500,
                                    color: 'inherit',
                                    fontFamily: 'var(--font-body)',
                                }}>
                  {action.label}
                </span>
                                <span style={{
                                    marginLeft: 'auto',
                                    color: 'var(--color-text-muted)',
                                    fontSize: '12px',
                                }}>
                  →
                </span>
                            </button>
                        ))}
                    </div>
                </Card>
            </div>
        </div>
    )
}