package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	CreateMaterialTx(ctx context.Context, arg CreateMaterialTxParams) (CreateMaterialTxResult, error)
	DeleteCoursesTx(ctx context.Context, arg DeleteCourseTxParams) (DeleteCourseTxResult, error)
	CreateLessonCompletionTx(ctx context.Context, arg CreateLessonCompletionTxParams) (CreateLessonCompletionTxResult, error)
	CreateSubscriptionTx(ctx context.Context, arg CreateSubscriptionTxParams) (CreateSubscriptionTxResult, error)
	CheckEmailTx(ctx context.Context, arg CheckEmailTxParams) (CheckEmailTxResult, error)
	UpdateRequestTx(ctx context.Context, arg UpdateRequestTxParams) (UpdateRequestTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

// store provide all funtions to execute db queries and data trival and transfers
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// create NewStore
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
