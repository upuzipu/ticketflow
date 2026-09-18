package domain

import "time"

type Role string

const (
	RoleGuest     Role = "guest"
	RoleBuyer     Role = "buyer"
	RoleOrganizer Role = "organizer"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
}

// IsValid reports whether r is one of the defined roles
func (r Role) IsValid() bool {
	switch r {
	case RoleGuest, RoleBuyer, RoleOrganizer, RoleAdmin:
		return true
	default:
		return false
	}
}

// CanCreateEvents reports whether u has a role that allows event creation (organizer or admin).
func (u User) CanCreateEvents() bool {
	switch u.Role {
	case RoleOrganizer, RoleAdmin:
		return true
	default:
		return false
	}
}

// CanManageEvent reports whether u is allowed to manage an event
// created by the given organizer: u either owns it or is an admin.
func (u User) CanManageEvent(organizerID string) bool {
	return u.Role == RoleAdmin || u.ID == organizerID
}
