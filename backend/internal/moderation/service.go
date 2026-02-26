package moderation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AgataPalma/biblios/internal/books"
)

type Service struct {
	repo *books.Repository
}

func NewService(repo *books.Repository) *Service {
	return &Service{repo: repo}
}

type ListSubmissionsResult struct {
	Submissions []books.Submission `json:"submissions"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	Limit       int                `json:"limit"`
}

func (s *Service) ListPending(ctx context.Context, page int, limit int) (ListSubmissionsResult, error) {
	var submissions []books.Submission
	var total int
	var err error

	submissions, total, err = s.repo.ListPendingSubmissions(ctx, page, limit)
	if err != nil {
		return ListSubmissionsResult{}, err
	}
	if submissions == nil {
		submissions = []books.Submission{}
	}

	return ListSubmissionsResult{
		Submissions: submissions,
		Total:       total,
		Page:        page,
		Limit:       limit,
	}, nil
}

func (s *Service) GetSubmission(ctx context.Context, id string) (*books.Submission, error) {
	return s.repo.FindSubmissionByID(ctx, id)
}

type ApproveInput struct {
	SubmissionID string
	ReviewerID   string
}

func (s *Service) Approve(ctx context.Context, input ApproveInput) error {
	var submission *books.Submission
	var err error

	submission, err = s.repo.FindSubmissionByID(ctx, input.SubmissionID)
	if err != nil {
		return fmt.Errorf("submission not found: %w", err)
	}

	// Approve all entities
	if submission.BookID != nil && submission.EditionID != nil {
		err = s.repo.ApproveBookEntities(ctx, *submission.BookID, *submission.EditionID)
		if err != nil {
			return err
		}

		// Create the book copy now that everything is approved
		var copy books.Copy
		copy, err = s.repo.InsertCopyDirect(ctx, *submission.EditionID, submission.SubmittedBy)
		if err != nil {
			return err
		}

		// Update submission with copy ID
		err = s.repo.ApproveSubmissionWithCopy(ctx, input.SubmissionID, input.ReviewerID, copy.ID)
		if err != nil {
			return err
		}
	} else {
		err = s.repo.ApproveSubmission(ctx, input.SubmissionID, input.ReviewerID)
		if err != nil {
			return err
		}
	}

	// Log it
	s.repo.InsertModerationLog(ctx, input.ReviewerID, "submission", input.SubmissionID, "approved", nil, nil)

	return nil
}

type RejectInput struct {
	SubmissionID string
	ReviewerID   string
	Reason       string
}

func (s *Service) Reject(ctx context.Context, input RejectInput) error {
	var err error = s.repo.RejectSubmission(ctx, input.SubmissionID, input.ReviewerID, input.Reason)
	if err != nil {
		return err
	}

	s.repo.InsertModerationLog(ctx, input.ReviewerID, "submission", input.SubmissionID, "rejected", nil, nil)

	return nil
}

type EditAndApproveInput struct {
	SubmissionID string
	ReviewerID   string
	Title        string
	Description  *string
	CoverURL     *string
	Edition      books.Edition
}

func (s *Service) EditAndApprove(ctx context.Context, input EditAndApproveInput) error {
	var submission *books.Submission
	var err error

	submission, err = s.repo.FindSubmissionByID(ctx, input.SubmissionID)
	if err != nil {
		return fmt.Errorf("submission not found: %w", err)
	}

	if submission.BookID == nil || submission.EditionID == nil {
		return fmt.Errorf("submission has no book or edition")
	}

	// Snapshot before state for audit log
	var before []byte
	before, _ = json.Marshal(submission)

	// Apply edits
	err = s.repo.UpdateBookDetails(ctx, *submission.BookID, input.Title, input.Description, input.CoverURL)
	if err != nil {
		return err
	}

	err = s.repo.UpdateEditionDetails(ctx, *submission.EditionID, input.Edition)
	if err != nil {
		return err
	}

	// Approve everything
	err = s.repo.ApproveBookEntities(ctx, *submission.BookID, *submission.EditionID)
	if err != nil {
		return err
	}

	// Create copy
	var copy books.Copy
	copy, err = s.repo.InsertCopyDirect(ctx, *submission.EditionID, submission.SubmittedBy)
	if err != nil {
		return err
	}

	err = s.repo.ApproveSubmissionWithCopy(ctx, input.SubmissionID, input.ReviewerID, copy.ID)
	if err != nil {
		return err
	}

	// Log with before/after
	var after []byte
	after, _ = json.Marshal(input)
	s.repo.InsertModerationLog(ctx, input.ReviewerID, "submission", input.SubmissionID, "edited", before, after)

	return nil
}
