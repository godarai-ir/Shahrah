import type { ModelCardData } from "@/components/app-ui/ModelCard";

export const mockModelCards: ModelCardData[] = [
  {
    id: "openai/gpt-4.1-mini",
    name: "GPT-4.1 Mini",
    provider: {
      id: "openai",
      name: "OpenAI",
    },
    description:
      "Fast, cost-efficient general-purpose model suitable for chat, structured output, lightweight reasoning, and production workloads.",
    category: ["chat", "vision", "multimodal"],
    status: "stable",
    recommended: true,
    popular: true,
    contextWindow: 1_000_000,
    maxOutputTokens: 32_000,
    pricing: {
      inputPerMillion: 0.15,
      outputPerMillion: 0.6,
      cachedInputPerMillion: 0.03,
      imagePerMillion: 1.2,
    },
    latencyTier: "low",
    qualityTier: "high",
    capabilities: [
      "function_calling",
      "json_output",
      "system_messages",
      "tool_choice",
    ],
    inputModalities: ["text", "image"],
    outputModalities: ["text"],
    invocationModes: ["chat"],
    tradeoffs: ["Lower reasoning depth than frontier models", "Less robust for complex coding agents"],
    metadata: {
      modelId: "openai/gpt-4.1-mini",
      mode: "chat",
      releaseDate: "2026-03-12",
      knowledgeCutoff: "2025-12",
      regions: ["us-east", "eu-west"],
      rateLimitNotes: "Tiered by organization quota",
      license: "Commercial API terms",
    },
  },
  {
    id: "openai/text-embedding-3-large",
    name: "Text Embedding 3 Large",
    provider: {
      id: "openai",
      name: "OpenAI",
    },
    description:
      "High-quality embedding model optimized for semantic search, retrieval pipelines, and multilingual similarity tasks.",
    category: ["embedding"],
    status: "stable",
    contextWindow: 8_192,
    maxOutputTokens: 0,
    pricing: {
      inputPerMillion: 0.13,
    },
    latencyTier: "low",
    qualityTier: "frontier",
    capabilities: ["Text", "Embeddings", "Streaming"],
    inputModalities: ["text"],
    outputModalities: ["vector"],
    invocationModes: ["embedding"],
    tradeoffs: ["Not suitable for generative chat responses"],
    metadata: {
      modelId: "openai/text-embedding-3-large",
      mode: "embedding",
      releaseDate: "2025-11-05",
      regions: ["global"],
      license: "Commercial API terms",
    },
  },
];
