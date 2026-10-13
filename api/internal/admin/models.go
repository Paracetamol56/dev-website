package admin

import (
	"dev/internal/audit"
	"dev/internal/user"
)

const (
	DefaultLimit = 25
	MaxLimit     = 100
)

type PageQuery struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (query *PageQuery) normalize() {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = DefaultLimit
	}
}

type UsersQuery struct {
	PageQuery
	Search string `form:"search" binding:"max=100"`
}

type AuditQuery struct {
	PageQuery
	UserId string `form:"userId" binding:"omitempty,len=24,hexadecimal"`
	Action string `form:"action" binding:"max=50"`
}

type RoleBody struct {
	Role string `json:"role" binding:"omitempty,oneof=admin"`
}

type UsersPage struct {
	Users []user.User `json:"users"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

type UserSummary struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AuditPage names the users its entries mention, by id
type AuditPage struct {
	Entries []audit.Entry          `json:"entries"`
	Users   map[string]UserSummary `json:"users"`
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
}
