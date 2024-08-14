package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateCourseProgressTxParams struct {
	CourseID    int64 `json:"course_id"`
	USerID      int64 `json:"user_id"`
	Progress    int64 `json:"progress"`
	AfterCreate func(courseprogress CourseProgress) error
}

type UpdateCourseProgressTxResult struct {
	CourseProgress CourseProgress
}

func (store *SQLStore) UpdateCourseProgressTx(ctx context.Context, arg UpdateCourseProgressTxParams) (UpdateCourseProgressTxResult, error) {
	var result UpdateCourseProgressTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.CourseProgress, err = q.UpdateCourseProgress(ctx, UpdateCourseProgressParams{
			CourseID: pgtype.Int8{Int64: arg.CourseID, Valid: true},
			UserID:   arg.USerID,
			Progress: pgtype.Int8{Int64: arg.Progress, Valid: true},
		})

		if err != nil {
			return err
		}

		return arg.AfterCreate(result.CourseProgress)
	})
	return result, err
}
