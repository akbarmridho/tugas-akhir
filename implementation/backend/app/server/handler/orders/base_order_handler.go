package orders

import (
	"bytes"
	"crypto/hmac"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/auth/entity"
	entity2 "tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/usecase/get_order"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	"tugas-akhir/backend/internal/orders/usecase/webhook"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/mock_payment"
	"tugas-akhir/backend/pkg/utility"
	myvalidator "tugas-akhir/backend/pkg/validator"
)

type BaseOrderHandler struct {
	validator           *myvalidator.TranslatedValidator
	placeOrderUsecase   place_order.PlaceOrderUsecase
	getOrderUsecase     get_order.GetOrderUsecase
	webhookOrderUsecase webhook.WebhookOrderUsecase
	config              *config.Config
}

func NewBaseOrderHandler(
	validator *myvalidator.TranslatedValidator,
	placeOrderUsecase place_order.PlaceOrderUsecase,
	getOrderUsecase get_order.GetOrderUsecase,
	webhookOrderUsecase webhook.WebhookOrderUsecase,
	config *config.Config,
) *BaseOrderHandler {
	return &BaseOrderHandler{
		validator:           validator,
		placeOrderUsecase:   placeOrderUsecase,
		getOrderUsecase:     getOrderUsecase,
		webhookOrderUsecase: webhookOrderUsecase,
		config:              config,
	}
}

func (h *BaseOrderHandler) PlaceOrder(c echo.Context) error {
	ctx := c.Request().Context()
	var payload entity2.PlaceOrderDto

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

	tokenClaim := c.Get(entity.JwtContextKey).(*entity.TokenClaim)

	payload.UserID = &tokenClaim.UserID

	result, httpErr := h.placeOrderUsecase.PlaceOrder(ctx, payload)

	if httpErr != nil {
		httpErr.Log(ctx)
		return c.JSON(httpErr.Code, httpErr)
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Data: result,
	})
}

func (h *BaseOrderHandler) GetOrder(c echo.Context) error {
	ctx := c.Request().Context()
	var payload entity2.GetOrderDto

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

	tokenClaim := c.Get(entity.JwtContextKey).(*entity.TokenClaim)

	payload.UserID = &tokenClaim.UserID

	result, httpErr := h.getOrderUsecase.GetOrder(ctx, payload)

	if httpErr != nil {
		httpErr.Log(ctx)
		return c.JSON(httpErr.Code, httpErr)
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Data: result,
	})
}

func (h *BaseOrderHandler) HandleWebhook(c echo.Context) error {
	ctx := c.Request().Context()

	raw, err := io.ReadAll(c.Request().Body)

	if err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: "Cannot read request body",
		})
	}

	rawStr := string(raw)

	c.Request().Body = io.NopCloser(bytes.NewReader(raw))

	given := c.Request().Header.Get("x-webhook-verify")

	// verify the webhook token
	computed := utility.ComputeHMACSHA256(h.config.WebhookSecret, rawStr)

	if !hmac.Equal([]byte(given), []byte(computed)) {
		return c.JSON(http.StatusForbidden, myerror.HttpError{
			Message: "Given and computed webhook token is different",
		})
	}

	var payload mock_payment.Invoice

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

	httpErr := h.webhookOrderUsecase.HandleWebhook(ctx, payload)

	if httpErr != nil {
		httpErr.Log(ctx)
		return c.JSON(httpErr.Code, httpErr)
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Message: "ok",
	})
}
