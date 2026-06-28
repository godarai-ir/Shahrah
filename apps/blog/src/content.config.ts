import { defineCollection } from "astro:content";
import { glob } from "astro/loaders";
import { z } from "astro/zod";

const blog = defineCollection({
  loader: glob({
    base: "./src/content/blog",
    pattern: "**/*.{md,mdx}",
  }),
  schema: z.object({
    title: z.string(),
    seoTitle: z.string().max(70).optional(),
    description: z.string(),
    excerpt: z.string().max(220).optional(),
    pubDate: z.coerce.date(),
    updatedDate: z.coerce.date().optional(),
    slug: z.string().optional(),
    category: z.string().optional(),
    tags: z.array(z.string()).default([]),
    featured: z.boolean().default(false),
    heroImage: z.string().optional(),
    canonicalUrl: z.string().url().optional(),
    draft: z.boolean().default(false),
  }),
});

export const collections = { blog };
