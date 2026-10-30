package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"
)

func Version1792180527(client mysql.ClientContext) migrator.Migration {
	return &version1792180527{
		client: client,
	}
}

type version1792180527 struct {
	client mysql.ClientContext
}

func (v *version1792180527) Version() int64 {
	return 1792180527
}

func (v *version1792180527) Description() string {
	return "Create account table"
}

func (v *version1792180527) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-account-table", `
		CREATE TABLE account (
		    id BINARY(16) NOT NULL,
		    email VARCHAR(254) NOT NULL,
		    name VARCHAR(255) NOT NULL,
		    created_at DATETIME NOT NULL,
		    PRIMARY KEY (id),
		    UNIQUE KEY uq_account_email (email)
		) 
		    ENGINE=InnoDB 
		    CHARACTER SET=utf8mb4
		    COLLATE=utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
