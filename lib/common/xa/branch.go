package xa

import (
	"context"
	"errors"
	"regexp"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	mysqldriver "github.com/go-sql-driver/mysql"
	pkgerrors "github.com/pkg/errors"
)

type TransactionBranches interface {
	Prepare(ctx context.Context, transactionID string, action func(ctx context.Context) error) error
	Commit(ctx context.Context, transactionID string) error
	Rollback(ctx context.Context, transactionID string) error
}

const errUnknownXID = 1397

var (
	transactionIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,48}$`)

	ErrInvalidTransactionID = errors.New("transaction id must match [A-Za-z0-9-]{1,48}")
)

type Branches struct {
	client   mysql.TransactionalClient
	branchID string
}

func NewBranches(client mysql.TransactionalClient, branchID string) *Branches {
	return &Branches{
		client:   client,
		branchID: branchID,
	}
}

func (b *Branches) Prepare(ctx context.Context, transactionID string, action func(ctx context.Context) error) (err error) {
	xid, err := b.xid(transactionID)
	if err != nil {
		return err
	}
	conn, err := b.client.Connection(ctx)
	if err != nil {
		return pkgerrors.WithStack(err)
	}
	defer func() {
		err = errors.Join(err, conn.Close())
	}()

	_, err = conn.ExecContext(ctx, "xa.start", "XA START "+xid)
	if err != nil {
		return pkgerrors.WithStack(err)
	}
	err = action(context.WithValue(ctx, branchKey{}, mysql.ClientContext(conn)))
	if err == nil {
		_, err = conn.ExecContext(ctx, "xa.end", "XA END "+xid)
	}
	if err == nil {
		_, err = conn.ExecContext(ctx, "xa.prepare", "XA PREPARE "+xid)
	}
	if err == nil {
		return nil
	}
	rollbackCtx := context.WithoutCancel(ctx)
	_, _ = conn.ExecContext(rollbackCtx, "xa.end", "XA END "+xid)
	_, rollbackErr := conn.ExecContext(rollbackCtx, "xa.rollback", "XA ROLLBACK "+xid)
	if rollbackErr != nil && !isUnknownXID(rollbackErr) {
		err = errors.Join(err, rollbackErr)
	}
	return err
}

func (b *Branches) Commit(ctx context.Context, transactionID string) error {
	return b.finish(ctx, transactionID, "xa.commit", "XA COMMIT ")
}

func (b *Branches) Rollback(ctx context.Context, transactionID string) error {
	return b.finish(ctx, transactionID, "xa.rollback", "XA ROLLBACK ")
}

func (b *Branches) finish(ctx context.Context, transactionID, queryID, statement string) error {
	xid, err := b.xid(transactionID)
	if err != nil {
		return err
	}
	_, err = b.client.ExecContext(ctx, queryID, statement+xid)
	if err != nil && !isUnknownXID(err) {
		return pkgerrors.WithStack(err)
	}
	return nil
}

func (b *Branches) xid(transactionID string) (string, error) {
	if !transactionIDPattern.MatchString(transactionID) {
		return "", pkgerrors.WithStack(ErrInvalidTransactionID)
	}
	return "'" + transactionID + "','" + b.branchID + "'", nil
}

func isUnknownXID(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == errUnknownXID
}

type branchKey struct{}

func Unit(unit uow.UnitOfWork) uow.UnitOfWork {
	return branchUnit{unit: unit}
}

type branchUnit struct {
	unit uow.UnitOfWork
}

func (u branchUnit) ExecuteWithClientContext(ctx context.Context, callback func(client mysql.ClientContext) error) error {
	if client, ok := ctx.Value(branchKey{}).(mysql.ClientContext); ok {
		return callback(client)
	}
	return u.unit.ExecuteWithClientContext(ctx, callback)
}

func LockableUnit(unit uow.LockableUnitOfWork) uow.LockableUnitOfWork {
	return lockableBranchUnit{unit: unit}
}

type lockableBranchUnit struct {
	unit uow.LockableUnitOfWork
}

func (u lockableBranchUnit) ExecuteWithClientContext(ctx context.Context, locks []uow.LockOptions, callback func(client mysql.ClientContext) error) error {
	if client, ok := ctx.Value(branchKey{}).(mysql.ClientContext); ok && len(locks) == 0 {
		return callback(client)
	}
	return u.unit.ExecuteWithClientContext(ctx, locks, callback)
}
