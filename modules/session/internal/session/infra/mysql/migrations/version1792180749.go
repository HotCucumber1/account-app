package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"
)

func Version1792180749(client mysql.ClientContext) migrator.Migration {
	return &version1792180749{
		client: client,
	}
}

type version1792180749 struct {
	client mysql.ClientContext
}

func (v *version1792180749) Version() int64 {
	return 1792180749
}

func (v *version1792180749) Description() string {
	return "Create session table"
}

func (v *version1792180749) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-session-table", `
		CREATE TABLE session (
		    id BINARY(16) NOT NULL,
		    device_id BINARY(16) NOT NULL,
		    ip_address VARCHAR(45) NOT NULL,
		    created_at DATETIME NOT NULL,
		    expires_at DATETIME NOT NULL,
		    revoked_at DATETIME NULL,
		    revoke_reason VARCHAR(50) NULL,
		    PRIMARY KEY (id),
		    CONSTRAINT fk_session_device FOREIGN KEY (device_id) REFERENCES device (id)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
