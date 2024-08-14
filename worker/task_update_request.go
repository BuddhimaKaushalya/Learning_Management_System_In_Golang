package worker

import (
	"context"
	db "eduApp/db/sqlc"
	"eduApp/token"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

const TaskUpdateRequest = "task:update_request"

type PayloadUpdateRequest struct {
	RequestID  int64          `json:"request_id"`
	GetPayload *token.Payload `json:"get_payload"`
}

func (distributor *RedisTaskDistributor) DistributeTaskUpdateRequest(
	ctx context.Context,
	payload *PayloadUpdateRequest,
	opts ...asynq.Option,
) error {

	val, ok := ctx.Value(authorizationPayloadKey).(*token.Payload)
	if !ok {
		return fmt.Errorf("failed to get authorization payload")
	}

	payload.GetPayload = val

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskUpdateRequest, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskUpdateRequest(ctx context.Context, task *asynq.Task) error {
	var payload PayloadUpdateRequest
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	AuthPayload := payload.GetPayload
	if AuthPayload == nil {
		return fmt.Errorf("no authorization payload")
	}

	RequestResponse, err := processor.store.GetRequest(ctx, payload.RequestID)
	if err != nil {
		log.Error().Err(err).Int64("request_id", payload.RequestID).Msg("Failed to get request")
		return fmt.Errorf("failed to get request: %w", err)
	}

	courseResponse, err := processor.store.GetCourses(ctx, RequestResponse.CourseID)
	if err != nil {
		log.Error().Err(err).Int64("course_id", RequestResponse.CourseID).Msg("Failed to get course")
		return fmt.Errorf("failed to get course: %w", err)
	}
	println("course_id1", RequestResponse.CourseID)

	if courseResponse.CourseID == payload.GetPayload.UserID {
		_, err = processor.store.UpdateRequest(ctx, db.UpdateRequestParams{
			CourseID: RequestResponse.CourseID,
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
			return fmt.Errorf("failed to update request: %w", err)
		}
		println("course_id2", RequestResponse.CourseID)
		err = processor.store.DeleteCourses(ctx, RequestResponse.CourseID)
		if err != nil {
			return fmt.Errorf("failed to delete course")
		}
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Msg("processed task")
	return nil
}
