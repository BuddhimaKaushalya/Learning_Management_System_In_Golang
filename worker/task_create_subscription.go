package worker

import (
	"context"
	db "eduApp/db/sqlc"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskCreateSubscription = "task:create_subscription"

type PayloadCreateSubscription struct {
	UserID   int64 `json:"user_id"`
	CourseID int64 `json:"course_id"`
}

func (distributor *RedisTaskDistributor) DistributeTaskCreateSubscription(
	ctx context.Context,
	payload *PayloadCreateSubscription,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskCreateSubscription, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskCreateSubscription(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCreateSubscription
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	params := db.CreateSubscriptionParams{
		UserID:   payload.UserID,
		CourseID: payload.CourseID,
	}

	subscription, err := processor.store.CreateSubscription(ctx, params)
	if err != nil {
		log.Error().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("failed to create subscription")
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	_, err = processor.store.CreateCourseProgress(ctx, db.CreateCourseProgressParams{
		CourseID: subscription.CourseID,
		UserID:   params.UserID,
		Progress: 0,
	})
	if err != nil {
		log.Error().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("failed to create course progress")
		return fmt.Errorf("failed to create course progress: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("processed task successfully")
	return nil
}
