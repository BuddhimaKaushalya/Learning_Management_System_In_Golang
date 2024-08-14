package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateRequestTxParams struct {
	Confirm  bool  `json:"confirm"`
	Pending  bool  `json:"pending"`
	CourseID int64 `json:"course_id"`
}
type UpdateRequestTxResult struct {
	Request Request
	Course  Course
}

func (store *SQLStore) UpdateRequestTx(ctx context.Context, arg UpdateRequestTxParams) (UpdateRequestTxResult, error) {
	var result UpdateRequestTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Update the request
		result.Request, err = q.UpdateRequest(ctx, UpdateRequestParams{
			CourseID: arg.CourseID,
			Confirm: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
			Pending: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}

		// Delete the course
		err = q.DeleteCourses(ctx, arg.CourseID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
