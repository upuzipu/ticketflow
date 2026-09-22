# PostgreSQL connection string
PG_DSN=postgres://user:password@localhost:5432/ticketflow?sslmode=disable
# HMAC secret for JWT signing, min 32 bytes
JWT_SECRET=change-me-at-least-32-bytes-long
# HTTP listen address
HTTP_ADDR=:8080