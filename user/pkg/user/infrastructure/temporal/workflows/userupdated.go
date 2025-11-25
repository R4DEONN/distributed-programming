package workflows

import (
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"

	appmodel "userservice/pkg/user/application/model"
	"userservice/pkg/user/domain/model"
	"userservice/pkg/user/infrastructure/temporal/activity"
)

var userServiceActivities *activity.UserServiceActivities

func UserUpdatedWorkflow(ctx workflow.Context, event model.UserUpdated) error {
	contactInfoChanged := (event.UpdatedFields != nil && (event.UpdatedFields.Telegram != nil || event.UpdatedFields.Email != nil)) ||
		(event.RemovedFields != nil && (event.RemovedFields.Telegram != nil || event.RemovedFields.Email != nil))

	if !contactInfoChanged {
		return nil
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})

	var user appmodel.User
	err := workflow.ExecuteActivity(ctx, userServiceActivities.FindUser, event.UserID).Get(ctx, &user)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil
		}
		return err
	}

	status := model.Blocked
	if user.Email != nil || user.Telegram != nil {
		status = model.Active
	}

	err = workflow.ExecuteActivity(ctx, userServiceActivities.SetUserStatus, event.UserID, int(status)).Get(ctx, nil)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil
		}
		return err
	}

	return nil
}
