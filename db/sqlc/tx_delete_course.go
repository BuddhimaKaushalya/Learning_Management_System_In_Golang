package db

import "context"

type DeleteCourseTxParams struct {
	CourseID    int64 `json:"course_id"`
	AfterCreate func(course Course) error
}

type DeleteCourseTxResult struct {
	Course Course
}

func (store *SQLStore) DeleteCoursesTx(ctx context.Context, arg DeleteCourseTxParams) (DeleteCourseTxResult, error) {
	var result DeleteCourseTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		println("DeleteCoursesTx1", arg.CourseID)
		err = q.DeleteCourses(ctx, arg.CourseID)

		if err != nil {
			return err
		}
		println("DeleteCoursesTx2", arg.CourseID)

		return arg.AfterCreate(result.Course)
	})
	return result, err
}
