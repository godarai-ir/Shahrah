import { relayPaymentCallback } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{ provider: string }>;
};

/** The gateway sends the buyer back here; the control plane verifies the payment and completes the order. */
export async function GET(request: Request, context: RouteContext) {
  return relayPaymentCallback(request, "seats", (await context.params).provider);
}
