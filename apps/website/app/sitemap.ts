import type { MetadataRoute } from "next";
import { absoluteUrl } from "@/app/lib/seo";
import { solutionHref, solutions } from "@/app/lib/solutions";

const staticRoutes = [
  "/",
  "/about",
  "/blog",
  "/contact",
  "/dashboard",
  "/demo",
  "/docs",
  "/features/agents",
  "/features/ai-gateway",
  "/features/guardrails",
  "/features/mcp-gateway",
  "/features/model-discovery",
  "/features/observability",
  "/features/prompt-engineering",
  "/features/security",
  "/on-premise",
  "/pricing",
  "/privacy",
  "/signup",
  "/terms",
  "/vs/litellm",
  "/vs/portkey",
  ...solutions.map((solution) => solutionHref(solution.slug)),
];

export default function sitemap(): MetadataRoute.Sitemap {
  const now = new Date();
  return staticRoutes.map((route) => ({
    url: absoluteUrl(route),
    lastModified: now,
    changeFrequency: route === "/" ? "daily" : "weekly",
    priority: route === "/" ? 1 : 0.7,
  }));
}
