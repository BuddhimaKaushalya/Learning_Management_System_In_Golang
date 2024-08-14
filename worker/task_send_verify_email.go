package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	db "eduApp/db/sqlc"
	"eduApp/util"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	UserName string `json:"user_name"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.UserName)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		UserID:     user.UserID,
		Email:      user.Email,
		SecretCode: util.RandomString(6),
	})

	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	_, err = processor.store.CreateUserStatus(ctx, db.CreateUserStatusParams{
		UserID:  user.UserID,
		Active:  false,
		Pending: true,
	})

	if err != nil {
		return fmt.Errorf("failed to create user status: %w", err)
	}

	baseUrl := os.Getenv("VERIFY_EMAIL_BASE_URL")
	if baseUrl == "" {
		baseUrl = "http://172.17.249.61:3390"
	}
	verifyUrl := fmt.Sprintf("%s/verifyemail?email_id=%d&secret_code=%s", baseUrl, verifyEmail.EmailID, verifyEmail.SecretCode)
	commonContent := fmt.Sprintf(`Hello %s,<br/>
	Your OTP is %s. Don't share it with others!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>`, user.FirstName, verifyEmail.SecretCode, verifyUrl)

	var subject, content string
	if user.Role == "admin" {
		subject = "Welcome to EduApp Teaching Committee"
		content = fmt.Sprintf(`%s<br/>Welcome to EduApp Teaching committee!<br/>`, commonContent)
	} else if user.Role == "student" {
		subject = "Welcome to EduApp"
		content = fmt.Sprintf(`%s<br/>Thank you for registering with us!<br/>`, commonContent)
	} else {
		return fmt.Errorf("unknown user role: %s", user.Role)
	}

	to := []string{user.Email}
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
