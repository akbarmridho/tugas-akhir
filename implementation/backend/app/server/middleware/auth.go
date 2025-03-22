package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	config2 "tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/internal/auth/entity"
	"tugas-akhir/backend/pkg/logger"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	JwtMiddleware echo.MiddlewareFunc
}

func wrapJWTMiddleware(jwtMiddleware echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return jwtMiddleware(func(c echo.Context) error {
			userToken := c.Get(entity.JwtContextKey).(*jwt.Token)
			claims := userToken.Claims.(*entity.TokenClaim)

			ctx := c.Request().Context()
			l := logger.FromCtx(ctx)
			l.With(zap.String("userId", claims.UserID))

			c.SetRequest(c.Request().WithContext(logger.WithCtx(ctx, l)))

			return next(c)
		})
	}
}

func NewAuthMiddleware(config *config2.Config,
) *AuthMiddleware {
	middleware := echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(entity.TokenClaim)
		},
		SigningKey: []byte(config.JwtSecret),
		ContextKey: entity.JwtContextKey,
	})

	return &AuthMiddleware{
		JwtMiddleware: wrapJWTMiddleware(middleware),
	}
}
