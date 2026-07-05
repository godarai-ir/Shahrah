// Package profiling exposes opt-in Go runtime profiling routes.
package profiling

import (
	"net/http"
	"net/http/pprof"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

// RegisterPprofRoutes attaches the standard net/http pprof handlers to Fiber.
func RegisterPprofRoutes(app *fiber.App) {
	app.Get("/debug/pprof", adaptor.HTTPHandlerFunc(pprof.Index))
	app.Get("/debug/pprof/", adaptor.HTTPHandlerFunc(pprof.Index))
	app.Get("/debug/pprof/cmdline", adaptor.HTTPHandlerFunc(pprof.Cmdline))
	app.Get("/debug/pprof/profile", adaptor.HTTPHandlerFunc(pprof.Profile))
	app.Post("/debug/pprof/symbol", adaptor.HTTPHandlerFunc(pprof.Symbol))
	app.Get("/debug/pprof/symbol", adaptor.HTTPHandlerFunc(pprof.Symbol))
	app.Get("/debug/pprof/trace", adaptor.HTTPHandlerFunc(pprof.Trace))
	app.Get("/debug/pprof/:profile", func(c *fiber.Ctx) error {
		handler := pprof.Handler(c.Params("profile"))
		return adaptor.HTTPHandler(http.HandlerFunc(handler.ServeHTTP))(c)
	})
}
