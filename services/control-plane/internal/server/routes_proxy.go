package server

import (
	"encoding/json"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/gofiber/fiber/v2"
)

// registerProxyRoutes registers the gateway routes. Registration order matters
// for Fiber matching, so the helpers below must stay in this sequence; the
// route table is pinned by TestProxyRouteTableIsStable.
func registerProxyRoutes(app *fiber.App, deps Dependencies) {
	v1 := app.Group("/v1")

	registerGeminiNativeRoutes(app, deps)
	registerOpenAICoreRoutes(app, v1, deps)
	registerMediaRoutes(app, v1, deps)
	registerMistralNativeRoutes(v1, deps)
	registerModelCatalogRoutes(v1, deps)
	registerAnthropicRoutes(v1, deps)
	registerFileRoutes(v1, deps)
	registerBudgetRoutes(v1, deps)
}

// registerBudgetRoutes registers the caller's budget status.
func registerBudgetRoutes(v1 fiber.Router, deps Dependencies) {
	if deps.BudgetMeHandler != nil {
		v1.Get("/budget/me", deps.BudgetMeHandler)
	}

	if deps.SeatRequestHandler != nil {
		v1.Post("/seat-requests", deps.SeatRequestHandler)
	}
}

// registerGeminiNativeRoutes registers Gemini-native and TokenBazaar generateContent routes.
func registerGeminiNativeRoutes(app *fiber.App, deps Dependencies) {
	// "/v1beta/models/*" is routed through a thin shim
	// (proxy.NewGeminiGenerateContentRouter): a
	// generateContent/streamGenerateContent request whose {model} path
	// segment is an explicit "tokenbazaar:<account>:<model-id>" FQN is
	// dispatched to TokenBazaarGenerateContentHandler (the narrow,
	// FQN-routed path added for TokenBazaar's native Gemini generateContent
	// capability); every other request -- every direct google-gemini
	// request, and every other Gemini-native model operation (countTokens,
	// embedContent, batchGenerateContent, model get) -- falls back to
	// GeminiNativeHandler.Handle, completely unmodified by that delivery,
	// exactly as before. See
	// plans/tokenbazaar-production-readiness.md, "Native Gemini
	// generateContent/streamGenerateContent delivery", for the full
	// reasoning.
	//
	// This route is registered independently of the direct google-gemini
	// block below (rather than nested inside its GeminiNativeHandler nil
	// guard) so that TokenBazaar's generateContent capability, which does
	// not depend on GeminiNativeHandler at all, stays reachable even in a
	// deployment that omits the direct Gemini provider.
	if deps.GeminiNativeHandler != nil || deps.TokenBazaarGenerateContentHandler != nil {
		var gemini fiber.Handler
		if deps.GeminiNativeHandler != nil {
			gemini = deps.GeminiNativeHandler.Handle
		}

		app.Group("/v1beta").All("/models/*", proxy.NewGeminiGenerateContentRouter(deps.TokenBazaarGenerateContentHandler, gemini))
	}

	if deps.GeminiNativeHandler != nil {
		gemini := deps.GeminiNativeHandler.Handle
		v1beta := app.Group("/v1beta")
		v1beta.All("/models", gemini)
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
}

// registerOpenAICoreRoutes registers chat, completions, responses, batches, embeddings, moderation and rerank routes.
func registerOpenAICoreRoutes(app *fiber.App, v1 fiber.Router, deps Dependencies) {
	if deps.ProxyHandler != nil {
		v1.Post("/chat/completions", deps.ProxyHandler.Handle)
		app.Post("/api/paas/v4/chat/completions", deps.ProxyHandler.Handle)
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
}

// registerMediaRoutes registers image, audio, layout-parsing and OCR routes.
func registerMediaRoutes(app *fiber.App, v1 fiber.Router, deps Dependencies) {
	if deps.VideosHandler != nil {
		v1.Post("/videos", deps.VideosHandler.HandleCreate)
		v1.Get("/videos", deps.VideosHandler.HandleList)
		v1.Post("/videos/generations", deps.VideosHandler.HandleGenerations)
		v1.Post("/videos/edits", deps.VideosHandler.HandleEdits)
		v1.Post("/videos/extensions", deps.VideosHandler.HandleExtensions)
		v1.Get("/videos/:video_id", deps.VideosHandler.HandleGet)
		v1.Delete("/videos/:video_id", deps.VideosHandler.HandleDelete)
		v1.Get("/videos/:video_id/content", deps.VideosHandler.HandleContent)
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

	if deps.ZAILayoutParsingHandler != nil {
		v1.Post("/layout_parsing", deps.ZAILayoutParsingHandler.Handle)
		app.Post("/api/paas/v4/layout_parsing", deps.ZAILayoutParsingHandler.Handle)
	}

	if deps.MistralOCRHandler != nil || deps.ZAILayoutParsingHandler != nil {
		v1.Post("/ocr", func(c *fiber.Ctx) error {
			var request struct {
				Model string `json:"model"`
			}
			if json.Unmarshal(c.Body(), &request) == nil &&
				strings.EqualFold(proxy.ProviderModelIDForRouting(request.Model), "glm-ocr") &&
				deps.ZAILayoutParsingHandler != nil {
				return deps.ZAILayoutParsingHandler.Handle(c)
			}

			if deps.MistralOCRHandler == nil {
				return fiber.ErrNotFound
			}

			return deps.MistralOCRHandler.Handle(c)
		})
	}
}

// registerMistralNativeRoutes registers Mistral-native batch, agent and conversation routes.
func registerMistralNativeRoutes(v1 fiber.Router, deps Dependencies) {
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
}

// registerModelCatalogRoutes registers model listing and xAI catalog routes.
func registerModelCatalogRoutes(v1 fiber.Router, deps Dependencies) {
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
}

// registerAnthropicRoutes registers Anthropic Messages, batch and model detail routes.
func registerAnthropicRoutes(v1 fiber.Router, deps Dependencies) {
	// POST /v1/messages is routed through a thin shim
	// (proxy.NewAnthropicMessagesRouter): a request whose body "model" is an
	// explicit "tokenbazaar:<account>:<model-id>" FQN is dispatched to
	// TokenBazaarMessagesHandler (the narrow, non-account-allowlisted,
	// FQN-routed path added for TokenBazaar's Messages capability); every
	// other request falls back to AnthropicNativeHandler.Handle, completely
	// unmodified by that delivery, exactly as before. See
	// plans/tokenbazaar-production-readiness.md, "Anthropic Messages
	// endpoint delivery", for the full reasoning.
	//
	// This route is registered independently of the direct anthropic block
	// below (rather than nested inside its AnthropicNativeHandler nil guard)
	// so that TokenBazaar's Messages capability, which does not depend on
	// AnthropicNativeHandler at all, stays reachable even in a deployment
	// that omits the direct Anthropic provider.
	if deps.AnthropicNativeHandler != nil || deps.TokenBazaarMessagesHandler != nil {
		var native fiber.Handler
		if deps.AnthropicNativeHandler != nil {
			native = deps.AnthropicNativeHandler.Handle
		}

		v1.Post("/messages", proxy.NewAnthropicMessagesRouter(deps.TokenBazaarMessagesHandler, native))
	}

	if deps.AnthropicNativeHandler != nil {
		native := deps.AnthropicNativeHandler.Handle
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
}

// registerFileRoutes registers the provider-dispatched file routes.
func registerFileRoutes(v1 fiber.Router, deps Dependencies) {
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
