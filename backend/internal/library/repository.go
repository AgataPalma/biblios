package library

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AgataPalma/biblios/internal/books"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

const libraryColumns = `id, owner_id, name, description, is_cooperative, visibility, deleted_at, created_at, updated_at`

func scanLibrary(row pgx.Row, l *Library) error {
	return row.Scan(
		&l.ID, &l.OwnerID, &l.Name, &l.Description,
		&l.IsCooperative, &l.Visibility,
		&l.DeletedAt, &l.CreatedAt, &l.UpdatedAt,
	)
}

// CreateLibrary creates a library and inserts the owner as a member with all permissions.
func (r *Repository) CreateLibrary(ctx context.Context, ownerID, name string, description *string, isCooperative bool, visibility string) (Library, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Library{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var lib Library
	err = scanLibrary(tx.QueryRow(ctx, `
		INSERT INTO libraries (owner_id, name, description, is_cooperative, visibility)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+libraryColumns,
		ownerID, name, description, isCooperative, visibility), &lib)
	if err != nil {
		return Library{}, fmt.Errorf("insert library: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO library_members
			(library_id, user_id, is_owner, can_view, can_add, can_remove, can_edit, can_invite, can_manage_members)
		VALUES ($1, $2, true, true, true, true, true, true, true)`,
		lib.ID, ownerID)
	if err != nil {
		return Library{}, fmt.Errorf("insert owner member: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return Library{}, fmt.Errorf("commit tx: %w", err)
	}
	return lib, nil
}

func (r *Repository) FindLibraryByID(ctx context.Context, id string) (*Library, error) {
	var lib Library
	err := scanLibrary(r.db.QueryRow(ctx, `
		SELECT `+libraryColumns+` FROM libraries WHERE id=$1 AND deleted_at IS NULL`, id), &lib)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find library: %w", err)
	}
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM library_members lm
		JOIN users u ON u.id=lm.user_id AND u.deleted_at IS NULL
		WHERE lm.library_id=$1`, id).Scan(&lib.MemberCount); err != nil {
		return nil, fmt.Errorf("count library members: %w", err)
	}
	return &lib, nil
}

func (r *Repository) ListUserLibraries(ctx context.Context, userID string) ([]Library, error) {
	rows, err := r.db.Query(ctx, `
			SELECT l.id, l.owner_id, l.name, l.description, l.is_cooperative, l.visibility,
			       l.deleted_at, l.created_at, l.updated_at,
			       COUNT(all_members.user_id) FILTER (WHERE all_member_users.deleted_at IS NULL)
			FROM libraries l
			JOIN library_members lm ON lm.library_id=l.id
			LEFT JOIN library_members all_members ON all_members.library_id=l.id
			LEFT JOIN users all_member_users ON all_member_users.id=all_members.user_id
			WHERE lm.user_id=$1 AND lm.can_view AND l.deleted_at IS NULL
			GROUP BY l.id
			ORDER BY l.created_at ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user libraries: %w", err)
	}
	defer rows.Close()

	var libs []Library
	for rows.Next() {
		var lib Library
		if err := rows.Scan(
			&lib.ID, &lib.OwnerID, &lib.Name, &lib.Description, &lib.IsCooperative, &lib.Visibility,
			&lib.DeletedAt, &lib.CreatedAt, &lib.UpdatedAt, &lib.MemberCount,
		); err != nil {
			return nil, err
		}
		libs = append(libs, lib)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return libs, nil
}

func (r *Repository) ListPublicLibraries(ctx context.Context, page, limit int) ([]Library, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM libraries WHERE visibility='public' AND deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count public libraries: %w", err)
	}

	rows, err := r.db.Query(ctx, `
			SELECT l.id, l.owner_id, l.name, l.description, l.is_cooperative, l.visibility,
			       l.deleted_at, l.created_at, l.updated_at,
			       COUNT(lm.user_id) FILTER (WHERE member_users.deleted_at IS NULL)
			FROM libraries l
			LEFT JOIN library_members lm ON lm.library_id=l.id
			LEFT JOIN users member_users ON member_users.id=lm.user_id
			WHERE l.visibility='public' AND l.deleted_at IS NULL
			GROUP BY l.id
			ORDER BY l.name ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list public libraries: %w", err)
	}
	defer rows.Close()

	var libs []Library
	for rows.Next() {
		var lib Library
		if err := rows.Scan(
			&lib.ID, &lib.OwnerID, &lib.Name, &lib.Description, &lib.IsCooperative, &lib.Visibility,
			&lib.DeletedAt, &lib.CreatedAt, &lib.UpdatedAt, &lib.MemberCount,
		); err != nil {
			return nil, 0, err
		}
		libs = append(libs, lib)
	}
	return libs, total, rows.Err()
}

func (r *Repository) UpdateLibrary(ctx context.Context, id string, name, description, visibility *string) (Library, error) {
	var lib Library
	err := scanLibrary(r.db.QueryRow(ctx, `
		UPDATE libraries SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			visibility  = COALESCE($4, visibility),
			updated_at  = NOW()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING `+libraryColumns, id, name, description, visibility), &lib)
	if err == pgx.ErrNoRows {
		return Library{}, fmt.Errorf("library not found")
	}
	if err != nil {
		return Library{}, fmt.Errorf("update library: %w", err)
	}
	return lib, nil
}

func (r *Repository) DeleteLibrary(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `UPDATE libraries SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete library: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("library not found")
	}
	return nil
}

// ─── Members ──────────────────────────────────────────────────────────────────

const memberColumns = `lm.library_id, lm.user_id, u.username, lm.joined_at, lm.is_owner,
	lm.can_view, lm.can_add, lm.can_remove, lm.can_edit, lm.can_invite, lm.can_manage_members`

func scanMember(row pgx.Row, m *LibraryMember) error {
	return row.Scan(
		&m.LibraryID, &m.UserID, &m.Username, &m.JoinedAt, &m.IsOwner,
		&m.CanView, &m.CanAdd, &m.CanRemove, &m.CanEdit, &m.CanInvite, &m.CanManageMembers,
	)
}

func (r *Repository) GetMember(ctx context.Context, libraryID, userID string) (*LibraryMember, error) {
	var m LibraryMember
	err := scanMember(r.db.QueryRow(ctx, `
		SELECT `+memberColumns+`
		FROM library_members lm
		JOIN users u ON u.id=lm.user_id
			WHERE lm.library_id=$1 AND lm.user_id=$2 AND u.deleted_at IS NULL`, libraryID, userID), &m)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}
	return &m, nil
}

