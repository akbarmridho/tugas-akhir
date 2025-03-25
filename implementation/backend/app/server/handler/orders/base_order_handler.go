package orders

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"tugas-akhir/backend/internal/auth/entity"
	entity2 "tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	myerror "tugas-akhir/backend/pkg/error"
	myvalidator "tugas-akhir/backend/pkg/validator"
)

type BaseOrderHandler struct {
	validator         *myvalidator.TranslatedValidator
	placeOrderUsecase place_order.PlaceOrderUsecase
}

func NewBaseOrderHandler(
	validator *myvalidator.TranslatedValidator,
	placeOrderUsecase place_order.PlaceOrderUsecase,
) *BaseOrderHandler {
	return &BaseOrderHandler{
		validator:         validator,
		placeOrderUsecase: placeOrderUsecase,
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
