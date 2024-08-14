package db

import "context"

type CreateSubscriptionTxParams struct {
	CreateSubscriptionParams
	AfterCreate func(subscription Subscription) error
}

type CreateSubscriptionTxResult struct {
	Subscription Subscription
}

func (store *SQLStore) CreateSubscriptionTx(ctx context.Context, arg CreateSubscriptionTxParams) (CreateSubscriptionTxResult, error) {
	var result CreateSubscriptionTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Subscription, err = q.CreateSubscription(ctx, arg.CreateSubscriptionParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.Subscription)
	})

	return result, err
}
