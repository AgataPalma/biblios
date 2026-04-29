package library

import "time"

type Library struct {
	ID            string     `json:"id"`
	OwnerID       string     `json:"owner_id"`
	Name          string     `json:"name"`
	Description   *string    `json:"description,omitempty"`
	IsCooperative bool       `json:"is_cooperative"`
	Visibility    string     `json:"visibility"` // private | semi_public | public
	MemberCount   int        `json:"member_count,omitempty"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type LibraryMember struct {
	LibraryID        string    `json:"library_id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username,omitempty"`
	JoinedAt         time.Time `json:"joined_at"`
	IsOwner          bool      `json:"is_owner"`
	CanView          bool      `json:"can_view"`
	CanAdd           bool      `json:"can_add"`
	CanRemove        bool      `json:"can_remove"`
	CanEdit          bool      `json:"can_edit"`
	CanInvite        bool      `json:"can_invite"`
	CanManageMembers bool      `json:"can_manage_members"`
}

type LibraryInvitation struct {
	ID            string     `json:"id"`
	LibraryID     string     `json:"library_id"`
	InvitedBy     string     `json:"invited_by"`
	InvitedUserID *string    `json:"invited_user_id,omitempty"`
	InvitedEmail  string     `json:"invited_email"`
	Token         string     `json:"token,omitempty"`
	Status        string     `json:"status"`
	AcceptedAt    *time.Time `json:"accepted_at,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
}