func (r *Repository) ListMembers(ctx context.Context, libraryID string) ([]LibraryMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+memberColumns+`
		FROM library_members lm
		JOIN users u ON u.id=lm.user_id
		WHERE lm.library_id=$1 AND u.deleted_at IS NULL
		ORDER BY lm.is_owner DESC, lm.joined_at ASC`, libraryID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []LibraryMember
	for rows.Next() {
		var m LibraryMember
		if err := scanMember(rows, &m); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) AddMember(ctx context.Context, libraryID, userID string, perms LibraryMember) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO library_members
			(library_id, user_id, is_owner, can_view, can_add, can_remove, can_edit, can_invite, can_manage_members)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (library_id, user_id) DO NOTHING`,
		libraryID, userID, false,
		perms.CanView, perms.CanAdd, perms.CanRemove, perms.CanEdit, perms.CanInvite, perms.CanManageMembers)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *Repository) UpdateMemberPermissions(ctx context.Context, libraryID, userID string, perms LibraryMember) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE library_members SET
			can_view          = $3,
			can_add           = $4,
			can_remove        = $5,
			can_edit          = $6,
			can_invite        = $7,
			can_manage_members = $8
		WHERE library_id=$1 AND user_id=$2 AND is_owner=false`,
		libraryID, userID,
		perms.CanView, perms.CanAdd, perms.CanRemove, perms.CanEdit, perms.CanInvite, perms.CanManageMembers)
	if err != nil {
		return fmt.Errorf("update member permissions: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("member not found or is owner")
	}
	return nil
}

