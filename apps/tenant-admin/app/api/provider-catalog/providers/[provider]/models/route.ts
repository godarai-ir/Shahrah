import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { fetchModelDiscoveryJSON } from "@/lib/model-discovery";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    provider: string;
  }>;
};

export async function GET(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const { provider } = await context.params;
  const normalizedProvider = provider.trim().toLowerCase();
  if (!normalizedProvider) {
    return NextResponse.json(
      { error: { message: "provider معتبر نیست." } },
      { status: 400 },
    );
  }

  try {
    const response = await fetchModelDiscoveryJSON<{
      data?: Array<{
        canonicalId: string;
        name?: string;
        contextWindow?: number;
        inputModalities?: string[];
        outputModalities?: string[];
        capabilities?: string[];
        pricing?: {
          inputPerMillionTokens?: number;
          outputPerMillionTokens?: number;
        } | null;
      }>;
    }>(`/models?provider=${encodeURIComponent(normalizedProvider)}&limit=100`);

    const data = (response.data ?? []).map((model) => ({
      provider: normalizedProvider,
      model: model.canonicalId,
      display_name: model.name ?? model.canonicalId,
      types: (model.capabilities ?? []).filter(Boolean),
      context_window: model.contextWindow,
      input_modalities: (model.inputModalities ?? []).filter(Boolean),
      output_modalities: (model.outputModalities ?? []).filter(Boolean),
      features: (model.capabilities ?? []).filter(Boolean),
      public_pricing: model.pricing
        ? {
            input_usd_per_1m: model.pricing.inputPerMillionTokens,
            output_usd_per_1m: model.pricing.outputPerMillionTokens,
          }
        : undefined,
    }));

    return NextResponse.json(
      { data },
      { status: 200, headers: { "Cache-Control": "public, max-age=300, s-maxage=1800" } },
    );
  } catch (error) {
    return NextResponse.json(
      {
        error: {
          message:
            error instanceof Error
              ? error.message
              : "Failed to load provider models from model-discovery.",
        },
      },
      { status: 502 },
    );
  }
}
