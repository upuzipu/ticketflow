package postgres_test

import (
	"github.com/upuzipu/ticketflow/internal/repository/postgres"
	"github.com/upuzipu/ticketflow/internal/service"
)

// Compile-time check: the adapter must satisfy the port exactly —
// names, signatures, everything.
var _ service.UserRepository = (*postgres.UserRepository)(nil)
var _ service.Inventory = (*postgres.Inventory)(nil)
var _ service.HoldRepository = (*postgres.HoldRepository)(nil)
