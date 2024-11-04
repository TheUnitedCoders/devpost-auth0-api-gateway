package audit

import (
	"context"
	"log/slog"

	"github.com/TheUnitedCoders/devpost-auth0-api-gateway/internal/domain"
)

// Fields that will be written to audit.
type Fields struct {
	Service string
	Method  string
	Subject *domain.SubjectInformation
	Result  Result
}

// Auditor is a audit message writer.
type Auditor interface {
	Write(ctx context.Context, fields Fields)
}

type logAuditor struct {
	logger *slog.Logger
}

// Write audit message.
func (l logAuditor) Write(ctx context.Context, fields Fields) {
	attrs := []slog.Attr{
		slog.String("service", fields.Service),
		slog.String("method", fields.Method),
		slog.String("result", fields.Result.String()),
	}

	if fields.Subject != nil {
		attrs = append(
			attrs,
			slog.Group(
				"subject",
				slog.String("id", fields.Subject.ID),
				slog.String("permissions", fields.Subject.Permissions.String()),
			),
		)
	}

	l.logger.LogAttrs(ctx, slog.LevelWarn, "audit for request", attrs...)
}

// NewLogAuditor ...
func NewLogAuditor(logger *slog.Logger) Auditor {
	if logger == nil {
		logger = slog.Default()
	}

	return &logAuditor{
		logger: logger,
	}
}
