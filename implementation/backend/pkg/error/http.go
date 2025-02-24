package myerror

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"tugas-akhir/backend/pkg/logger"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
}

type HttpError struct {
	Code         int          `json:"-"`
	Message      string       `json:"message"`
	Errors       []FieldError `json:"errors,omitempty"`
	ErrorContext error        `json:"-"`
}

func (e *HttpError) Log(ctx context.Context) {
	if e.ErrorContext != nil {
		l := logger.FromCtx(ctx)
		l.Error("an error occured",
			zap.String("error", e.ErrorContext.Error()),
			zap.String("stackTrace", fmt.Sprintf("%+v", e.ErrorContext)))
	}
}

type HttpPayload struct {
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewFromFieldError(payload []FieldError) HttpError {
	return HttpError{
		Message: "Input validation error",
		Errors:  payload,
	}
}
