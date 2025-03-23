package events

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/usecase"
	myerror "tugas-akhir/backend/pkg/error"
	myvalidator "tugas-akhir/backend/pkg/validator"
)

type EventHandler struct {
	validator    *myvalidator.TranslatedValidator
	eventUsecase *usecase.EventUsecase
}

func NewEventHandler(
	validator *myvalidator.TranslatedValidator,
	eventUsecase *usecase.EventUsecase,
) *EventHandler {
	return &EventHandler{
		validator:    validator,
		eventUsecase: eventUsecase,
	}
}

func (h *EventHandler) GetAvailability(c echo.Context) error {
	ctx := c.Request().Context()
	var payload entity.GetAvailabilityDto

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: "Malformed payload",
		})
	}

	validationError, err := h.validator.Validate(payload)

	if err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: err.Error(),
		})
	}

	if len(validationError) != 0 {
		return c.JSON(http.StatusBadRequest, myerror.NewFromFieldError(validationError))
	}

	result, httpErr := h.eventUsecase.GetAvailability(ctx, payload)

	if httpErr != nil {
		httpErr.Log(ctx)
		return c.JSON(httpErr.Code, httpErr)
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Data: result,
	})
}

func (h *EventHandler) GetSeats(c echo.Context) error {
	ctx := c.Request().Context()
	var payload entity.GetSeatsDto

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: "Malformed payload",
		})
	}

	validationError, err := h.validator.Validate(payload)

	if err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: err.Error(),
		})
	}

	if len(validationError) != 0 {
		return c.JSON(http.StatusBadRequest, myerror.NewFromFieldError(validationError))
	}

	result, httpErr := h.eventUsecase.GetSeats(ctx, payload)

	if httpErr != nil {
		httpErr.Log(ctx)
		return c.JSON(httpErr.Code, httpErr)
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Data: result,
	})
}

func (h *EventHandler) GetEvent(c echo.Context) error {
	ctx := c.Request().Context()
	var payload entity.GetEventDto

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: "Malformed payload",
		})
	}

	validationError, err := h.validator.Validate(payload)

	if err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: err.Error(),
		})
	}

	if len(validationError) != 0 {
		return c.JSON(http.StatusBadRequest, myerror.NewFromFieldError(validationError))
	}

	result, httpErr := h.eventUsecase.GetEvent(ctx, payload)

	if httpErr != nil {
		httpErr.Log(ctx)
		return c.JSON(httpErr.Code, httpErr)
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Data: result,
	})
}

func (h *EventHandler) GetEvents(c echo.Context) error {
	ctx := c.Request().Context()

	result, httpErr := h.eventUsecase.GetEvents(ctx)

	if httpErr != nil {
		httpErr.Log(ctx)
		return c.JSON(httpErr.Code, httpErr)
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Data: result,
	})
}
