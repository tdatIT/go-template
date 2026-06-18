package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	userapp "github.com/tdatIT/go-template/internal/app/user"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/domain/msgs"
	mqttclient "github.com/tdatIT/go-template/pkgs/mqtt"
)

const userEventsTopic = "events/user"

// userEventPayload is the expected JSON structure on the events/user topic.

// UserEventWorker subscribes to user-domain events and dispatches them to the user application.
type UserEventWorker struct {
	client mqttclient.Client
	app    *userapp.Application
}

func NewUserEventWorker(client mqttclient.Client, app *userapp.Application) *UserEventWorker {
	return &UserEventWorker{client: client, app: app}
}

// Start implements Worker. Subscribes to the user events topic and blocks until ctx is cancelled.
func (w *UserEventWorker) Start(ctx context.Context) error {
	err := w.client.Subscribe(userEventsTopic, 1, func(topic string, payload []byte) {
		// Dispatch in a new goroutine so the MQTT message loop is never blocked.
		go w.handle(ctx, payload)
	})
	if err != nil {
		return fmt.Errorf("user event worker: subscribe %s: %w", userEventsTopic, err)
	}

	slog.Info("user event worker started", slog.String("topic", userEventsTopic))
	<-ctx.Done()
	slog.Info("user event worker stopped")
	return nil
}

func (w *UserEventWorker) handle(ctx context.Context, raw []byte) {
	var event msgs.UserEventPayload
	if err := json.Unmarshal(raw, &event); err != nil {
		slog.Error("user event: malformed payload", slog.String("error", err.Error()))
		return
	}

	slog.Info("user event received",
		slog.String("action", event.Action),
		slog.Uint64("user_id", uint64(event.UserID)),
	)

	switch event.Action {
	case "delete":
		req := &userdtos.DeleteUserReq{ID: event.UserID}
		if err := w.app.Command.DeleteUser.Handle(ctx, req); err != nil {
			slog.Error("user event: delete user failed",
				slog.Uint64("user_id", uint64(event.UserID)),
				slog.String("error", err.Error()),
			)
		}
	default:
		slog.Warn("user event: unknown action", slog.String("action", event.Action))
	}
}
