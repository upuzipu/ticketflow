// Package domain contains entities from the TicketFlow domain.
//
// Package rules (violation = rework):
// 1. No imports from other project packages.
// 2. No external libraries allowed—only the Go standard library.
// 3. No database access, network access, or JSON tags for HTTP (mapping is in other layers).
// 4. All IDs are strings. Uuid generation is not here, at the application boundaries.
package domain
