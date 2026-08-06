package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func registerProxyRoutes(app *fiber.App, deps Dependencies) {
	v1 := app.Group("/v1")

	if deps.GeminiNativeHandler != nil {
		gemini := deps.GeminiNativeHandler.Handle
		v1beta := app.Group("/v1beta")
		v1beta.All("/models", gemini)
		v1beta.All("/models/*", gemini)
		v1beta.All("/interactions", gemini)
		v1beta.All("/interactions/*", gemini)
		v1beta.All("/files", gemini)
		v1beta.All("/files/*", gemini)
		v1beta.All("/cachedContents", gemini)
		v1beta.All("/cachedContents/*", gemini)
		v1beta.All("/batches", gemini)
		v1beta.All("/batches/*", gemini)
		app.Post("/upload/v1beta/files", gemini)
	}

	if deps.ProxyHandler != nil {
		v1.Post("/chat/completions", deps.ProxyHandler.Handle)
		v1.Get("/chat/deferred-completion/:request_id", deps.ProxyHandler.HandleXAIDeferredGet)
	}

	if deps.CompletionsHandler != nil {
		v1.Post("/completions", deps.CompletionsHandler.Handle)
		v1.Post("/fim/completions", deps.CompletionsHandler.Handle)
	}

	if deps.ResponsesHandler != nil {
		v1.Post("/responses", deps.ResponsesHandler.Handle)
		v1.Post("/responses/compact", deps.ResponsesHandler.HandleCompact)
		v1.Post("/responses/input_tokens", deps.ResponsesHandler.HandleInputTokens)
		v1.Get("/responses/:response_id", deps.ResponsesHandler.HandleRetrieve)
		v1.Delete("/responses/:response_id", deps.ResponsesHandler.HandleDelete)
		v1.Post("/responses/:response_id/cancel", deps.ResponsesHandler.HandleCancel)
		v1.Get("/responses/:response_id/input_items", deps.ResponsesHandler.HandleInputItems)
		v1.Post("/batches", deps.ResponsesHandler.HandleBatchCreate)
		v1.Get("/batches", deps.ResponsesHandler.HandleBatchList)
		v1.Get("/batches/:batch_id", deps.ResponsesHandler.HandleBatchGet)
		v1.Post("/batches/:batch_id/cancel", deps.ResponsesHandler.HandleBatchCancel)
		v1.Get("/batches/:batch_id/requests", deps.ResponsesHandler.HandleXAIBatchRequestsList)
		v1.Post("/batches/:batch_id/requests", deps.ResponsesHandler.HandleXAIBatchRequestsAdd)
		v1.Get("/batches/:batch_id/results", deps.ResponsesHandler.HandleXAIBatchResults)
		v1.Post("/batches/*", deps.ResponsesHandler.HandleXAIBatchCancelColon)
	}

	if deps.EmbeddingsHandler != nil {
		v1.Post("/embeddings", deps.EmbeddingsHandler.Handle)
	}

	if deps.ModerationsHandler != nil {
		v1.Post("/moderations", deps.ModerationsHandler.Handle)
		v1.Post("/chat/moderations", deps.ModerationsHandler.Handle)
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

	if deps.MistralOCRHandler != nil {
		v1.Post("/ocr", deps.MistralOCRHandler.Handle)
	}

	if deps.MistralNativeHandler != nil {
		native := deps.MistralNativeHandler.Handle
		v1.Post("/batch/jobs", native)
		v1.Get("/batch/jobs", native)
		v1.Get("/batch/jobs/:id", native)
		v1.Post("/batch/jobs/:id/cancel", native)
		v1.Delete("/batch/jobs/:id", native)
		v1.Post("/agents", native)
		v1.Get("/agents", native)
		v1.Get("/agents/pages", native)
		v1.Get("/agents/:id", native)
		v1.Patch("/agents/:id", native)
		v1.Delete("/agents/:id", native)
		v1.Post("/conversations", native)
		v1.Get("/conversations", native)
		v1.Get("/conversations/:id", native)
		v1.Post("/conversations/:id", native)
		v1.Delete("/conversations/:id", native)
		v1.Get("/conversations/:id/history", native)
		v1.Get("/conversations/:id/messages", native)
		v1.Post("/conversations/:id/restart", native)
	}

	if deps.ModelsHandler != nil || deps.ResponsesHandler != nil {
		v1.Get("/models", func(c *fiber.Ctx) error {
			if deps.AnthropicNativeHandler != nil && c.Get("anthropic-version") != "" {
				return deps.AnthropicNativeHandler.Handle(c)
			}

			if deps.ResponsesHandler != nil && strings.EqualFold(strings.TrimSpace(c.Get("x-godarai-provider")), "xai") {
				return deps.ResponsesHandler.HandleXAIModelsList(c)
			}

			if deps.ModelsHandler == nil {
				return fiber.ErrNotFound
			}

			return deps.ModelsHandler.Handle(c)
		})
	}

	if deps.ResponsesHandler != nil {
		v1.Get("/language-models", deps.ResponsesHandler.HandleXAILanguageModelsList)
		v1.Get("/language-models/:model_id", deps.ResponsesHandler.HandleXAILanguageModelsGet)
		v1.Get("/image-generation-models", deps.ResponsesHandler.HandleXAIImageGenerationModelsList)
		v1.Get("/image-generation-models/:model_id", deps.ResponsesHandler.HandleXAIImageGenerationModelsGet)
		v1.Post("/files\\:initialize", deps.ResponsesHandler.HandleXAIFileInitialize)
		v1.Post("/files\\:uploadChunks", deps.ResponsesHandler.HandleXAIFileUploadChunks)
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

		v1.Get("/models/:id", func(c *fiber.Ctx) error {
			if c.Get("anthropic-version") != "" {
				return native(c)
			}

			if deps.ResponsesHandler != nil && strings.EqualFold(strings.TrimSpace(c.Get("x-godarai-provider")), "xai") {
				return deps.ResponsesHandler.HandleXAIModelsGet(c)
			}

			return fiber.ErrNotFound
		})
	} else if deps.ResponsesHandler != nil {
		v1.Get("/models/:id", func(c *fiber.Ctx) error {
			if !strings.EqualFold(strings.TrimSpace(c.Get("x-godarai-provider")), "xai") {
				return fiber.ErrNotFound
			}

			return deps.ResponsesHandler.HandleXAIModelsGet(c)
		})
	}

	if deps.ResponsesHandler != nil || deps.AnthropicNativeHandler != nil || deps.MistralNativeHandler != nil {
		fileHandler := func(openAI fiber.Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				if deps.AnthropicNativeHandler != nil && c.Get("anthropic-version") != "" {
					return deps.AnthropicNativeHandler.Handle(c)
				}
				provider := strings.ToLower(strings.TrimSpace(c.Get("x-godarai-provider")))
				fileID := strings.TrimSpace(c.Params("file_id"))
				if deps.MistralNativeHandler != nil &&
					(provider == "mistral" || provider == "mistral-ai" || strings.HasPrefix(fileID, "mfile_gdr_")) {
					return deps.MistralNativeHandler.Handle(c)
				}

				if openAI == nil {
					return fiber.ErrNotFound
				}

				return openAI(c)
			}
		}

		var create, list, get, deleteFile, content fiber.Handler
		if deps.ResponsesHandler != nil {
			create = deps.ResponsesHandler.HandleFileCreate
			list = deps.ResponsesHandler.HandleFileList
			get = deps.ResponsesHandler.HandleFileGet
			deleteFile = deps.ResponsesHandler.HandleFileDelete
			content = deps.ResponsesHandler.HandleFileContent
		}

		v1.Post("/files", fileHandler(create))
		v1.Get("/files", fileHandler(list))
		v1.Get("/files/:file_id", fileHandler(get))

		if deps.ResponsesHandler != nil {
			v1.Put("/files/:file_id", deps.ResponsesHandler.HandleFileUpdate)
		}

		v1.Delete("/files/:file_id", fileHandler(deleteFile))
		v1.Get("/files/:file_id/content", fileHandler(content))
		if deps.MistralNativeHandler != nil {
			v1.Get("/files/:id/url", deps.MistralNativeHandler.Handle)
		}

		if deps.ResponsesHandler != nil {
			v1.Post("/files/:file_id/public-url", deps.ResponsesHandler.HandleFilePublicURLCreate)
			v1.Post("/files/:file_id/public-url/revoke", deps.ResponsesHandler.HandleFilePublicURLRevoke)
		}
	}
}
