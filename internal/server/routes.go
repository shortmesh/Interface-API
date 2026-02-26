package server

import (
	"fmt"
	"net/http"
	"os"

	"interface-api/docs"
	v1 "interface-api/internal/api/v1"
	"interface-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "Oops! something went wrong. Please try again later."

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if code < 500 {
				if msg, ok := he.Message.(string); ok {
					message = msg
				} else {
					message = http.StatusText(code)
				}
			}
		}

		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				c.NoContent(code)
			} else {
				c.JSON(code, map[string]string{"error": message})
			}
		}
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRemoteIP: true,
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logLine := fmt.Sprintf(`%s - - [%s] "%s %s %s" %d %d "%s" "%s" %v`,
				v.RemoteIP,
				v.StartTime.Format("2006-01-02 15:04:05"),
				v.Method,
				v.URI,
				c.Request().Proto,
				v.Status,
				v.ResponseSize,
				c.Request().Referer(),
				c.Request().UserAgent(),
				v.Latency,
			)
			fmt.Fprintln(os.Stdout, logLine)
			return nil
		},
	}))

	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/docs/*", func(c echo.Context) error {
		docs.SwaggerInfo.Host = c.Request().Host
		return echoSwagger.WrapHandler(c)
	})

	apiV1 := e.Group("/api/v1")
	v1.RegisterRoutes(apiV1, s.db)

	return e
}
