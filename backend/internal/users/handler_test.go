package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/AgataPalma/biblios/internal/apictx"
)

type fakeUserService struct {
	deleteErr error
	calls     *[]string
	onDelete  func()
}

func (f *fakeUserService) GetByID(context.Context, string) (User, error) {
	return User{}, nil
}

func (f *fakeUserService) UpdateProfile(context.Context, string, *string, *string, *string) (User, error) {
	return User{}, nil
}

func (f *fakeUserService) UpdateEmail(context.Context, string, string, string) error {
	return nil
}

func (f *fakeUserService) UpdatePassword(context.Context, string, string, string) error {
	return nil
}

func (f *fakeUserService) UpdateTheme(context.Context, string, string) error {
	return nil
}

func (f *fakeUserService) DeleteUser(context.Context, string) error {
	*f.calls = append(*f.calls, "delete-account")
	if f.onDelete != nil {
		f.onDelete()
	}
	return f.deleteErr
}

type fakeSessionController struct {
	disableErr error
	enableErr  error
	cleanupErr error
	calls      *[]string
	enableCtx  error
	cleanupCtx error
}

func (f *fakeSessionController) DisableUser(context.Context, string) error {
	*f.calls = append(*f.calls, "disable-sessions")
	return f.disableErr
}

func (f *fakeSessionController) EnableUser(ctx context.Context, _ string) error {
	*f.calls = append(*f.calls, "enable-sessions")
	f.enableCtx = ctx.Err()
	return f.enableErr
}

func (f *fakeSessionController) DeleteAllUserSessions(ctx context.Context, _ string) error {
	*f.calls = append(*f.calls, "cleanup-sessions")
	f.cleanupCtx = ctx.Err()
	return f.cleanupErr
}

func TestDeleteUserDetachesCompensationFromCancelledRequest(t *testing.T) {
	tests := []struct {
		name       string
		deleteErr  error
		wantEnable bool
		wantCalls  []string
	}{
		{
			name:       "rollback re-enables after request cancellation",
			deleteErr:  errors.New("database cancelled"),
			wantEnable: true,
			wantCalls:  []string{"disable-sessions", "delete-account", "enable-sessions"},
		},
		{
			name:      "successful deletion cleans sessions after request cancellation",
			wantCalls: []string{"disable-sessions", "delete-account", "cleanup-sessions"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := []string{}
			requestContext, cancelRequest := context.WithCancel(context.Background())
			service := &fakeUserService{
				deleteErr: test.deleteErr,
				calls:     &calls,
				onDelete:  cancelRequest,
			}
			sessions := &fakeSessionController{calls: &calls}
			handler := NewHandler(service, sessions)

			claims := apictx.Claims{UserID: "user-1"}
			requestContext = context.WithValue(requestContext, apictx.UserClaimsKey, claims)
			request := httptest.NewRequest(http.MethodDelete, "/users/me", nil).WithContext(requestContext)
			response := httptest.NewRecorder()

			handler.DeleteUser(response, request)

			if !reflect.DeepEqual(calls, test.wantCalls) {
				t.Fatalf("DeleteUser() calls = %#v, want %#v", calls, test.wantCalls)
			}
			if test.wantEnable {
				if sessions.enableCtx != nil {
					t.Fatalf("rollback context error = %v, want nil", sessions.enableCtx)
				}
			} else if sessions.cleanupCtx != nil {
				t.Fatalf("cleanup context error = %v, want nil", sessions.cleanupCtx)
			}
		})
	}
}

func TestDeleteUserDisablesSessionsBeforeDatabaseDeletion(t *testing.T) {
	tests := []struct {
		name       string
		disableErr error
		deleteErr  error
		cleanupErr error
		wantStatus int
		wantCalls  []string
	}{
		{
			name:       "successful deletion disables then cleans sessions",
			wantStatus: http.StatusOK,
			wantCalls:  []string{"disable-sessions", "delete-account", "cleanup-sessions"},
		},
		{
			name:       "redis disable failure stops database deletion",
			disableErr: errors.New("redis unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantCalls:  []string{"disable-sessions"},
		},
		{
			name:       "database failure rolls back disabled marker",
			deleteErr:  errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantCalls:  []string{"disable-sessions", "delete-account", "enable-sessions"},
		},
		{
			name:       "session cleanup failure remains safely disabled",
			cleanupErr: errors.New("cleanup unavailable"),
			wantStatus: http.StatusOK,
			wantCalls:  []string{"disable-sessions", "delete-account", "cleanup-sessions"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := []string{}
			service := &fakeUserService{deleteErr: test.deleteErr, calls: &calls}
			sessions := &fakeSessionController{
				disableErr: test.disableErr,
				cleanupErr: test.cleanupErr,
				calls:      &calls,
			}
			handler := NewHandler(service, sessions)

			request := httptest.NewRequest(http.MethodDelete, "/users/me", nil)
			claims := apictx.Claims{UserID: "user-1"}
			request = request.WithContext(context.WithValue(request.Context(), apictx.UserClaimsKey, claims))
			response := httptest.NewRecorder()

			handler.DeleteUser(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("DeleteUser() status = %d, want %d; body=%s", response.Code, test.wantStatus, response.Body.String())
			}
			if !reflect.DeepEqual(calls, test.wantCalls) {
				t.Fatalf("DeleteUser() calls = %#v, want %#v", calls, test.wantCalls)
			}
		})
	}
}
