package server

import (
	"io"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/admin"
	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	prommetrics "github.com/Delpak/shahrah/services/control-plane/internal/observability/prometheus"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

var redactedLogHeaders = map[string]struct{}{
	"authorization":            {},
	"proxy-authorization":      {},
	"cookie":                   {},
	"set-cookie":               {},
	"x-api-key":                {},
	"api-key":                  {},
	"x-auth-token":             {},
	"x-forwarded-access-token": {},
}

// Dependencies contains the concrete services required by the control-plane server.
type Dependencies struct {
	AuthMiddleware          *auth.AuthMiddleware
	TenantResolver          *tenant.Middleware
	TracingMiddleware       fiber.Handler
	TraceHealth             TraceHealthReporter
	AdminHandlers           *admin.Handlers
	AuthHandlers            *admin.AuthHandlers
	ProxyHandler            *proxy.ChatCompletionHandler
	CompletionsHandler      *proxy.CompletionsHandler
	ResponsesHandler        *proxy.ResponsesHandler
	EmbeddingsHandler       *proxy.EmbeddingsHandler
	ModerationsHandler      *proxy.ModerationsHandler
	ImagesHandler           *proxy.ImagesHandler
	ModelsHandler           *proxy.ModelsHandler
	Logger                  *zap.Logger
	PlatformServiceAccounts auth.PlatformServiceAccountStore
}

type TraceHealthReporter interface {
	ReadinessReport() map[string]interface{}
}

// NewFiberApp creates and configures a new Fiber application.
func NewFiberApp(cfg *config.Config, deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorHandler: customErrorHandler,
	})

	httpMetrics := prommetrics.NewHTTPMetrics()

	app.Use(fiberlogger.New(buildHTTPLoggerConfig(nil)))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, api-key, x-api-key, " + apiheaders.HeaderTenant + ", " + apiheaders.LegacyHeaderTenant + ", " + apiheaders.HeaderGuardrails + ", " + apiheaders.LegacyHeaderGuardrails + ", " + apiheaders.HeaderGuardrailsScope + ", " + apiheaders.LegacyHeaderGuardrailScope + ", x-tfy-guardrails, x-tfy-guardrails-scope",
	}))

	if deps.TracingMiddleware != nil {
		app.Use(deps.TracingMiddleware)
	}

	app.Use(httpMetrics.Middleware())

	registerPublicRoutes(app, cfg, deps, httpMetrics)

	if deps.TenantResolver != nil {
		app.Use(deps.TenantResolver.Handle())
	}

	if deps.AuthMiddleware != nil {
		app.Use(deps.AuthMiddleware.Middleware())
	}

	registerProtectedRoutes(app, deps)

	return app
}

func buildHTTPLoggerConfig(output io.Writer) fiberlogger.Config {
	cfg := fiberlogger.Config{
		TimeFormat:    "2006-01-02T15:04:05.000Z07:00",
		TimeZone:      "UTC",
		DisableColors: true,
		Format:        "ts=${time} status=${status} latency=${latency} ip=${ip} method=${method} path=${path} trace_id=${respHeader:" + tracing.TraceIDHeader + "} ua=\"${safe_ua}\" bytes_in=${bytesReceived} bytes_out=${bytesSent} req_headers=\"${safe_req_headers}\" resp_headers=\"${safe_resp_headers}\" error=\"${safe_error}\"\n",
		CustomTags: map[string]fiberlogger.LogFunc{
			"safe_req_headers": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				return output.WriteString(formatHeaderLogValue(requestHeaders(c)))
			},
			"safe_resp_headers": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				return output.WriteString(formatHeaderLogValue(responseHeaders(c)))
			},
			"safe_ua": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				return output.WriteString(sanitizeLogValue(c.Get(fiber.HeaderUserAgent)))
			},
			"safe_error": func(output fiberlogger.Buffer, c *fiber.Ctx, _ *fiberlogger.Data, _ string) (int, error) {
				if err := c.Context().Err(); err != nil {
					return output.WriteString(sanitizeLogValue(err.Error()))
				}

				return 0, nil
			},
		},
	}
	if output != nil {
		cfg.Output = output
	}

	return cfg
}

func requestHeaders(c *fiber.Ctx) map[string][]string {
	out := map[string][]string{}

	c.Request().Header.VisitAll(func(key, value []byte) {
		k := string(key)
		out[k] = append(out[k], string(value))
	})

	return out
}

func responseHeaders(c *fiber.Ctx) map[string][]string {
	out := map[string][]string{}

	c.Response().Header.VisitAll(func(key, value []byte) {
		k := string(key)
		out[k] = append(out[k], string(value))
	})

	return out
}

func formatHeaderLogValue(headers map[string][]string) string {
	if len(headers) == 0 {
		return ""
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		values := headers[key]
		for _, value := range values {
			parts = append(parts, key+"="+sanitizeHeaderValue(key, value))
		}
	}

	return sanitizeLogValue(strings.Join(parts, ";"))
}

func sanitizeHeaderValue(key string, value string) string {
	if _, ok := redactedLogHeaders[strings.ToLower(strings.TrimSpace(key))]; ok {
		return "[REDACTED]"
	}

	return sanitizeLogValue(value)
}

func sanitizeLogValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}

	v = strings.ReplaceAll(v, "\\", "\\\\")
	v = strings.ReplaceAll(v, "\"", "\\\"")
	v = strings.ReplaceAll(v, "\r", "\\r")
	v = strings.ReplaceAll(v, "\n", "\\n")

	return v
}

func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": "0.2.0",
	})
}

func readyHandler(deps Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		resp := fiber.Map{
			"status":  "ready",
			"version": "0.2.0",
		}
		if deps.TraceHealth != nil {
			resp["tracing"] = deps.TraceHealth.ReadinessReport()
		}

		return c.JSON(resp)
	}
}

func customErrorHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}

	return ctx.Status(code).JSON(fiber.Map{
		"error": map[string]interface{}{
			"message": msg,
			"type":    "server_error",
			"code":    code,
		},
	})
}
