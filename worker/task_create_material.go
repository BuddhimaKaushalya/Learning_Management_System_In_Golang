package worker

import (
	"context"
	"eduApp/token"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskCreateMaterials = "task:create_materials"

type PayloadCreateMaterials struct {
	MaterialID   int64          `json:"material_id"`
	GetPayload   *token.Payload `json:"get_payload"`
	Title        string         `form:"title"`
	MaterialFile string         `json:"material_file"`
	OrderNumber  int64          `form:"order_number"`
}

func (distributor *RedisTaskDistributor) DistributeTaskCreateMaterials(
	ctx context.Context,
	payload *PayloadCreateMaterials,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskCreateMaterials, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskCreateMaterial(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCreateMaterials
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	authPayload := payload.GetPayload
	if authPayload == nil {
		return fmt.Errorf("no authorization payload")
	}

	user, err := processor.store.GetUser(ctx, authPayload.UserName)
	if err != nil {
		log.Error().Err(err).Str("user_name", authPayload.UserName).Msg("Failed to get user")
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.UserName == "" {
		return fmt.Errorf("user not found")
	}

	if user.UserID != authPayload.UserID {
		log.Info().Int64("user_id", authPayload.UserID).Msg("User does not match course creator")
		return fmt.Errorf("user does not match course creator")
	}

	subject := "Material Created"
	content := fmt.Sprintf(`Hello %s,<br/>
	Your material has been created. You can view it <a href="%s">here</a>.<br/>
	Please do not share this material with others.<br/>`, user.FirstName, payload.MaterialFile)

	to := []string{user.Email}
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
