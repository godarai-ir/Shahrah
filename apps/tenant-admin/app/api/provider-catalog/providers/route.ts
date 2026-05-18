import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { fetchModelDiscoveryJSON } from "@/lib/model-discovery";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  try {
    const providers = await fetchModelDiscoveryJSON<Array<{
      slug: string;
      name: string;
      enabled: boolean;
      connectionProfile?: { authMode?: string | null } | null;
    }>>("/providers");

    const data = providers.map((provider) => ({
      provider: provider.slug,
      display_name: provider.name,
      enabled: provider.enabled,
      auth_modes: provider.connectionProfile?.authMode ? [provider.connectionProfile.authMode] : ["api_key"],
      supports_bring_your_own_key: true,
      supports_virtual_keys: true,
    }));

    return NextResponse.json({ data }, { status: 200, headers: { "Cache-Control": "public, max-age=300, s-maxage=1800" } });
  } catch (error) {
    return NextResponse.json(
      { error: { message: error instanceof Error ? error.message : "Failed to load providers from model-discovery." } },
      { status: 502 },
    );
  }
}
