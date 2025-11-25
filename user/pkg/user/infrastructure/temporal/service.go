package temporal

import (
	"context"

	"go.temporal.io/sdk/client"

	"userservice/pkg/user/domain/model"
	"userservice/pkg/user/infrastructure/temporal/workflows"
)

const TaskQueue = "userservice_task_queue"

type WorkflowService interface {
	RunUserUpdatedWorkflow(ctx context.Context, id string, event model.UserUpdated) error
}

func NewWorkflowService(temporalClient client.Client) WorkflowService {
	return &workflowService{
		temporalClient: temporalClient,
	}
}

type workflowService struct {
	temporalClient client.Client
}

func (s *workflowService) RunUserUpdatedWorkflow(ctx context.Context, id string, event model.UserUpdated) error {
	_, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        id,
			TaskQueue: TaskQueue,
		},
		workflows.UserUpdatedWorkflow, event,
	)
	return err
}
