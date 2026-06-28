# Blog App

Astro site for the Shahrah marketing and technical blog.

## Purpose

- product announcements
- tutorials
- technical deep-dives
- company updates

## Commands

Run from the repository root:

```bash
pnpm --filter @shahrah/blog dev
pnpm --filter @shahrah/blog build
pnpm --filter @shahrah/blog lint
```

## Content Model

Place articles in `src/content/blog/` with `.md` or `.mdx` format.

Recommended frontmatter:

```md
---
title: "عنوان مقاله"
seoTitle: "SEO title with strong keyword focus"
description: "Short meta description"
excerpt: "Short card summary"
pubDate: 2026-06-28
updatedDate: 2026-06-28
slug: "clean-url-slug"
category: "AI Strategy"
tags: ["LLM Gateway", "Product Management"]
featured: true
draft: false
---
```

Notes:

- `category` drives archive grouping and category pages.
- `slug` keeps article URLs stable and SEO-friendly.
- If `category` is omitted, the first tag is used as a fallback.
- If `slug` is omitted, Astro uses the file name.