func (r *Repository) RemoveMember(ctx context.Context, libraryID, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin remove member: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		DELETE FROM library_book_copies lbc
		USING book_copies bc
		WHERE lbc.book_copy_id=bc.id
		  AND lbc.library_id=$1
		  AND bc.owner_id=$2`, libraryID, userID); err != nil {
		return fmt.Errorf("remove member books: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		DELETE FROM library_members
		WHERE library_id=$1 AND user_id=$2 AND is_owner=false`, libraryID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("member not found or is owner")
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit remove member: %w", err)
	}
	return nil
}

// ─── Invitations ──────────────────────────────────────────────────────────────

const invitationColumns = `id, library_id, invited_by, invited_user_id, invited_email,
	token, status, accepted_at, expires_at, created_at`

func scanInvitation(row pgx.Row, inv *LibraryInvitation) error {
	return row.Scan(
		&inv.ID, &inv.LibraryID, &inv.InvitedBy, &inv.InvitedUserID, &inv.InvitedEmail,
		&inv.Token, &inv.Status, &inv.AcceptedAt, &inv.ExpiresAt, &inv.CreatedAt,
	)
}

func (r *Repository) CreateInvitation(ctx context.Context, libraryID, invitedBy, invitedEmail string, expiresAt time.Time) (LibraryInvitation, error) {
	token, err := generateToken()
	if err != nil {
		return LibraryInvitation{}, err
	}

	// Try to resolve invited_user_id from email
	var invitedUserID *string
	var uid string
	if err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE email=$1 AND deleted_at IS NULL`, invitedEmail).Scan(&uid); err == nil {
		invitedUserID = &uid
	}

	var inv LibraryInvitation
	err = scanInvitation(r.db.QueryRow(ctx, `
		INSERT INTO library_invitations
			(library_id, invited_by, invited_user_id, invited_email, token, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING `+invitationColumns,
		libraryID, invitedBy, invitedUserID, invitedEmail, token, expiresAt), &inv)
	if err != nil {
		return LibraryInvitation{}, fmt.Errorf("create invitation: %w", err)
	}
	return inv, nil
}

func (r *Repository) FindInvitationByToken(ctx context.Context, token string) (*LibraryInvitation, error) {
	var inv LibraryInvitation
	err := scanInvitation(r.db.QueryRow(ctx, `
		SELECT `+invitationColumns+` FROM library_invitations WHERE token=$1`, token), &inv)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find invitation by token: %w", err)
	}
	return &inv, nil
}

func (r *Repository) FindInvitationByID(ctx context.Context, id string) (*LibraryInvitation, error) {
	var inv LibraryInvitation
	err := scanInvitation(r.db.QueryRow(ctx, `
		SELECT `+invitationColumns+` FROM library_invitations WHERE id=$1`, id), &inv)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find invitation by id: %w", err)
	}
	return &inv, nil
}

// AcceptInvitation marks the invitation accepted and adds the user as a library member.
func (r *Repository) AcceptInvitation(ctx context.Context, token, userID string) error {
	inv, err := r.FindInvitationByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv == nil {
		return fmt.Errorf("invitation not found")
	}
	if inv.Status != "pending" {
		return fmt.Errorf("invitation is no longer pending")
	}
	if time.Now().After(inv.ExpiresAt) {
		return fmt.Errorf("invitation has expired")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `
		UPDATE library_invitations
		SET status='accepted', accepted_at=NOW(), invited_user_id=$2
		WHERE token=$1`, token, userID)
	if err != nil {
		return fmt.Errorf("accept invitation: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO library_members
			(library_id, user_id, is_owner, can_view, can_add, can_remove, can_edit, can_invite, can_manage_members)
		VALUES ($1,$2,false,true,false,false,false,false,false)
		ON CONFLICT (library_id, user_id) DO NOTHING`,
		inv.LibraryID, userID)
	if err != nil {
		return fmt.Errorf("add member on accept: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *Repository) DeclineInvitation(ctx context.Context, token string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE library_invitations SET status='declined' WHERE token=$1 AND status='pending'`, token)
	if err != nil {
		return fmt.Errorf("decline invitation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("invitation not found or already actioned")
	}
	return nil
}

// ListUserInvitations returns pending invitations for the given user's email.
func (r *Repository) ListUserInvitations(ctx context.Context, userID string) ([]LibraryInvitation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+invitationColumns+`
		FROM library_invitations li
		WHERE (li.invited_user_id=$1 OR li.invited_email=(SELECT email FROM users WHERE id=$1 AND deleted_at IS NULL))
		  AND li.status='pending'
		ORDER BY li.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user invitations: %w", err)
	}
	defer rows.Close()

	var invs []LibraryInvitation
	for rows.Next() {
		var inv LibraryInvitation
		if err := scanInvitation(rows, &inv); err != nil {
			return nil, err
		}
		// Strip token from response for security
		inv.Token = ""
		invs = append(invs, inv)
	}
	return invs, rows.Err()
}

// ─── Library books ────────────────────────────────────────────────────────────

func (r *Repository) AddBookCopyToLibrary(ctx context.Context, libraryID, userID, copyID string) error {
	tag, err := r.db.Exec(ctx, `
			INSERT INTO library_book_copies (library_id, book_copy_id)
			SELECT l.id, bc.id
			FROM libraries l
			JOIN library_members lm ON lm.library_id=l.id AND lm.user_id=$2
			JOIN book_copies bc ON bc.id=$3
			WHERE l.id=$1
			  AND l.deleted_at IS NULL
			  AND lm.can_add
			  AND bc.owner_id=$2
			  AND bc.deleted_at IS NULL
			ON CONFLICT (library_id, book_copy_id) DO UPDATE
			SET book_copy_id=EXCLUDED.book_copy_id`, libraryID, userID, copyID)
	if err != nil {
		return fmt.Errorf("add book to library: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCopyNotOwned
	}
	return nil
}

func (r *Repository) RemoveBookCopyFromLibrary(ctx context.Context, libraryID, userID, copyID string) error {
	tag, err := r.db.Exec(ctx, `
			DELETE FROM library_book_copies lbc
			USING libraries l, library_members lm
			WHERE lbc.library_id=l.id
			  AND lm.library_id=l.id
			  AND l.id=$1
			  AND lm.user_id=$2
			  AND lm.can_remove
			  AND l.deleted_at IS NULL
			  AND lbc.book_copy_id=$3`, libraryID, userID, copyID)
	if err != nil {
		return fmt.Errorf("remove book from library: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBookNotInLibrary
	}
	return nil
}

func (r *Repository) ListLibraryBooks(ctx context.Context, libraryID string, page, limit int) ([]LibraryBook, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM library_book_copies lbc
			JOIN book_copies bc ON bc.id=lbc.book_copy_id
			JOIN book_editions be ON be.id=bc.edition_id
			JOIN books b ON b.id=be.book_id
			WHERE lbc.library_id=$1
			  AND bc.deleted_at IS NULL
			  AND be.deleted_at IS NULL AND be.status='approved'
			  AND b.deleted_at IS NULL AND b.status='approved'`, libraryID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count library books: %w", err)
	}

	rows, err := r.db.Query(ctx, `
			SELECT lbc.library_id, lbc.book_copy_id, b.id, be.id, b.title,
			       ARRAY(
			           SELECT c.name
			           FROM book_contributors bco
			           JOIN contributors c ON c.id=bco.contributor_id
			           WHERE bco.book_id=b.id
			             AND bco.role IN ('author', 'co_author')
			             AND c.deleted_at IS NULL
			             AND c.status='approved'
			           ORDER BY c.name
			       ),
			       be.format, be.language, be.cover_url, lbc.added_at
			FROM library_book_copies lbc
			JOIN book_copies bc ON bc.id=lbc.book_copy_id
			JOIN book_editions be ON be.id=bc.edition_id
			JOIN books b ON b.id=be.book_id
			WHERE lbc.library_id=$1
			  AND bc.deleted_at IS NULL
			  AND be.deleted_at IS NULL AND be.status='approved'
			  AND b.deleted_at IS NULL AND b.status='approved'
			ORDER BY lbc.added_at DESC
			LIMIT $2 OFFSET $3`, libraryID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list library books: %w", err)
	}
	defer rows.Close()

	var libraryBooks []LibraryBook
	for rows.Next() {
		var book LibraryBook
		if err := rows.Scan(
			&book.LibraryID, &book.CopyID, &book.BookID, &book.EditionID, &book.Title,
			&book.Authors, &book.Format, &book.Language, &book.CoverURL, &book.AddedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan library book: %w", err)
		}
		libraryBooks = append(libraryBooks, book)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return libraryBooks, total, nil
}

func (r *Repository) ListUserLibraryBooks(ctx context.Context, libraryID, userID string, page, limit int) ([]books.UserBook, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM library_book_copies lbc
		JOIN book_copies bc ON bc.id=lbc.book_copy_id
		WHERE lbc.library_id=$1 AND bc.owner_id=$2 AND bc.deleted_at IS NULL`, libraryID, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user library books: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT bc.id, bc.reading_status, bc.current_page, bc.started_reading_at, bc.finished_reading_at,
		       bc.owned_by_user, bc.borrowed_from, bc.location, bc.condition, bc.reread_count,
		       bc.personal_notes, lbc.added_at,
		       be.id, be.format, be.language, be.cover_url,
		       b.id, b.title, b.description, b.series_id, b.series_position, b.status,
		       b.deleted_at, b.created_at, b.updated_at
		FROM library_book_copies lbc
		JOIN book_copies bc ON bc.id=lbc.book_copy_id
		JOIN book_editions be ON be.id=bc.edition_id
		JOIN books b ON b.id=be.book_id
		WHERE lbc.library_id=$1
		  AND bc.owner_id=$2
		  AND bc.deleted_at IS NULL
		  AND be.deleted_at IS NULL
		  AND b.deleted_at IS NULL
		ORDER BY lbc.added_at DESC
		LIMIT $3 OFFSET $4`, libraryID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list user library books: %w", err)
	}
	defer rows.Close()

	var userBooks []books.UserBook
	for rows.Next() {
		var ub books.UserBook
		if err := rows.Scan(
			&ub.CopyID, &ub.ReadingStatus, &ub.CurrentPage, &ub.StartedReadingAt, &ub.FinishedReadingAt,
			&ub.OwnedByUser, &ub.BorrowedFrom, &ub.Location, &ub.Condition, &ub.RereadCount,
			&ub.PersonalNotes, &ub.AddedAt,
			&ub.EditionID, &ub.Format, &ub.Language, &ub.CoverURL,
			&ub.Book.ID, &ub.Book.Title, &ub.Book.Description, &ub.Book.SeriesID, &ub.Book.SeriesPosition,
			&ub.Book.Status, &ub.Book.DeletedAt, &ub.Book.CreatedAt, &ub.Book.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user library book: %w", err)
		}
		ub.Book.Authors = []books.Contributor{}
		ub.Book.Genres = []books.Genre{}
		userBooks = append(userBooks, ub)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Batch-load authors for each book
	if len(userBooks) > 0 {
		bookIDs := make([]string, len(userBooks))
		idx := map[string][]int{}
		for i, ub := range userBooks {
			bookIDs[i] = ub.Book.ID
			idx[ub.Book.ID] = append(idx[ub.Book.ID], i)
		}
		aRows, err := r.db.Query(ctx, `
			SELECT bc2.book_id, c.id, c.name, c.status, c.deleted_at, c.created_at, c.updated_at
			FROM contributors c JOIN book_contributors bc2 ON bc2.contributor_id=c.id
			WHERE bc2.book_id=ANY($1) AND bc2.role IN ('author','co_author') AND c.deleted_at IS NULL`, bookIDs)
		if err == nil {
			for aRows.Next() {
				var bookID string
				var c books.Contributor
				if err := aRows.Scan(&bookID, &c.ID, &c.Name, &c.Status, &c.DeletedAt, &c.CreatedAt, &c.UpdatedAt); err == nil {
					for _, i := range idx[bookID] {
						userBooks[i].Book.Authors = append(userBooks[i].Book.Authors, c)
					}
				}
			}
			aRows.Close()
		}
	}

	return userBooks, total, nil
}
