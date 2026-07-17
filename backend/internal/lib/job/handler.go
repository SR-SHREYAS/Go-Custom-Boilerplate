package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/config"
	"github.com/SR-SHREYAS/Go-Custom-Boilerplate/internal/lib/email"
	"github.com/hibiken/asynq"
	zerolog "github.com/jackc/pgx-zerolog"
)

var emailClient *email.Client

func (j *JobService) InitHandler(config *config.Config, logger *zerolog.Logger) {
	emailClient = email.NewClient(config, logger)
}

func (j *JobService) handleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("Failed to unmarshal welcome email payload: %w", err)
	}

	j.logger.Info().
		Str("type", "welcome").
		Str("to", p.To).
		Msg("Processing welcome email task")

	err := emailClient.SendWelcomeEmail(
		p.To,
		p.FirstName,
	)

	if err != nil {
		j.logger.Error().
			Str("type", "welcome").
			Str("to", p.To).
			Err(err).
			Msg("Failed to send welcome email")
		return err
	}

	j.logger.Info().
		Str("type", "welcome").
		Str("to", p.To).
		Msg("Successfully sent welcome email")
	return nil
}
