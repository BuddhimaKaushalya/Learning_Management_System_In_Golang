package worker

import (
	"context"
	db "eduApp/db/sqlc"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

const TaskCreateLessonCompletion = "task:create_lesson_completion"

type PayloadCreateLessonCompletion struct {
	UserID     int64 `json:"user_id"`
	CourseID   int64 `json:"course_id"`
	MaterialID int64 `json:"material_id"`
}

func (distributor *RedisTaskDistributor) DistributeTaskCreateLessonCompletion(
	ctx context.Context,
	payload *PayloadCreateLessonCompletion,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskCreateLessonCompletion, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskCreateLessonCompletion(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCreateLessonCompletion
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	params := db.CreateLessonCompletionParams{
		UserID:     payload.UserID,
		CourseID:   payload.CourseID,
		MaterialID: payload.MaterialID,
		Completed:  true,
	}
	_, err := processor.store.CreateLessonCompletion(ctx, params)
	if err != nil {
		log.Error().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("failed to create lesson Completion")
		return fmt.Errorf("failed to create lesson Completion: %w", err)
	}

	count1, err := processor.store.GetTotalMaterialsInCourse(ctx, payload.CourseID)

	if err != nil {
		log.Error().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("failed to get number of materials in a course ")
		return fmt.Errorf("failed to get number of materials in a course : %w", err)
	}

	// Ensure count1 is not 0 to avoid division by zero
	if count1 == 0 {
		return fmt.Errorf("total materials count is 0")
	}

	count2, err := processor.store.GetCompletedLessonsCount(ctx, db.GetCompletedLessonsCountParams{
		CourseID: payload.CourseID,
		UserID:   payload.UserID,
	})
	if err != nil {
		log.Error().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("failed to get number of lessons in a course")
		return fmt.Errorf("failed to get number of lessons in a course: %w", err)
	}

	// Calculate progress as a float64
	progress := float64(count2) / float64(count1) * 100
	println("count2", count2)
	println("count1", count1)

	// Ensure progress is within the expected range
	if progress < 0 || progress > 100 {
		return fmt.Errorf("invalid progress value: %f", progress)
	}

	_, err = processor.store.UpdateCourseProgress(ctx, db.UpdateCourseProgressParams{
		CourseID: pgtype.Int8{Int64: payload.CourseID, Valid: true},
		UserID:   payload.UserID,
		Progress: pgtype.Int8{Int64: int64(progress), Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to update course progress: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("processed task successfully")
	return nil
}
