package worker

import (
	"context"
	"database/sql"
	db "eduApp/db/sqlc"
	"eduApp/token"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskDeleteCourses = "task:delete_course"
const authorizationPayloadKey = "authorization_payload"

type PayloadDeleteCourse struct {
	CourseID   int64          `json:"course_id"`
	GetPayload *token.Payload `json:"get_payload"`
}

func (distributor *RedisTaskDistributor) DistributeTaskDeleteCourse(
	ctx context.Context,
	payload *PayloadDeleteCourse,
	opts ...asynq.Option,
) error {

	val, ok := ctx.Value(authorizationPayloadKey).(*token.Payload)
	if !ok {
		return fmt.Errorf("failed to get authorization payload")
	}

	payload.GetPayload = val

	print("distribute1", payload.CourseID)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskDeleteCourses, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskDeleteCourse(ctx context.Context, task *asynq.Task) error {
	var payload PayloadDeleteCourse
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	AuthPayload := payload.GetPayload
	if AuthPayload == nil {
		return fmt.Errorf("no authorization payload")
	}

	courseResponse, err := processor.store.GetCourses(ctx, payload.CourseID)
	if err != nil {
		log.Error().Err(err).Int64("course_id", payload.CourseID).Msg("Failed to get course")
		return fmt.Errorf("failed to get course: %w", err)
	}

	if courseResponse.CourseID == 0 {
		return fmt.Errorf("course not found")
	}

	if courseResponse.UserID != payload.GetPayload.UserID {
		log.Info().Int64("user_id", payload.GetPayload.UserID).Msg("User does not match course creator")

		_, err = processor.store.CreateRequest(ctx, db.CreateRequestParams{
			CourseID: courseResponse.CourseID,
			Pending:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		log.Info().Msg("Request created")
		user, err := processor.store.GetUserByID(ctx, courseResponse.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		fmt.Print("Hellooo----")
		// Checking if the email exists in the database
		email, err := processor.store.CheckEmail(ctx, user.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("invalid email address: %w", err)
			}
			return fmt.Errorf("failed to get data: %w", err)
		}

		subject := "Request to Delete Course"
		content := `Hello,<br/>
        You have a request to delete the course created by you.<br/>
        Please check it!<br/>
        Thank you!`

		to := []string{email}

		err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
			Str("email", email).Msg("processed_task")

		return nil
	}

	log.Info().Int64("course_id", payload.CourseID).Msg("Course will be deleted")
	err = processor.store.DeleteCourses(ctx, payload.CourseID)
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}

	return nil
}
