package vt

import (
	"apisrv/pkg/db"
)

const maxPageSize = 500

type ViewOps struct {
	// page number, default - 1
	Page int `json:"page"`
	// items count per page, max - 500
	PageSize int `json:"pageSize"`
	// sort by column name
	SortColumn string `json:"sortColumn"`
	// descending sort
	SortDesc bool `json:"sortDesc"`
}

func (v *ViewOps) Pager() db.Pager {
	if v == nil {
		return db.PagerDefault
	}

	if v.PageSize > maxPageSize {
		v.PageSize = maxPageSize
	} else if v.PageSize < 1 {
		v.PageSize = 1
	}

	return db.Pager{Page: v.Page, PageSize: v.PageSize}
}

type Status struct {
	ID    int    `json:"id"`
	Alias string `json:"alias" validate:"required,max=32"`
	Title string `json:"title" validate:"required,max=255"`
}

func NewStatus(id int) *Status {
	switch id {
	case db.StatusEnabled:
		return &Status{ID: db.StatusEnabled, Alias: "enabled", Title: "Опубликован"}
	case db.StatusDisabled:
		return &Status{ID: db.StatusDisabled, Alias: "disabled", Title: "Не опубликован"}
	case db.StatusDeleted:
		return &Status{ID: db.StatusDeleted, Alias: "deleted", Title: "Удален"}
	}
	return nil
}

type StatusUpdate struct {
	StatusID  int   `json:"statusId" validate:"required,status"`
	ObjectIDs []int `json:"ids" validate:"required,gt=0"`
}
