> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Anthropic Stream Overload Fallback

> How TrueFoundry AI Gateway handles Anthropic overloaded_error in streaming with automatic fallback.

When Anthropic is overloaded, the provider can return an `overloaded_error` in streaming flows. A common failure mode is receiving HTTP `200` but an error payload in the early stream events.

## TrueFoundry behavior

For streaming requests routed through routing rules/virtual models, TrueFoundry AI Gateway waits until it sees the first non-empty stream chunk.

* If that first meaningful chunk indicates an Anthropic `overloaded_error`, the gateway marks the attempt as failed and falls back to the next eligible target.
* If a normal first meaningful chunk arrives, streaming continues on the same target.

This helps avoid "successful but unusable" first responses during Anthropic overload events.

## Why this matters

In streaming mode, some provider-side overload failures can appear as an error payload in early stream events instead of a non-2xx HTTP status. The first-chunk check ensures these requests fail fast and route to the next eligible target.

## Practical recommendation

* If you use Anthropic with streaming in TrueFoundry, configure fallback targets in your [Routing Config](/docs/ai-gateway/load-balancing-overview) or [Virtual Models](/docs/ai-gateway/virtual-model#retries-and-fallbacks).
* Keep fallback chains short and provider-diverse (for example Anthropic primary, OpenAI or Bedrock secondary).
* Continue using retries and rate limits to reduce pressure during provider-side overload windows.
