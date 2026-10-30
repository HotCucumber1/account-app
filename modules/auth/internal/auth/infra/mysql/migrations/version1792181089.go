package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"
)

func Version1792181089(client mysql.ClientContext) migrator.Migration {
	return &version1792181089{
		client: client,
	}
}

type version1792181089 struct {
	client mysql.ClientContext
}

func (v *version1792181089) Version() int64 {
	return 1792181089
}

func (v *version1792181089) Description() string {
	return "Create credential table"
}

func (v *version1792181089) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-credential-table", `
		CREATE TABLE credential (
		    account_id BINARY(16) NOT NULL,
		    password_hash VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
		    created_at DATETIME(6) NOT NULL,
		    updated_at DATETIME(6) NOT NULL,
		    PRIMARY KEY (account_id)
		) ENGINE=InnoDB
	`)
	return errors.WithStack(err)
}
