package vt

import (
	"apisrv/pkg/db"
)

func NewUser(in *db.User) *User {
	if in == nil {
		return nil
	}

	user := &User{
		ID:             in.ID,
		CreatedAt:      in.CreatedAt,
		Login:          in.Login,
		LastActivityAt: in.LastActivityAt,
		StatusID:       in.StatusID,
		Status:         NewStatus(in.StatusID),
	}

	return user
}

func NewUserSummary(in *db.User) *UserSummary {
	if in == nil {
		return nil
	}

	return &UserSummary{
		ID:             in.ID,
		CreatedAt:      in.CreatedAt,
		Login:          in.Login,
		LastActivityAt: in.LastActivityAt,
		Status:         NewStatus(in.StatusID),
	}
}

func NewUserProfile(in *db.User) *UserProfile {
	if in == nil {
		return nil
	}

	return &UserProfile{
		ID:             in.ID,
		CreatedAt:      in.CreatedAt,
		Login:          in.Login,
		LastActivityAt: in.LastActivityAt,
		StatusID:       in.StatusID,
	}
}
