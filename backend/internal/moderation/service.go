package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AgataPalma/biblios/internal/books"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ModerationLog represents a single audit entry.
type ModerationLog struct {
	ID          string          `json:"id"`
	ModeratorID string          `json:"moderator_id"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Action      string          `json:"action"`
	Before      json.RawMessage `json:"before,omitempty"`
	After       json.RawMessage `json:"after,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Service struct {
	repo *books.Repository
	db   *pgxpool.Pool
}

func NewService(repo *books.Repository, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, db: db}
}

// ListPending returns all pending submissions with pagination.
func (s *Service) ListPending(ctx context.Context, page, limit int) ([]books.Submission, int, error) {
	return s.repo.ListPendingSubmissions(ctx, page, limit)
}

// GetSubmission returns a single submission with full book/edition details.
func (s *Service) GetSubmission(ctx context.Context, id string) (*books.Submission, error) {
	sub, err := s.repo.FindSubmissionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("submission not found")
	}
	return sub, nil
}

// Approve approves a submission:
//  1. Sets book/edition/contributors/genres to approved.
//  2. If the submission has no copy yet (catalogue_only=false), creates one.
//  3. Marks the submission approved.
//  4. Writes a moderation log entry.
func (s *Service) Approve(ctx context.Context, submissionID, moderatorID string) error {
	sub, err := s.repo.FindSubmissionByID(ctx, submissionID)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("submission not found")
	}
	if sub.Status != "pending" {
		return fmt.Errorf("submission is not pending")
	}

	// Snapshot before state for audit log
	beforeJSON, _ := json.Marshal(sub)

	// Approve book/edition entities
	if sub.BookID != nil && sub.EditionID != nil {
		if err := s.repo.ApproveBookEntities(ctx, *sub.BookID, *sub.EditionID); err != nil {
			return fmt.Errorf("approve entities: %w", err)
		}
	}

	// If no copy exists yet and not catalogue_only, create one now
	if !sub.CatalogueOnly && sub.CopyID == nil && sub.EditionID != nil {
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		txRepo := s.repo.WithDB(tx)
		copy, err := txRepo.InsertCopy(ctx, *sub.EditionID, sub.SubmittedBy, nil, books.CopyOptions{})
		if err != nil {
			return fmt.Errorf("create copy on approve: %w", err)
		}

		// Update submission with copy ID and approved status
		if _, err := tx.Exec(ctx, `
			UPDATE submissions SET status='approved', reviewed_by=$2, reviewed_at=NOW(),
			       copy_id=$3, updated_at=NOW()
			WHERE id=$1`, submissionID, moderatorID, copy.ID); err != nil {
			return fmt.Errorf("update submission: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}
	} else {
		if err := s.repo.ApproveSubmission(ctx, submissionID, moderatorID); err != nil {
			return err
		}
	}

	// Reload for after snapshot
	after, _ := s.repo.FindSubmissionByID(ctx, submissionID)
	afterJSON, _ := json.Marshal(after)

	entityID := submissionID
	if sub.BookID != nil {
		entityID = *sub.BookID
	}
	_ = s.repo.InsertModerationLog(ctx, moderatorID, "submission", entityID, "approved", beforeJSON, afterJSON)

	return nil
}

// Reject rejects a submission, soft-deletes the copy if one exists, and logs the action.
func (s *Service) Reject(ctx context.Context, submissionID, moderatorID, reason string) error {
	sub, err := s.repo.FindSubmissionByID(ctx, submissionID)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("submission not found")
	}
	if sub.Status != "pending" {
		return fmt.Errorf("submission is not pending")
	}
	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}

	beforeJSON, _ := json.Marshal(sub)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Soft-delete the copy if it exists
	if sub.CopyID != nil {
		if _, err := tx.Exec(ctx, `UPDATE book_copies SET deleted_at=NOW() WHERE id=$1`, *sub.CopyID); err != nil {
			return fmt.Errorf("soft delete copy: %w", err)
		}
	}

	// Reject the submission
	if _, err := tx.Exec(ctx, `
		UPDATE submissions SET status='rejected', reviewed_by=$2, reviewed_at=NOW(),
		       rejection_reason=$3, updated_at=NOW()
		WHERE id=$1`, submissionID, moderatorID, reason); err != nil {
		return fmt.Errorf("reject submission: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	after, _ := s.repo.FindSubmissionByID(ctx, submissionID)
	afterJSON, _ := json.Marshal(after)

	entityID := submissionID
	if sub.BookID != nil {
		entityID = *sub.BookID
	}
	_ = s.repo.InsertModerationLog(ctx, moderatorID, "submission", entityID, "rejected", beforeJSON, afterJSON)

	return nil
}

// EditAndApprove updates book/edition fields then approves the submission.
func (s *Service) EditAndApprove(ctx context.Context, submissionID, moderatorID string, bookInput *books.UpdateBookInput) error {
	sub, err := s.repo.FindSubmissionByID(ctx, submissionID)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("submission not found")
	}
	if sub.Status != "pending" {
		return fmt.Errorf("submission is not pending")
	}

	beforeJSON, _ := json.Marshal(sub)

	// Apply edits if provided
	if bookInput != nil && sub.BookID != nil {
		bookInput.ID = *sub.BookID
		if sub.EditionID != nil {
			bookInput.EditionID = *sub.EditionID
		}
		svc := books.NewService(s.repo, s.db)
		if err := svc.UpdateBook(ctx, *bookInput); err != nil {
			return fmt.Errorf("edit book: %w", err)
		}
	}

	// Now approve
	if err := s.Approve(ctx, submissionID, moderatorID); err != nil {
		return err
	}

	after, _ := s.repo.FindSubmissionByID(ctx, submissionID)
	afterJSON, _ := json.Marshal(after)

	entityID := submissionID
	if sub.BookID != nil {
		entityID = *sub.BookID
	}
	_ = s.repo.InsertModerationLog(ctx, moderatorID, "submission", entityID, "edited", beforeJSON, afterJSON)

	return nil
}

// ListLogs returns moderation log entries with pagination.
func (s *Service) ListLogs(ctx context.Context, page, limit int) ([]ModerationLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM moderation_log`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count logs: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, moderator_id, entity_type, entity_id, action, before, after, created_at
		FROM moderation_log
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list logs: %w", err)
	}
	defer rows.Close()

	var logs []ModerationLog
	for rows.Next() {
		var l ModerationLog
		if err := rows.Scan(&l.ID, &l.ModeratorID, &l.EntityType, &l.EntityID,
			&l.Action, &l.Before, &l.After, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}
