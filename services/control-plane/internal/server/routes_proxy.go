package server

import "github.com/gofiber/fiber/v2"

func registerProxyRoutes(app *fiber.App, deps Dependencies) {
	v1 := app.Group("/v1")
	if deps.ProxyHandler != nil {
		v1.Post("/chat/completions", deps.ProxyHandler.Handle)
	}

	if deps.CompletionsHandler != nil {
		v1.Post("/completions", deps.CompletionsHandler.Handle)
	}

	if deps.ResponsesHandler != nil {
		v1.Post("/responses", deps.ResponsesHandler.Handle)
		v1.Post("/responses/compact", deps.ResponsesHandler.HandleCompact)
		v1.Post("/responses/input_tokens", deps.ResponsesHandler.HandleInputTokens)
		v1.Get("/responses/:response_id", deps.ResponsesHandler.HandleRetrieve)
		v1.Delete("/responses/:response_id", deps.ResponsesHandler.HandleDelete)
		v1.Post("/responses/:response_id/cancel", deps.ResponsesHandler.HandleCancel)
		v1.Get("/responses/:response_id/input_items", deps.ResponsesHandler.HandleInputItems)
	}

	if deps.EmbeddingsHandler != nil {
		v1.Post("/embeddings", deps.EmbeddingsHandler.Handle)
	}

	if deps.ModerationsHandler != nil {
		v1.Post("/moderations", deps.ModerationsHandler.Handle)
	}

	if deps.RerankHandler != nil {
		v1.Post("/rerank", deps.RerankHandler.Handle)
	}

	if deps.ImagesHandler != nil {
		v1.Post("/images/generations", deps.ImagesHandler.HandleGenerations)
		v1.Post("/images/edits", deps.ImagesHandler.HandleEdits)
		v1.Post("/images/variations", deps.ImagesHandler.HandleVariations)
	}

	if deps.AudioTranscriptionsHandler != nil {
		v1.Post("/audio/transcriptions", deps.AudioTranscriptionsHandler.Handle)
	}

	if deps.AudioSpeechHandler != nil {
		v1.Post("/audio/speech", deps.AudioSpeechHandler.Handle)
	}

	if deps.ModelsHandler != nil {
		v1.Get("/models", func(c *fiber.Ctx) error {
			if deps.AnthropicNativeHandler != nil && c.Get("anthropic-version") != "" {
				return deps.AnthropicNativeHandler.Handle(c)
			}
			return deps.ModelsHandler.Handle(c)
		})
	}

	if deps.AnthropicNativeHandler != nil {
		native := deps.AnthropicNativeHandler.Handle
		v1.Post("/messages", native)
		v1.Post("/messages/count_tokens", native)
		v1.Post("/messages/batches", native)
		v1.Get("/messages/batches", native)
		v1.Get("/messages/batches/:id", native)
		v1.Post("/messages/batches/:id/cancel", native)
		v1.Get("/messages/batches/:id/results", native)

		requireAnthropicVersion := func(c *fiber.Ctx) error {
			if c.Get("anthropic-version") == "" {
				return c.Next()
			}
			return native(c)
		}
		v1.Post("/files", requireAnthropicVersion)
		v1.Get("/files", requireAnthropicVersion)
		v1.Get("/files/:id", requireAnthropicVersion)
		v1.Delete("/files/:id", requireAnthropicVersion)
		v1.Get("/files/:id/content", requireAnthropicVersion)
		v1.Get("/models/:id", requireAnthropicVersion)
	}
}
