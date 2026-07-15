import type { MetadataRoute } from "next";
import { absoluteUrl } from "@/app/lib/seo";

const staticRoutes = [
  "/",
  "/about",
  "/blog",
  "/contact",
  "/dashboard",
  "/docs",
  "/features/agents",
  "/features/ai-gateway",
  "/features/guardrails",
  "/features/mcp-gateway",
  "/features/model-discovery",
  "/features/observability",
  "/features/prompt-engineering",
  "/features/security",
  "/pricing",
  "/privacy",
  "/signup",
  "/terms",
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
