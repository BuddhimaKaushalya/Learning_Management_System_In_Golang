package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskResetPassword = "task:reset_password"

// PayloadResetPassword contains the task payload data
type PayloadResetPassword struct {
	Email string `json:"email"`
}

// DistributorTaskResetpassword funtion is funtions that takes the payload and marshal it into json and pasit as an task to processor
func (distributor *RedisTaskDistributor) DistributorTaskResetpassword(
	ctx context.Context,
	payload *PayloadResetPassword,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload data: %w", err)
	}

	task := asynq.NewTask(TaskResetPassword, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued_task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskResetpassword(ctx context.Context, task *asynq.Task) error {
	var payload PayloadResetPassword

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	//checking the email exist in db or not
	email, err := processor.store.CheckEmail(ctx, payload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalied email address: %w", err)
		}
		return fmt.Errorf("failed to get data: %w", err)
	}

	subject := "Welcome to EduApp"
	// front-end URL to reset Password
	verifyUrl := " http://172.17.249.61:3390/reset/password"
	content := fmt.Sprintf(`Hello ,<br/>
	We received a request to reset your account password.<br/>
	If you initiated this request, <br/>
	Please <a href="%s">click here</a> to reset your password.<br/>
	Thankyou!
	`, verifyUrl)
	to := []string{email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send password reset request")
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", email).Msg("processed_task")

	return nil
}
