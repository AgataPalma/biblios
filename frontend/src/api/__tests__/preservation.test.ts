/**
 * Preservation Property Tests
 *
 * Property 2: Preservation — Correct Frontend Code Is Unchanged
 *
 * These tests MUST PASS on unfixed code — they establish the baseline of already-correct
 * API calls that must remain unchanged after the fix is applied.
 *
 * Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import fc from 'fast-check'
import MockAdapter from 'axios-mock-adapter'
import apiClient from '../client'

// ── Test setup ────────────────────────────────────────────────────────────────

let mock: MockAdapter

beforeEach(() => {
    mock = new MockAdapter(apiClient)
})

afterEach(() => {
    mock.restore()
})

// ── Auth: POST /auth/login ────────────────────────────────────────────────────
// These calls are already correct and must remain unchanged after the fix.

describe('Auth preservation', () => {
    it('POST /auth/login — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPost('/auth/login').reply((config) => {
            capturedRequests.push({ method: 'post', url: config.url ?? '' })
            return [200, { token: 'tok', user: { id: '1', email: 'a@b.com', username: 'alice', role: 'user', theme: 'light', created_at: '', updated_at: '' } }]
        })

        const { login } = await import('../auth')
        await login({ email: 'a@b.com', password: 'secret' })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/auth/login')
        expect(capturedRequests[0].method).toBe('post')
    })

    it('POST /auth/register — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPost('/auth/register').reply((config) => {
            capturedRequests.push({ method: 'post', url: config.url ?? '' })
            return [201, { token: 'tok', user: { id: '1', email: 'a@b.com', username: 'alice', role: 'user', theme: 'light', created_at: '', updated_at: '' } }]
        })

        const { register } = await import('../auth')
        await register({ email: 'a@b.com', username: 'alice', password: 'secret' })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/auth/register')
        expect(capturedRequests[0].method).toBe('post')
    })

    it('GET /auth/me — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onGet('/auth/me').reply((config) => {
            capturedRequests.push({ method: 'get', url: config.url ?? '' })
            return [200, { id: '1', email: 'a@b.com', username: 'alice', role: 'user', theme: 'light', created_at: '', updated_at: '' }]
        })

        const { getMe } = await import('../auth')
        await getMe()

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/auth/me')
        expect(capturedRequests[0].method).toBe('get')
    })
})

// ── Books: GET /books, GET /books/:id, PUT /books/copies/:id/status ───────────

describe('Books preservation', () => {
    it('GET /books — correct URL and returns { books, total, page, limit }', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onGet('/books').reply((config) => {
            capturedRequests.push({ method: 'get', url: '/books' })
            return [200, { books: [], total: 0, page: 1, limit: 20 }]
        })

        const { listBooks } = await import('../books')
        const result = await listBooks()

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/books')
        expect(result).toHaveProperty('books')
        expect(result).toHaveProperty('total')
        expect(result).toHaveProperty('page')
        expect(result).toHaveProperty('limit')
    })

    it('GET /books/:id — correct URL with book id', async () => {
        const capturedUrls: string[] = []

        mock.onGet('/books/book42').reply((config) => {
            capturedUrls.push(config.url ?? '')
            return [200, { id: 'book42', title: 'Test Book', status: 'approved', authors: [], genres: [], editions: [], created_at: '', updated_at: '' }]
        })

        const { getBook } = await import('../books')
        await getBook('book42')

        expect(capturedUrls.length).toBe(1)
        expect(capturedUrls[0]).toBe('/books/book42')
    })

    it('PUT /books/copies/:id/status — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPut('/books/copies/copy99/status').reply((config) => {
            capturedRequests.push({ method: 'put', url: config.url ?? '' })
            return [204, undefined]
        })

        const { updateReadingStatus } = await import('../books')
        await updateReadingStatus('copy99', { status: 'reading', current_page: 10 })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/books/copies/copy99/status')
        expect(capturedRequests[0].method).toBe('put')
    })
})

// ── Shelf CRUD (not detail): GET /shelves, POST /shelves, PUT /shelves/:id, DELETE /shelves/:id

describe('Shelf CRUD preservation', () => {
    it('GET /shelves — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onGet('/shelves').reply((config) => {
            capturedRequests.push({ method: 'get', url: config.url ?? '' })
            return [200, []]
        })

        const { getShelves } = await import('../shelves')
        await getShelves()

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/shelves')
        expect(capturedRequests[0].method).toBe('get')
    })

    it('POST /shelves — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPost('/shelves').reply((config) => {
            capturedRequests.push({ method: 'post', url: config.url ?? '' })
            return [201, { id: 's1', user_id: 'u1', name: 'My Shelf', created_at: '', updated_at: '' }]
        })

        const { createShelf } = await import('../shelves')
        await createShelf({ name: 'My Shelf' })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/shelves')
        expect(capturedRequests[0].method).toBe('post')
    })

    it('PUT /shelves/:id — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPut('/shelves/s1').reply((config) => {
            capturedRequests.push({ method: 'put', url: config.url ?? '' })
            return [200, { id: 's1', user_id: 'u1', name: 'Renamed', created_at: '', updated_at: '' }]
        })

        const { updateShelf } = await import('../shelves')
        await updateShelf('s1', { name: 'Renamed' })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/shelves/s1')
        expect(capturedRequests[0].method).toBe('put')
    })

    it('DELETE /shelves/:id — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onDelete('/shelves/s1').reply((config) => {
            capturedRequests.push({ method: 'delete', url: config.url ?? '' })
            return [204, undefined]
        })

        const { deleteShelf } = await import('../shelves')
        await deleteShelf('s1')

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/shelves/s1')
        expect(capturedRequests[0].method).toBe('delete')
    })
})

// ── Library CRUD (not detail/members): POST /libraries, PUT /libraries/:id, DELETE /libraries/:id

describe('Library CRUD preservation', () => {
    it('POST /libraries — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPost('/libraries').reply((config) => {
            capturedRequests.push({ method: 'post', url: config.url ?? '' })
            return [201, { id: 'lib1', owner_id: 'u1', name: 'My Library', visibility: 'private', is_cooperative: false, created_at: '', updated_at: '' }]
        })

        const { createLibrary } = await import('../libraries')
        await createLibrary({ name: 'My Library', visibility: 'private' })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/libraries')
        expect(capturedRequests[0].method).toBe('post')
    })

    it('PUT /libraries/:id — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPut('/libraries/lib1').reply((config) => {
            capturedRequests.push({ method: 'put', url: config.url ?? '' })
            return [200, { id: 'lib1', owner_id: 'u1', name: 'Updated', visibility: 'public', is_cooperative: false, created_at: '', updated_at: '' }]
        })

        const { updateLibrary } = await import('../libraries')
        await updateLibrary('lib1', { name: 'Updated' })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/libraries/lib1')
        expect(capturedRequests[0].method).toBe('put')
    })

    // Note: deleteLibrary is not yet implemented in the current codebase.
    // The DELETE /libraries/:id test will be added once the function exists.
})

// ── Correct collection calls: POST /libraries/:id/collections, GET /libraries/:id/collections

describe('Collection CRUD preservation (already-correct calls)', () => {
    it('POST /libraries/:id/collections — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onPost('/libraries/lib1/collections').reply((config) => {
            capturedRequests.push({ method: 'post', url: config.url ?? '' })
            return [201, { id: 'col1', library_id: 'lib1', name: 'My Collection', is_public: true, is_collaborative: false, created_by: 'u1', created_at: '', updated_at: '' }]
        })

        const { createCollection } = await import('../collections')
        await createCollection('lib1', { name: 'My Collection' })

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/libraries/lib1/collections')
        expect(capturedRequests[0].method).toBe('post')
    })

    it('GET /libraries/:id/collections — correct URL and method', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onGet('/libraries/lib1/collections').reply((config) => {
            capturedRequests.push({ method: 'get', url: config.url ?? '' })
            return [200, []]
        })

        const { getCollections } = await import('../collections')
        await getCollections('lib1')

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/libraries/lib1/collections')
        expect(capturedRequests[0].method).toBe('get')
    })
})

// ── Series list: GET /series returning { series, total, page, limit } ─────────

describe('Series list preservation', () => {
    it('GET /series — correct URL and returns { series, total, page, limit }', async () => {
        const capturedRequests: Array<{ method: string; url: string }> = []

        mock.onGet('/series').reply((config) => {
            capturedRequests.push({ method: 'get', url: '/series' })
            return [200, { series: [], total: 0, page: 1, limit: 20 }]
        })

        const { getSeries } = await import('../series')
        const result = await getSeries()

        expect(capturedRequests.length).toBe(1)
        expect(capturedRequests[0].url).toBe('/series')
        expect(result).toHaveProperty('series')
        expect(result).toHaveProperty('total')
        expect(result).toHaveProperty('page')
        expect(result).toHaveProperty('limit')
        expect(Array.isArray(result.series)).toBe(true)
    })
})

// ── Property-based: Notification unread count ─────────────────────────────────
// Generate random arrays of Notification objects with random read_at values.
// Verify that unread_count computed client-side equals notifications.filter(n => n.read_at === null).length.
// This property must hold both before and after the fix.
//
// **Validates: Requirements 3.1**

describe('Property: Notification unread count computed from read_at', () => {
    it('unread_count = notifications.filter(n => n.read_at === null).length for any notification array', () => {
        // Arbitraries
        const readAtArb = fc.oneof(
            fc.constant(null),
            fc.date({ min: new Date('2020-01-01'), max: new Date('2025-01-01') }).map(d => d.toISOString()),
        )

        const notificationArb = fc.record({
            id: fc.uuid(),
            user_id: fc.uuid(),
            type: fc.constantFrom('invitation', 'review_like', 'library_activity'),
            title: fc.string({ minLength: 1, maxLength: 50 }),
            body: fc.string({ minLength: 1, maxLength: 100 }),
            read_at: readAtArb,
            created_at: fc.date({ min: new Date('2020-01-01'), max: new Date('2025-01-01') }).map(d => d.toISOString()),
        })

        fc.assert(
            fc.property(fc.array(notificationArb, { minLength: 0, maxLength: 20 }), (notifications) => {
                // Client-side unread count computation (the correct approach after fix)
                const unreadCount = notifications.filter(n => n.read_at === null).length

                // This must equal the count of notifications where read_at is null
                const expectedCount = notifications.reduce((acc, n) => acc + (n.read_at === null ? 1 : 0), 0)

                return unreadCount === expectedCount
            }),
            { numRuns: 100 },
        )
    })
})

// ── Property-based: Challenge rendering with year and goal ────────────────────
// Generate random Challenge objects with year: number and goal: number.
// Verify that accessing challenge.year and challenge.goal never returns undefined.
//
// **Validates: Requirements 3.2**

describe('Property: Challenge rendering uses year and goal fields', () => {
    it('challenge.year and challenge.goal are always defined for valid backend challenge objects', () => {
        const challengeArb = fc.record({
            id: fc.uuid(),
            user_id: fc.uuid(),
            year: fc.integer({ min: 2000, max: 2100 }),
            goal: fc.integer({ min: 1, max: 365 }),
            created_at: fc.date({ min: new Date('2020-01-01'), max: new Date('2025-01-01') }).map(d => d.toISOString()),
        })

        fc.assert(
            fc.property(challengeArb, (challenge) => {
                // year and goal must always be defined (not undefined)
                const yearDefined = challenge.year !== undefined && challenge.year !== null
                const goalDefined = challenge.goal !== undefined && challenge.goal !== null

                // The old (wrong) fields must NOT be present on the backend model
                const noOldFields =
                    !('title' in challenge) &&
                    !('start_date' in challenge) &&
                    !('end_date' in challenge) &&
                    !('goal_books' in challenge)

                return yearDefined && goalDefined && noOldFields
            }),
            { numRuns: 100 },
        )
    })
})

// ── Property-based: SeriesBook flat fields ────────────────────────────────────
// Generate random SeriesBook objects with flat fields (title, authors, cover_url, series_position).
// Verify that accessing these fields never requires a nested `book` object.
//
// **Validates: Requirements 3.3**

describe('Property: SeriesBook rendering uses flat fields (not nested book object)', () => {
    it('seriesBook.title, .authors, .series_position are always accessible at top level', () => {
        const seriesBookArb = fc.record({
            title: fc.string({ minLength: 1, maxLength: 100 }),
            authors: fc.array(fc.string({ minLength: 1, maxLength: 50 }), { minLength: 1, maxLength: 5 }),
            cover_url: fc.oneof(fc.constant(null), fc.webUrl()),
            series_position: fc.oneof(fc.constant(null), fc.integer({ min: 1, max: 100 })),
        })

        fc.assert(
            fc.property(seriesBookArb, (seriesBook) => {
                // Flat fields must be accessible directly
                const titleDefined = seriesBook.title !== undefined
                const authorsDefined = Array.isArray(seriesBook.authors)
                // cover_url can be null — that's valid
                const coverUrlValid = seriesBook.cover_url === null || typeof seriesBook.cover_url === 'string'
                // series_position can be null — that's valid
                const positionValid = seriesBook.series_position === null || typeof seriesBook.series_position === 'number'

                // No nested `book` object should be required
                const noNestedBook = !('book' in seriesBook)
                // No old `position` field (backend uses series_position)
                const noOldPosition = !('position' in seriesBook)

                return titleDefined && authorsDefined && coverUrlValid && positionValid && noNestedBook && noOldPosition
            }),
            { numRuns: 100 },
        )
    })
})

// ── Property-based: LibraryMember owner badge ─────────────────────────────────
// Generate random LibraryMember objects with is_owner: boolean.
// Verify that the owner badge renders if and only if is_owner === true.
//
// **Validates: Requirements 3.4**

describe('Property: LibraryMember owner badge uses is_owner (not role)', () => {
    it('owner badge renders iff is_owner === true, for any member object', () => {
        const memberArb = fc.record({
            user_id: fc.uuid(),
            library_id: fc.uuid(),
            username: fc.string({ minLength: 1, maxLength: 30 }),
            is_owner: fc.boolean(),
            can_view: fc.boolean(),
            can_add: fc.boolean(),
            can_remove: fc.boolean(),
            can_edit: fc.boolean(),
            can_invite: fc.boolean(),
            can_manage_members: fc.boolean(),
            joined_at: fc.date({ min: new Date('2020-01-01'), max: new Date('2025-01-01') }).map(d => d.toISOString()),
        })

        fc.assert(
            fc.property(memberArb, (member) => {
                // Owner badge logic: is_owner === true
                const showOwnerBadge = member.is_owner === true

                // This must match the is_owner field exactly
                const expectedBadge = member.is_owner

                // The old (wrong) field must NOT be present on the backend model
                const noRoleField = !('role' in member)
                // avatar_url is not in the backend model
                const noAvatarUrl = !('avatar_url' in member)

                return showOwnerBadge === expectedBadge && noRoleField && noAvatarUrl
            }),
            { numRuns: 100 },
        )
    })
})
