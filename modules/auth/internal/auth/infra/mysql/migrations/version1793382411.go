package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"
)

func Version1793382411(client mysql.ClientContext) migrator.Migration {
	return &version1793382411{
		client: client,
	}
}

type version1793382411 struct {
	client mysql.ClientContext
}

func (v *version1793382411) Version() int64 {
	return 1793382411
}

func (v *version1793382411) Description() string {
	return "Create registration_transaction table"
}

func (v *version1793382411) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-registration-transaction-table", `
		CREATE TABLE registration_transaction (
		    transaction_id VARCHAR(50) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
		    account_id BINARY(16) NOT NULL,
		    status VARCHAR(20) CHARACTER SET ascii NOT NULL,
		    deadline DATETIME NOT NULL,
		    created_at DATETIME NOT NULL,
		    updated_at DATETIME NOT NULL,
		    PRIMARY KEY (transaction_id),
		    KEY ix_registration_transaction_status (status, updated_at)
		) ENGINE=InnoDB
	`)
	return errors.WithStack(err)
}
