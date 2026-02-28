package http

import (
	"time"

	"go-prod-app/internal/domain"
)

func toUserResponse(u *domain.User) UserResponse {
	var deletedAt *string
	if u.DeletedAt() != nil {
		s := u.DeletedAt().Format(time.RFC3339)
		deletedAt = &s
	}

	return UserResponse{
		ID:        string(u.ID()),
		Name:      u.Name(),
		Email:     u.Email(),
		Version:   u.Version(),
		CreatedAt: u.CreatedAt().Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt().Format(time.RFC3339),
		DeletedAt: deletedAt,
	}
}
