//nolint:dupl
package vt

import (
	"time"

	"apisrv/pkg/db"
)

type User struct {
	ID             int        `json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	Login          string     `json:"login" validate:"required,max=64"`
	Password       string     `json:"password" validate:"max=64"`
	LastActivityAt *time.Time `json:"lastActivityAt"`
	StatusID       int        `json:"statusId" validate:"required,status"`

	Status *Status `json:"status"`
}

func (u *User) ToDB() *db.User {
	if u == nil {
		return nil
	}

	user := &db.User{
		ID:             u.ID,
		Login:          u.Login,
		LastActivityAt: u.LastActivityAt,
		StatusID:       u.StatusID,
	}

	return user
}

type UserSearch struct {
	ID                 *int       `json:"id"`
	Login              *string    `json:"login" validate:"max=64"`
	StatusID           *int       `json:"statusId" validate:"status"`
	LastActivityAtFrom *time.Time `json:"lastActivityAtFrom"`
	LastActivityAtTo   *time.Time `json:"lastActivityAtTo"`
	IDs                []int      `json:"ids"`
	NotID              *int       `json:"notId"`
}

func (us *UserSearch) ToDB() *db.UserSearch {
	if us == nil {
		return nil
	}

	return &db.UserSearch{
		ID:                 us.ID,
		LoginILike:         us.Login,
		StatusID:           us.StatusID,
		LastActivityAtFrom: us.LastActivityAtFrom,
		LastActivityAtTo:   us.LastActivityAtTo,
		IDs:                us.IDs,
		NotID:              us.NotID,
	}
}

type UserSummary struct {
	ID             int        `json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	Login          string     `json:"login"`
	LastActivityAt *time.Time `json:"lastActivityAt"`

	Status *Status `json:"status"`
}

type UserProfile struct {
	ID             int        `json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	Login          string     `json:"login"`
	LastActivityAt *time.Time `json:"lastActivityAt"`
	StatusID       int        `json:"statusId"`
}
