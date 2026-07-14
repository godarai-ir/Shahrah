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

	if deps.ModelsHandler != nil {
		v1.Get("/models", deps.ModelsHandler.Handle)
	}
}
