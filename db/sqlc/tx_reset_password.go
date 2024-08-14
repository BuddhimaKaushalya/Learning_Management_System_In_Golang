package db

import "context"

// CheckEmailTxParams contains the input parameters for check email , and after funtion
type CheckEmailTxParams struct {
	Email       string `form:"email"`
	AfterCreate func(user User) error
}

// CheckEmailTxResult contains the result of the email checking process
type CheckEmailTxResult struct {
	User User
}

// CheckEmailTx db handler for api call to check if given email exits in the db or not
func (store *SQLStore) CheckEmailTx(ctx context.Context, arg CheckEmailTxParams) (CheckEmailTxResult, error) {
	var result CheckEmailTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result, err := q.CheckEmail(ctx, arg.Email)
		if err != nil {
			return err
		}
		return arg.AfterCreate(User{Email: result})
	})
	return result, err
}
