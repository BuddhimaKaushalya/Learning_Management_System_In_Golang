// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: lesson_completion.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLessonCompletion = `-- name: CreateLessonCompletion :one
INSERT INTO lesson_completion (
    course_id,
    user_id,
    material_id,
    completed
) VALUES (
    $1, $2, $3, $4
) RETURNING completion_id, user_id, course_id, material_id, completed, completed_at
`

type CreateLessonCompletionParams struct {
	CourseID   int64 `json:"course_id"`
	UserID     int64 `json:"user_id"`
	MaterialID int64 `json:"material_id"`
	Completed  bool  `json:"completed"`
}

func (q *Queries) CreateLessonCompletion(ctx context.Context, arg CreateLessonCompletionParams) (LessonCompletion, error) {
	row := q.db.QueryRow(ctx, createLessonCompletion,
		arg.CourseID,
		arg.UserID,
		arg.MaterialID,
		arg.Completed,
	)
	var i LessonCompletion
	err := row.Scan(
		&i.CompletionID,
		&i.UserID,
		&i.CourseID,
		&i.MaterialID,
		&i.Completed,
		&i.CompletedAt,
	)
	return i, err
}

const deleteLessonCompletion = `-- name: DeleteLessonCompletion :exec
DELETE FROM lesson_completion
WHERE completion_id = $1
`

func (q *Queries) DeleteLessonCompletion(ctx context.Context, completionID int64) error {
	_, err := q.db.Exec(ctx, deleteLessonCompletion, completionID)
	return err
}

const getCompletedLessonsCount = `-- name: GetCompletedLessonsCount :one
SELECT COUNT(DISTINCT material_id) AS completed_lessons_count
FROM lesson_completion
WHERE user_id = $1 AND course_id = $2 AND completed = true
`

type GetCompletedLessonsCountParams struct {
	UserID   int64 `json:"user_id"`
	CourseID int64 `json:"course_id"`
}

func (q *Queries) GetCompletedLessonsCount(ctx context.Context, arg GetCompletedLessonsCountParams) (int64, error) {
	row := q.db.QueryRow(ctx, getCompletedLessonsCount, arg.UserID, arg.CourseID)
	var completed_lessons_count int64
	err := row.Scan(&completed_lessons_count)
	return completed_lessons_count, err
}

const getLessonCompletion = `-- name: GetLessonCompletion :one
SELECT completion_id, user_id, course_id, material_id, completed, completed_at FROM lesson_completion
WHERE 
    user_id = $1
    AND course_id = $2
    AND material_id = $3
LIMIT 1
`

type GetLessonCompletionParams struct {
	UserID     int64 `json:"user_id"`
	CourseID   int64 `json:"course_id"`
	MaterialID int64 `json:"material_id"`
}

func (q *Queries) GetLessonCompletion(ctx context.Context, arg GetLessonCompletionParams) (LessonCompletion, error) {
	row := q.db.QueryRow(ctx, getLessonCompletion, arg.UserID, arg.CourseID, arg.MaterialID)
	var i LessonCompletion
	err := row.Scan(
		&i.CompletionID,
		&i.UserID,
		&i.CourseID,
		&i.MaterialID,
		&i.Completed,
		&i.CompletedAt,
	)
	return i, err
}

const updateLessonCompletion = `-- name: UpdateLessonCompletion :one
UPDATE lesson_completion
SET  
       completed = COALESCE($1, completed)
WHERE  user_id = $2 AND course_id = $3 AND material_id = $4
RETURNING completion_id, user_id, course_id, material_id, completed, completed_at
`

type UpdateLessonCompletionParams struct {
	Completed  pgtype.Bool `json:"completed"`
	UserID     int64       `json:"user_id"`
	CourseID   int64       `json:"course_id"`
	MaterialID int64       `json:"material_id"`
}

func (q *Queries) UpdateLessonCompletion(ctx context.Context, arg UpdateLessonCompletionParams) (LessonCompletion, error) {
	row := q.db.QueryRow(ctx, updateLessonCompletion,
		arg.Completed,
		arg.UserID,
		arg.CourseID,
		arg.MaterialID,
	)
	var i LessonCompletion
	err := row.Scan(
		&i.CompletionID,
		&i.UserID,
		&i.CourseID,
		&i.MaterialID,
		&i.Completed,
		&i.CompletedAt,
	)
	return i, err
}
