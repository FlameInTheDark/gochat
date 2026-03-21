package observability

import (
	"context"
	"errors"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

type contextualHandler struct {
	next               slog.Handler
	serviceName        string
	deploymentEnvValue string
	fixedAttrs         []slog.Attr
}

func (h *contextualHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *contextualHandler) Handle(ctx context.Context, record slog.Record) error {
	cloned := record.Clone()
	if h.serviceName != "" {
		cloned.AddAttrs(slog.String("service.name", h.serviceName))
	}
	if h.deploymentEnvValue != "" {
		cloned.AddAttrs(slog.String("deployment.environment", h.deploymentEnvValue))
	}
	for _, attr := range h.fixedAttrs {
		cloned.AddAttrs(attr)
	}
	for _, attr := range helper.AttrsFromContext(ctx) {
		cloned.AddAttrs(attr)
	}
	return h.next.Handle(ctx, cloned)
}

func (h *contextualHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextualHandler{
		next:               h.next.WithAttrs(attrs),
		serviceName:        h.serviceName,
		deploymentEnvValue: h.deploymentEnvValue,
		fixedAttrs:         cloneSlogAttrs(h.fixedAttrs),
	}
}

func (h *contextualHandler) WithGroup(name string) slog.Handler {
	return &contextualHandler{
		next:               h.next.WithGroup(name),
		serviceName:        h.serviceName,
		deploymentEnvValue: h.deploymentEnvValue,
		fixedAttrs:         cloneSlogAttrs(h.fixedAttrs),
	}
}

type fanoutHandler struct {
	handlers []slog.Handler
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler != nil && handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler == nil {
			continue
		}
		if err := handler.Handle(ctx, record.Clone()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		if handler == nil {
			continue
		}
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	return &fanoutHandler{handlers: handlers}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		if handler == nil {
			continue
		}
		handlers = append(handlers, handler.WithGroup(name))
	}
	return &fanoutHandler{handlers: handlers}
}

func cloneSlogAttrs(attrs []slog.Attr) []slog.Attr {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]slog.Attr, len(attrs))
	copy(out, attrs)
	return out
}
