package db

import "context"

type CreateLessonCompletionTxParams struct {
	CreateLessonCompletionParams
	AfterCreate func(LessonCompletion) error
}

type CreateLessonCompletionTxResult struct {
	LessonCompletion LessonCompletion
}

func (store *SQLStore) CreateLessonCompletionTx(ctx context.Context, arg CreateLessonCompletionTxParams) (CreateLessonCompletionTxResult, error) {
	var result CreateLessonCompletionTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.LessonCompletion, err = q.CreateLessonCompletion(ctx, CreateLessonCompletionParams{
			CourseID:   arg.CourseID,
			UserID:     arg.UserID,
			MaterialID: arg.MaterialID,
			Completed:  true,
		})

		if err != nil {
			return err
		}

		if arg.AfterCreate != nil {
			return arg.AfterCreate(result.LessonCompletion)
		}

		return nil

	})

	return result, err
}
