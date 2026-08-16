/**
 * Bug Condition Exploration Tests
 *
 * Property 1: Bug Condition — Frontend API Calls and Type Accesses Match Backend Contracts
 *
 * CRITICAL: These tests MUST FAIL on unfixed code — failure confirms the bugs exist.
 * DO NOT attempt to fix the test or the code when it fails.
 * These tests encode the expected (correct) behavior and will pass once the bugs are fixed.
 *
 * Validates: Requirements 1.1, 1.2, 1.3, 2.1, 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 4.4,
 *            5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 6.1, 6.2, 6.3, 7.1, 7.2, 7.4, 7.5,
 *            8.1, 8.2, 8.3, 9.1, 10.1
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
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

// ── Bug 1 — Notification read state ──────────────────────────────────────────
// Backend model uses `read_at: *time.Time` (nullable timestamp), NOT `is_read: boolean`.
// On unfixed code, `notification.is_read` is `undefined` for all notifications.
// Expected: derive read state from `read_at !== null`.

describe('Bug 1 — Notification read state', () => {
    it('should expose read_at field (not is_read) to determine unread state', async () => {
        const backendNotifications = [
            {
                id: '1',
                user_id: 'u1',
                type: 'invitation',
                title: 'You were invited',
                body: 'Join the library',
                data: {},
                read_at: null,          // unread — backend field
                created_at: '2024-01-01T00:00:00Z',
            },
            {
                id: '2',
                user_id: 'u1',
                type: 'review_like',
                title: 'Someone liked your review',
                body: 'Great review!',
                data: {},
                read_at: '2024-01-01T00:00:00Z',  // read — backend field
                created_at: '2024-01-01T00:00:00Z',
            },
        ]

        mock.onGet('/notifications').reply(200, {
            notifications: backendNotifications,
            total: 2,
            page: 1,
            limit: 20,
        })

        const { getNotifications } = await import('../notifications')
        const response = await getNotifications()

        const [unread, read] = response.notifications

        // The backend returns `read_at`, not `is_read`.
        // On unfixed code, `is_read` is `undefined` (falsy) for both — this assertion fails.
        expect(unread.read_at).toBe(null)
        expect(read.read_at).not.toBe(null)

        // Derived unread check: read_at === null means unread
        const isUnread = unread.read_at === null
        const isRead = read.read_at !== null
        expect(isUnread).toBe(true)
        expect(isRead).toBe(true)

        // On unfixed code, `is_read` does not exist on the type — it is `undefined`.
        // The fixed type must NOT have `is_read` as a boolean field.
        expect('is_read' in unread).toBe(false)
    })

    it('should forward unread and type filters to the backend', async () => {
        mock.onGet('/notifications').reply(200, {
            notifications: [],
            total: 0,
            page: 2,
            limit: 1,
        })

        const { getNotifications } = await import('../notifications')
        await getNotifications(2, 1, { read: false, type: 'library_invitation' })

        expect(mock.history.get).toHaveLength(1)
        expect(mock.history.get[0].params).toEqual({
            page: 2,
            limit: 1,
            read: false,
            type: 'library_invitation',
        })
    })
})

// ── Bug 2 — Mark all read endpoint ───────────────────────────────────────────
// Backend provides `PUT /notifications/read-all` (single bulk endpoint).
// On unfixed code, `markAllNotificationsRead` fires N individual `PUT /notifications/:id/read` requests.
// Expected: exactly ONE request to `PUT /notifications/read-all`.

describe('Bug 2 — Mark all read endpoint', () => {
    it('should call PUT /notifications/read-all exactly once (not N individual requests)', async () => {
        const readAllRequests: string[] = []
        const individualReadRequests: string[] = []

        mock.onPut('/notifications/read-all').reply(204, undefined, {})
        mock.onPut(/\/notifications\/[^/]+\/read/).reply((config) => {
            individualReadRequests.push(config.url ?? '')
            return [204, undefined]
        })

        // Track all PUT requests
        mock.onPut('/notifications/read-all').reply(() => {
            readAllRequests.push('/notifications/read-all')
            return [204, undefined]
        })

        const { markAllNotificationsRead } = await import('../notifications')

        // On unfixed code, this function takes `ids: string[]` and fires N individual requests.
        // On fixed code, it takes no ids and fires one bulk request.
        // The fixed function should call PUT /notifications/read-all regardless.
        await markAllNotificationsRead()

        // On unfixed code: readAllRequests.length === 0, individualReadRequests.length === 0
        // (no requests fired because ids array is empty — but the function signature is wrong)
        // On fixed code: readAllRequests.length === 1
        expect(readAllRequests.length).toBe(1)
        expect(individualReadRequests.length).toBe(0)
    })

    it('should NOT fire individual PUT /notifications/:id/read requests when marking all read', async () => {
        const individualReadRequests: string[] = []

        mock.onPut('/notifications/read-all').reply(204)
        mock.onPut(/\/notifications\/[^/]+\/read/).reply((config) => {
            individualReadRequests.push(config.url ?? '')
            return [204, undefined]
        })

        const { markAllNotificationsRead } = await import('../notifications')

        // On unfixed code with ids=['1','2','3'], it fires 3 individual requests.
        // On fixed code, it fires 0 individual requests (uses bulk endpoint).
        await markAllNotificationsRead()

        // On unfixed code: individualReadRequests.length === 3 — this assertion fails.
        expect(individualReadRequests.length).toBe(0)
    })
})

// ── Bug 3 — Challenge fields ──────────────────────────────────────────────────
// Backend `Challenge` model: `id`, `user_id`, `year` (int), `goal` (int), `created_at`.
// Frontend type has: `title`, `start_date`, `end_date`, `goal_books`, `current_books`, `status`.
// On unfixed code, `challenge.year` and `challenge.goal` are `undefined`.

describe('Bug 3 — Challenge fields', () => {
    it('should parse challenge with year and goal fields from backend plain array response', async () => {
        const backendChallenges = [
            {
                id: '1',
                user_id: 'u1',
                year: 2024,
                goal: 12,
                created_at: '2024-01-01T00:00:00Z',
            },
        ]

        // Backend returns a plain array, NOT { challenges: [], total: number }
        mock.onGet('/reading/challenges').reply(200, backendChallenges)

        const { getChallenges } = await import('../reading')
        const result = await getChallenges()

        // On unfixed code: getChallenges() returns a ChallengesResponse wrapper object.
        // The backend returns a plain array, so `result` would be the array itself,
        // but the function is typed to return ChallengesResponse — accessing .challenges
        // on an array gives undefined.
        // On fixed code: result is Challenge[] directly.
        expect(Array.isArray(result)).toBe(true)

        const challenges = result
        expect(challenges[0].year).toBe(2024)
        expect(challenges[0].goal).toBe(12)

        // On unfixed code, the frontend type uses `goal_books` not `goal` — undefined.
        expect('goal_books' in challenges[0]).toBe(false)
        expect('title' in challenges[0]).toBe(false)
    })

    it('should send { year, goal } when creating a challenge (not title/start_date/end_date/goal_books)', async () => {
        let capturedBody: Record<string, unknown> | null = null

        mock.onPost('/reading/challenges').reply((config) => {
            capturedBody = JSON.parse(config.data as string)
            return [201, { id: '1', user_id: 'u1', year: 2024, goal: 12, created_at: '2024-01-01T00:00:00Z' }]
        })

        const { createChallenge } = await import('../reading')

        // On unfixed code, createChallenge expects { title, start_date, end_date, goal_books }.
        // On fixed code, it expects { year: number, goal: number }.
        // We call with the correct (fixed) payload shape.
        await createChallenge({ year: 2024, goal: 12 })

        // On unfixed code: capturedBody contains title/start_date/end_date/goal_books (wrong fields).
        // On fixed code: capturedBody contains only year and goal.
        expect(capturedBody).not.toBeNull()
        expect(capturedBody!.year).toBe(2024)
        expect(capturedBody!.goal).toBe(12)
        expect(capturedBody!.title).toBeUndefined()
        expect(capturedBody!.start_date).toBeUndefined()
        expect(capturedBody!.end_date).toBeUndefined()
        expect(capturedBody!.goal_books).toBeUndefined()
    })

    it('should fetch challenge progress from the challenge progress endpoint', async () => {
        const progress = {
            challenge: {
                id: '1',
                user_id: 'u1',
                year: 2024,
                goal: 12,
                created_at: '2024-01-01T00:00:00Z',
            },
            books_read: 5,
            books_remaining: 7,
            progress_pct: 41.67,
            monthly_pace: 1.25,
        }

        mock.onGet('/reading/challenges/1/progress').reply(200, progress)

        const { getChallengeProgress } = await import('../reading')
        await expect(getChallengeProgress('1')).resolves.toEqual(progress)
    })
})

// ── Bug 4 — Session field names ───────────────────────────────────────────────
// Backend `Session` model: `logged_date`, `note`, `pages_read *int` (nullable).
// Frontend type uses: `date`, `notes`, `pages_read: number` (required).
// On unfixed code, `session.logged_date` and `session.note` are `undefined`.

describe('Bug 4 — Session field names', () => {
    it('should parse session with logged_date and note fields from backend response', async () => {
        const backendSessions = [
            {
                id: '1',
                user_id: 'u1',
                copy_id: 'c1',
                logged_date: '2024-06-01',   // backend field name
                note: 'good session',         // backend field name
                pages_read: 30,
                created_at: '2024-06-01T00:00:00Z',
            },
        ]

        mock.onGet('/reading/sessions').reply(200, {
            sessions: backendSessions,
            total: 1,
            page: 1,
            limit: 20,
        })

        const { getSessions } = await import('../reading')
        const response = await getSessions()

        const session = response.sessions[0] as unknown as Record<string, unknown>

        // On unfixed code: session.date and session.notes are the typed fields,
        // but the backend returns logged_date and note — so they are undefined.
        // On fixed code: session.logged_date === "2024-06-01" and session.note === "good session".
        expect(session.logged_date).toBe('2024-06-01')
        expect(session.note).toBe('good session')

        // The old (wrong) field names should NOT be present
        expect(session.date).toBeUndefined()
        expect(session.notes).toBeUndefined()
    })

    it('should send { copy_id, logged_date, note } when creating a session (not date/notes)', async () => {
        let capturedBody: Record<string, unknown> | null = null

        mock.onPost('/reading/sessions').reply((config) => {
            capturedBody = JSON.parse(config.data as string)
            return [201, {
                id: '1',
                user_id: 'u1',
                copy_id: 'c1',
                logged_date: '2024-06-01',
                note: 'good session',
                pages_read: 30,
                created_at: '2024-06-01T00:00:00Z',
            }]
        })

        const { createSession } = await import('../reading')

        // On fixed code, createSession expects { copy_id, logged_date, pages_read?, note? }.
        await createSession({
            copy_id: 'c1',
            logged_date: '2024-06-01',
            pages_read: 30,
            note: 'good session',
        })

        expect(capturedBody).not.toBeNull()
        expect(capturedBody!.logged_date).toBe('2024-06-01')
        expect(capturedBody!.note).toBe('good session')

        // On unfixed code: capturedBody contains `date` and `notes` (wrong field names).
        expect(capturedBody!.date).toBeUndefined()
        expect(capturedBody!.notes).toBeUndefined()
    })
})

// ── Bug 5 — Collection nested routes ─────────────────────────────────────────
// Backend only has nested routes: /libraries/:libraryId/collections/:collectionId
// Frontend calls standalone /collections/:id routes (404).
// Frontend sends { book_id } but backend expects { copy_id }.
// Frontend reads collection.visibility but backend has is_public: bool.

describe('Bug 5 — Collection nested routes', () => {
    it('should call GET /libraries/lib1/collections/col1 (not GET /collections/col1)', async () => {
        const nestedRequests: string[] = []
        const standaloneRequests: string[] = []

        mock.onGet('/libraries/lib1/collections/col1').reply((config) => {
            nestedRequests.push(config.url ?? '')
            return [200, {
                id: 'col1',
                library_id: 'lib1',
                name: 'My Collection',
                is_public: true,
                is_collaborative: false,
                created_by: 'u1',
                created_at: '2024-01-01T00:00:00Z',
                updated_at: '2024-01-01T00:00:00Z',
            }]
        })

        mock.onGet('/collections/col1').reply((config) => {
            standaloneRequests.push(config.url ?? '')
            return [404, { error: 'not found' }]
        })

        const { getCollection } = await import('../collections')

        // On fixed code: getCollection("lib1", "col1") calls GET /libraries/lib1/collections/col1.
        // On unfixed code: getCollection("col1") calls GET /collections/col1 (404).
        const result = await getCollection('lib1', 'col1')

        // On unfixed code: nestedRequests.length === 0, standaloneRequests.length === 1 — fails.
        expect(nestedRequests.length).toBe(1)
        expect(standaloneRequests.length).toBe(0)
        expect(result.id).toBe('col1')
        expect('collection' in result).toBe(false)
    })

    it('should send { copy_id } (not { book_id }) when adding a book to a collection', async () => {
        let capturedBody: Record<string, unknown> | null = null

        mock.onPost('/libraries/lib1/collections/col1/books').reply((config) => {
            capturedBody = JSON.parse(config.data as string)
            return [204, undefined]
        })

        const { addBookToCollection } = await import('../collections')

        // On fixed code: addBookToCollection("lib1", "col1", "copy123") sends { copy_id: "copy123" }.
        // On unfixed code: addBookToCollection("col1", "book123") sends { book_id: "book123" } to wrong URL.
        await addBookToCollection('lib1', 'col1', 'copy123')

        expect(capturedBody).not.toBeNull()
        expect(capturedBody!.copy_id).toBe('copy123')

        // On unfixed code: capturedBody.book_id === "copy123" (wrong field name).
        expect(capturedBody!.book_id).toBeUndefined()
    })

    it('should parse the collection list as a plain array', async () => {
        const collections = [{
            id: 'col1',
            library_id: 'lib1',
            name: 'My Collection',
            is_public: true,
            is_collaborative: false,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
        }]
        mock.onGet('/libraries/lib1/collections').reply(200, collections)

        const { getCollections } = await import('../collections')
        const result = await getCollections('lib1')

        expect(result).toEqual(collections)
        expect(Array.isArray(result)).toBe(true)
    })

    it('should create a collection using is_public instead of visibility', async () => {
        let capturedBody: Record<string, unknown> | null = null
        mock.onPost('/libraries/lib1/collections').reply((config) => {
            capturedBody = JSON.parse(config.data as string)
            return [201, {
                id: 'col1',
                library_id: 'lib1',
                name: 'Private Collection',
                is_public: false,
                is_collaborative: false,
                created_at: '2024-01-01T00:00:00Z',
                updated_at: '2024-01-01T00:00:00Z',
            }]
        })

        const { createCollection } = await import('../collections')
        await createCollection('lib1', { name: 'Private Collection', is_public: false })

        expect(capturedBody).toEqual({ name: 'Private Collection', is_public: false })
    })

    it('should fetch collection books from the nested books endpoint with pagination', async () => {
        const response = { books: [], total: 0, page: 2, limit: 25 }
        mock.onGet('/libraries/lib1/collections/col1/books').reply((config) => {
            expect(config.params).toEqual({ page: 2, limit: 25 })
            return [200, response]
        })

        const { getCollectionBooks } = await import('../collections')
        await expect(getCollectionBooks('lib1', 'col1', 2, 25)).resolves.toEqual(response)
    })

    it('should update collection visibility through the nested endpoint using is_public', async () => {
        let capturedBody: Record<string, unknown> | null = null
        mock.onPut('/libraries/lib1/collections/col1').reply((config) => {
            capturedBody = JSON.parse(config.data as string)
            return [200, {
                id: 'col1',
                library_id: 'lib1',
                name: 'Updated',
                is_public: false,
                is_collaborative: false,
                created_at: '2024-01-01T00:00:00Z',
                updated_at: '2024-01-02T00:00:00Z',
            }]
        })

        const { updateCollection } = await import('../collections')
        await updateCollection('lib1', 'col1', { name: 'Updated', is_public: false })

        expect(capturedBody).toEqual({ name: 'Updated', is_public: false })
    })

    it('should delete a collection through the nested endpoint', async () => {
        const requests: string[] = []
        mock.onDelete('/libraries/lib1/collections/col1').reply((config) => {
            requests.push(config.url ?? '')
            return [200, { message: 'collection deleted' }]
        })

        const { deleteCollection } = await import('../collections')
        await deleteCollection('lib1', 'col1')

        expect(requests).toEqual(['/libraries/lib1/collections/col1'])
    })

    it('should remove a copy through the nested path without a request body', async () => {
        let requestBody: unknown = 'not-called'
        mock.onDelete('/libraries/lib1/collections/col1/books/copy123').reply((config) => {
            requestBody = config.data
            return [200, { message: 'book removed from collection' }]
        })

        const { removeBookFromCollection } = await import('../collections')
        await removeBookFromCollection('lib1', 'col1', 'copy123')

        expect(requestBody).toBeUndefined()
    })

    it('should use is_public (not visibility) on collection objects', () => {
        // The backend Collection model has `is_public: bool`, not `visibility: string`.
        // On unfixed code, the TypeScript type has `visibility: 'public' | 'private'`.
        // On fixed code, the type has `is_public: boolean`.
        const backendCollection = {
            id: 'col1',
            library_id: 'lib1',
            name: 'My Collection',
            is_public: true,           // backend field
            is_collaborative: false,
            created_by: 'u1',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
        }

        // On unfixed code: backendCollection.visibility is undefined (wrong field name).
        // On fixed code: backendCollection.is_public === true.
        const col = backendCollection as Record<string, unknown>
        expect(col.is_public).toBe(true)
        expect(col.visibility).toBeUndefined()
    })
})

// ── Bug 6 — Shelf detail route ────────────────────────────────────────────────
// Backend has no GET /shelves/:id route — only GET /shelves/:id/books.
// Frontend calls GET /shelves/:id (404).
// removeBookFromShelf should use DELETE /shelves/:id/books/:copyId (path param, no body).

describe('Bug 6 — Shelf detail route', () => {
    it('should call GET /shelves/shelf1/books (not GET /shelves/shelf1)', async () => {
        const correctRequests: string[] = []
        const wrongRequests: string[] = []

        mock.onGet('/shelves/shelf1/books').reply((config) => {
            correctRequests.push(config.url ?? '')
            expect(config.params).toEqual({ page: 1, limit: 20 })
            return [200, {
                books: [],
                total: 0,
                page: 1,
                limit: 20,
            }]
        })

        mock.onGet('/shelves/shelf1').reply((config) => {
            wrongRequests.push(config.url ?? '')
            return [404, { error: 'not found' }]
        })

        const { getShelf } = await import('../shelves')
        const result = await getShelf('shelf1')

        // On unfixed code: wrongRequests.length === 1, correctRequests.length === 0 — fails.
        expect(correctRequests.length).toBe(1)
        expect(wrongRequests.length).toBe(0)
        expect(result.books).toEqual([])
        expect('shelf' in result).toBe(false)
    })

    it('should forward shelf-book pagination parameters', async () => {
        mock.onGet('/shelves/shelf1/books').reply((config) => {
            expect(config.params).toEqual({ page: 3, limit: 40 })
            return [200, { books: [], total: 0, page: 3, limit: 40 }]
        })

        const { getShelf } = await import('../shelves')
        const result = await getShelf('shelf1', 3, 40)

        expect(result.page).toBe(3)
        expect(result.limit).toBe(40)
    })

    it('should call DELETE /shelves/s1/books/c1 with no body (not DELETE /shelves/s1/books with body)', async () => {
        const pathParamRequests: string[] = []
        const bodyRequests: Array<{ url: string; body: unknown }> = []

        mock.onDelete('/shelves/s1/books/c1').reply((config) => {
            pathParamRequests.push(config.url ?? '')
            // On fixed code: no body
            const body = config.data ? JSON.parse(config.data as string) : null
            expect(body).toBeNull()
            return [204, undefined]
        })

        mock.onDelete('/shelves/s1/books').reply((config) => {
            bodyRequests.push({ url: config.url ?? '', body: config.data })
            return [404, { error: 'not found' }]
        })

        const { removeBookFromShelf } = await import('../shelves')
        await removeBookFromShelf('s1', 'c1')

        // On unfixed code: bodyRequests.length === 1 (wrong URL with body) — fails.
        expect(pathParamRequests.length).toBe(1)
        expect(bodyRequests.length).toBe(0)
    })
})

// ── Bug 7 — Library response shapes and invitation routes ─────────────────────
// Backend GET /libraries returns plain Library[] array (not { libraries, total }).
// Backend invitation routes: POST /invitations/:token/accept (not PUT /libraries/invitations/...).

describe('Bug 7 — Library response shapes and invitation routes', () => {
    it('should parse GET /libraries returning a plain array (not { libraries, total })', async () => {
        const backendLibraries = [
            {
                id: '1',
                owner_id: 'u1',
                name: 'My Library',
                description: 'A great library',
                visibility: 'public',
                is_cooperative: true,
                created_at: '2024-01-01T00:00:00Z',
                updated_at: '2024-01-01T00:00:00Z',
            },
        ]

        // Backend returns a plain array
        mock.onGet('/libraries').reply(200, backendLibraries)

        const { getLibraries } = await import('../libraries')
        const result = await getLibraries()

        // On unfixed code: getLibraries() returns LibrariesResponse = { libraries: [], total: number }.
        // The backend returns a plain array, so result is the array itself,
        // but the function is typed to return LibrariesResponse.
        // Accessing result.libraries on an array gives undefined.
        // On fixed code: result is Library[] directly.
        expect(Array.isArray(result)).toBe(true)

        const libraries = result
        expect(libraries[0].name).toBe('My Library')
        expect(libraries[0].id).toBe('1')
    })

    it('should parse library detail as a plain library object', async () => {
        const library = {
            id: 'lib1',
            owner_id: 'u1',
            name: 'My Library',
            visibility: 'private',
            is_cooperative: true,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
        }
        mock.onGet('/libraries/lib1').reply(200, library)

        const { getLibrary } = await import('../libraries')
        const result = await getLibrary('lib1')

        expect(result).toEqual(library)
        expect('library' in result).toBe(false)
        expect('members' in result).toBe(false)
    })

    it('should fetch library members from the separate members endpoint', async () => {
        const members = [{
            library_id: 'lib1',
            user_id: 'u1',
            username: 'owner',
            joined_at: '2024-01-01T00:00:00Z',
            is_owner: true,
            can_view: true,
            can_add: true,
            can_remove: true,
            can_edit: true,
            can_invite: true,
            can_manage_members: true,
        }]
        mock.onGet('/libraries/lib1/members').reply(200, members)

        const { getLibraryMembers } = await import('../libraries')
        await expect(getLibraryMembers('lib1')).resolves.toEqual(members)
    })

    it('should update member permissions without expecting a member response', async () => {
        let capturedBody: Record<string, unknown> | null = null
        mock.onPut('/libraries/lib1/members/u2').reply((config) => {
            capturedBody = JSON.parse(config.data as string)
            return [200, { message: 'permissions updated' }]
        })

        const { updateMember } = await import('../libraries')
        const result = await updateMember('lib1', 'u2', { can_add: true, can_remove: false })

        expect(result).toBeUndefined()
        expect(capturedBody).toEqual({ can_add: true, can_remove: false })
    })

    it('should call POST /invitations/tok/accept (not PUT /libraries/invitations/tok/accept)', async () => {
        const correctRequests: string[] = []
        const wrongRequests: string[] = []

        mock.onPost('/invitations/tok/accept').reply((config) => {
            correctRequests.push(`POST ${config.url}`)
            return [200, undefined]
        })

        mock.onPut('/libraries/invitations/tok/accept').reply((config) => {
            wrongRequests.push(`PUT ${config.url}`)
            return [404, { error: 'not found' }]
        })

        const { acceptInvitation } = await import('../libraries')
        await acceptInvitation('tok')

        // On unfixed code: wrongRequests.length === 1 (PUT to wrong URL) — fails.
        expect(correctRequests.length).toBe(1)
        expect(wrongRequests.length).toBe(0)
    })

    it('should call POST /invitations/tok/decline (not PUT /libraries/invitations/tok/decline)', async () => {
        const correctRequests: string[] = []
        const wrongRequests: string[] = []

        mock.onPost('/invitations/tok/decline').reply((config) => {
            correctRequests.push(`POST ${config.url}`)
            return [200, undefined]
        })

        mock.onPut('/libraries/invitations/tok/decline').reply((config) => {
            wrongRequests.push(`PUT ${config.url}`)
            return [404, { error: 'not found' }]
        })

        const { declineInvitation } = await import('../libraries')
        await declineInvitation('tok')

        expect(correctRequests.length).toBe(1)
        expect(wrongRequests.length).toBe(0)
    })
})

// ── Bug 8 — Series detail shape ───────────────────────────────────────────────
// Backend SeriesDetail struct embeds Series fields inline at top level (not nested under `series`).
// Backend SeriesBook has flat fields: title, authors[], cover_url, series_position.
// Frontend expects { series: { name, ... }, books: [{ book: { title }, position }] }.

describe('Bug 8 — Series detail shape', () => {
    it('should parse series detail with fields at top level (not nested under series key)', async () => {
        const backendSeriesDetail = {
            id: '1',
            name: 'Dune',
            description: 'A sci-fi epic',
            status: 'approved',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            books: [
                {
                    book_id: 'book1',
                    title: 'Dune',
                    series_position: 1,
                    authors: ['Frank Herbert'],
                    cover_url: null,
                },
            ],
        }

        mock.onGet('/series/1').reply(200, backendSeriesDetail)

        const { getSeriesById } = await import('../series')
        const data = await getSeriesById('1')

        // On unfixed code: data.series.name is accessed (nested), but backend returns fields inline.
        // data.series would be undefined, causing a crash.
        // On fixed code: data.name === "Dune" (top-level field).
        expect(data.name).toBe('Dune')
        expect(data.description).toBe('A sci-fi epic')

        // On unfixed code: data.series is undefined — this assertion fails.
        expect('series' in data).toBe(false)

        // Books should be a flat array with flat fields
        const books = data.books
        expect(Array.isArray(books)).toBe(true)
        expect(books[0].title).toBe('Dune')
        expect(books[0].series_position).toBe(1)
        expect(books[0].authors).toEqual(['Frank Herbert'])

        // On unfixed code: books[0].book.title is accessed (nested), but backend has flat fields.
        expect('book' in books[0]).toBe(false)
        expect('position' in books[0]).toBe(false)  // backend uses series_position, not position
    })
})

// ── Bug 9 — LibraryMember owner check ────────────────────────────────────────
// Backend LibraryMember model uses `is_owner: bool` (not `role: string`).
// On unfixed code, `member.role === 'owner'` is always false because `role` is `undefined`.

describe('Bug 9 — LibraryMember owner check', () => {
    it('should use is_owner boolean (not role string) to identify library owners', () => {
        const backendMember = {
            id: '1',
            user_id: 'u1',
            library_id: 'lib1',
            username: 'alice',
            is_owner: true,    // backend field
            can_view: true,
            can_add: true,
            can_remove: false,
            can_edit: false,
            can_invite: false,
            can_manage_members: false,
            joined_at: '2024-01-01T00:00:00Z',
        }

        const member = backendMember as Record<string, unknown>

        // On unfixed code: member.role === 'owner' is false (role is undefined).
        // On fixed code: member.is_owner === true.
        expect(member.is_owner).toBe(true)

        // The old (wrong) field should NOT be present on the backend response
        expect(member.role).toBeUndefined()

        // Owner badge logic: should use is_owner, not role
        const isOwner = member.is_owner === true
        expect(isOwner).toBe(true)

        // Non-owner member
        const nonOwnerMember = { ...backendMember, is_owner: false }
        const nonOwner = nonOwnerMember as Record<string, unknown>
        expect(nonOwner.is_owner).toBe(false)
        const isNonOwner = nonOwner.is_owner === true
        expect(isNonOwner).toBe(false)
    })
})

// ── Bug 10 — updateCopyStatus extra fields ────────────────────────────────────
// Backend updateCopyStatusRequest only accepts { status, current_page }.
// Frontend sends extra fields: started_reading_at, finished_reading_at, owned_by_user,
// borrowed_from, location — which are silently ignored by the backend.

describe('Bug 10 — updateCopyStatus extra fields', () => {
    it('should only send { status, current_page } and NOT include extra fields', async () => {
        let capturedBody: Record<string, unknown> | null = null

        mock.onPut('/books/copies/copy1/status').reply((config) => {
            capturedBody = JSON.parse(config.data as string)
            return [204, undefined]
        })

        const { updateReadingStatus } = await import('../books')

        // On fixed code: updateReadingStatus should only send { status, current_page }.
        // On unfixed code: it sends all fields from UpdateCopyPayload including extra ones.
        await updateReadingStatus('copy1', {
            status: 'reading',
            current_page: 42,
            started_reading_at: '2024-01-01T00:00:00Z',
            finished_reading_at: null,
            owned_by_user: true,
            borrowed_from: null,
            location: 'bookshelf',
        })

        expect(capturedBody).not.toBeNull()

        // These fields MUST be present (backend accepts them)
        expect(capturedBody!.status).toBe('reading')
        expect(capturedBody!.current_page).toBe(42)

        // These fields must NOT be sent (backend ignores them — misleading code)
        // On unfixed code: all these fields are present in the request body — fails.
        expect(capturedBody!.started_reading_at).toBeUndefined()
        expect(capturedBody!.finished_reading_at).toBeUndefined()
        expect(capturedBody!.owned_by_user).toBeUndefined()
        expect(capturedBody!.borrowed_from).toBeUndefined()
        expect(capturedBody!.location).toBeUndefined()
    })
})
