package orders

import (
	"bytes"
	"crypto/hmac"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	"net/http"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/internal/auth/entity"
	entity3 "tugas-akhir/backend/internal/bookings/entity"
	entity4 "tugas-akhir/backend/internal/events/entity"
	entity2 "tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/service/idempotent_place_order"
	"tugas-akhir/backend/internal/orders/usecase/get_order"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	"tugas-akhir/backend/internal/orders/usecase/webhook"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/logger"
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
	redis               *redis.Redis
}

func NewBaseOrderHandler(
	validator *myvalidator.TranslatedValidator,
	placeOrderUsecase place_order.PlaceOrderUsecase,
	getOrderUsecase get_order.GetOrderUsecase,
	webhookOrderUsecase webhook.WebhookOrderUsecase,
	config *config.Config,
	redis *redis.Redis,
) *BaseOrderHandler {
	return &BaseOrderHandler{
		validator:           validator,
		placeOrderUsecase:   placeOrderUsecase,
		getOrderUsecase:     getOrderUsecase,
		webhookOrderUsecase: webhookOrderUsecase,
		config:              config,
		redis:               redis,
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

	userToken := c.Get(entity.JwtContextKey).(*jwt.Token)
	tokenClaim := userToken.Claims.(*entity.TokenClaim)

	payload.UserID = &tokenClaim.UserID

	idempotencyKey := c.Request().Header.Get(HeaderIdempotencyKey)

	if idempotencyKey != "" {
		payload.IdempotencyKey = &idempotencyKey
	}

	// check for request type
	var requestType entity4.AreaType

	if payload.Items[0].TicketSeatID == nil {
		requestType = entity4.AreaType__FreeStanding
	} else {
		requestType = entity4.AreaType__NumberedSeating
	}

	l := logger.FromCtx(ctx)
	l = l.With(zap.String("area_type", string(requestType)))
	ctx = logger.WithCtx(ctx, l)
	c.SetRequest(c.Request().WithContext(ctx))

	result, httpErr := idempotent_place_order.WrapIdempotency(ctx, h.redis, h.placeOrderUsecase.PlaceOrder, payload)

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

	OrderID, TicketAreaID, err := utility.ParseNumberString(payload.CompositePK)

	if err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: err.Error(),
		})
	}

	payload.OrderID = OrderID
	payload.TicketAreID = TicketAreaID

	userToken := c.Get(entity.JwtContextKey).(*jwt.Token)
	tokenClaim := userToken.Claims.(*entity.TokenClaim)

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

func (h *BaseOrderHandler) GetIssuedTickets(c echo.Context) error {
	ctx := c.Request().Context()
	var payload entity3.GetIssuedTicketDto

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

	OrderID, TicketAreaID, err := utility.ParseNumberString(payload.CompositePK)

	if err != nil {
		return c.JSON(http.StatusBadRequest, myerror.HttpError{
			Message: err.Error(),
		})
	}

	payload.OrderID = OrderID
	payload.TicketAreaID = TicketAreaID

	userToken := c.Get(entity.JwtContextKey).(*jwt.Token)
	tokenClaim := userToken.Claims.(*entity.TokenClaim)

	payload.UserID = &tokenClaim.UserID

	result, httpErr := h.getOrderUsecase.GetIssuedTicket(ctx, payload)

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
