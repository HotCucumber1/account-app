package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"
)

func Version1792180748(client mysql.ClientContext) migrator.Migration {
	return &version1792180748{
		client: client,
	}
}

type version1792180748 struct {
	client mysql.ClientContext
}

func (v *version1792180748) Version() int64 {
	return 1792180748
}

func (v *version1792180748) Description() string {
	return "Create device table"
}

func (v *version1792180748) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-device-table", `
		CREATE TABLE device (
		    id BINARY(16) NOT NULL,
		    account_id BINARY(16) NOT NULL,
		    user_agent_sha256 BINARY(32) NOT NULL,
		    user_agent VARCHAR(1024) NOT NULL,
		    first_seen_at DATETIME NOT NULL,
		    PRIMARY KEY (id),
		    UNIQUE KEY uq_device_account_user_agent (account_id, user_agent_sha256)
		)
		ENGINE=InnoDB 
		CHARACTER SET=utf8mb4
		COLLATE=utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
