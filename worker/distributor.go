package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskCreateSubscription(
		ctx context.Context,
		payload *PayloadCreateSubscription,
		opts ...asynq.Option,
	) error
	DistributeTaskDeleteCourse(
		ctx context.Context,
		payload *PayloadDeleteCourse,
		opts ...asynq.Option,
	) error
	DistributeTaskUpdateCourseprogress(
		ctx context.Context,
		payload *PayloadUpdateCourseprogress,
		opts ...asynq.Option,
	) error
	DistributeTaskUpdateRequest(
		ctx context.Context,
		payload *PayloadUpdateRequest,
		opts ...asynq.Option,
	) error
	DistributorTaskResetpassword(
		ctx context.Context,
		payload *PayloadResetPassword,
		opts ...asynq.Option,
	) error
	DistributeTaskCreateLessonCompletion(
		ctx context.Context,
		payload *PayloadCreateLessonCompletion,
		opts ...asynq.Option,
	) error
	DistributeTaskCreateMaterials(
		ctx context.Context,
		payload *PayloadCreateMaterials,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
