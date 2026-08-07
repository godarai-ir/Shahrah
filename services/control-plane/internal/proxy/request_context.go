package proxy

import (
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/guardrails"
	"github.com/gofiber/fiber/v2"
)

type chatRequestContextKey struct{}

type chatRequestContext struct {
	virtualModel         *platform.VirtualModel
	virtualModelTargets  []platform.VirtualModelTarget
	virtualChatTargets   map[string]virtualChatTargetPreparation
	virtualModelRouteAt  time.Time
	virtualModelSnapshot bool
	virtualModelActive   bool
	fallbackCount        int
	failureReason        string
	guardrailOverrides   map[string]guardrails.Guardrail
	prompt               promptRequestMetadata
	promptSet            bool
	responseCache        responseCacheRequestMetadata
	responseCacheSet     bool
}

type promptRequestMetadata struct {
	promptID        string
	promptVersionID string
	promptFQN       string
	promptVersion   int
}

type responseCacheRequestMetadata struct {
	status             string
	cacheType          string
	embeddingLatencyMS int64
}

func requestContext(c *fiber.Ctx) *chatRequestContext {
	if c == nil {
		return nil
	}

	if state, ok := c.Locals(chatRequestContextKey{}).(*chatRequestContext); ok && state != nil {
		return state
	}

	state := &chatRequestContext{}
	c.Locals(chatRequestContextKey{}, state)

	return state
}

func existingRequestContext(c *fiber.Ctx) *chatRequestContext {
	if c == nil {
		return nil
	}

	state, ok := c.Locals(chatRequestContextKey{}).(*chatRequestContext)
	if !ok {
		return nil
	}

	return state
}

func setVirtualModelRoutePlan(
	c *fiber.Ctx,
	model *platform.VirtualModel,
	targets []platform.VirtualModelTarget,
	fromSnapshot bool,
	routeStart time.Time,
) {
	state := requestContext(c)
	state.virtualModel = model

	state.virtualModelTargets = append([]platform.VirtualModelTarget(nil), targets...)
	state.virtualModelSnapshot = fromSnapshot
	state.virtualModelRouteAt = routeStart
	state.virtualModelActive = model != nil
}

func setVirtualModelFailure(c *fiber.Ctx, fallbackCount int, failureReason string) {
	state := requestContext(c)
	state.fallbackCount = fallbackCount
	state.failureReason = failureReason
}

func setGuardrailOverrides(c *fiber.Ctx, overrides map[string]guardrails.Guardrail) {
	requestContext(c).guardrailOverrides = overrides
}

func setPromptRequestMetadata(c *fiber.Ctx, metadata promptRequestMetadata) {
	state := requestContext(c)
	state.prompt = metadata
	state.promptSet = true
}

func setResponseCacheRequestMetadata(c *fiber.Ctx, metadata responseCacheRequestMetadata) {
	state := requestContext(c)
	state.responseCache = metadata
	state.responseCacheSet = true
}
