# Shahrah Blog Service - Technical Blog Platform

## Overview

This document describes the blog service for Shahrah, a technical blog platform accessible at `https://sharah.ir/blog/`. This service is marked as non-critical for the initial launch but is included for future completeness. The blog serves as a channel for announcements, technical deep-dives, tutorials, and company updates.

## 1. Technology Options

### 1.1 Option Comparison

| Option | Pros | Cons | Recommendation |
|--------|------|------|----------------|
| Docusaurus Blog (Built-in) | Integrated with docs, same stack | Limited blog features | ⭐⭐⭐ |
| Ghost | Full-featured, great editor, newsletter | Separate deployment, Node.js heavy | ⭐⭐⭐⭐ |
| WordPress | Mature ecosystem, plugins | Heavy, security concerns | ⭐⭐ |
| Hashnode | Developer-focused, free hosting | Limited customization | ⭐⭐⭐ |
| Astro + MDX | Fast, modern, flexible | Requires more setup | ⭐⭐⭐⭐⭐ |

### 1.2 Recommended: Astro + MDX

**Why Astro?**

1. **Performance**: Astro generates static HTML, extremely fast
2. **Content Collections**: Excellent MDX support with type safety
3. **Flexibility**: Can be deployed anywhere
4. **Developer Experience**: Great DX for technical writers
5. **Integration**: Easy to integrate with main sharah.ir domain

## 2. Architecture

### 2.1 Deployment Options

```mermaid
graph TB
    subgraph "Option A: Subpath Integration"
        A[sharah.ir] --> B[/docs - Docusaurus]
        A --> C[/blog - Astro]
    end
    
    subgraph "Option B: Subdomain"
        D[blog.sharah.ir] --> E[Astro Blog]
    end
    
    subgraph "Option C: Separate Platform"
        F[sharah.ir/blog] -. iframe/API .-> G[Ghost/Hashnode]
    end
```

### 2.2 Recommended: Subpath Integration

```
https://sharah.ir/        -> Main landing page
https://sharah.ir/docs/   -> Docusaurus documentation
https://sharah.ir/blog/   -> Astro blog
```

## 3. Project Structure (Astro)

### 3.1 Directory Layout

```
shahrah-blog/
├── src/
│   ├── content/                    # Content collections
│   │   ├── blog/                   # Blog posts
│   │   │   ├── getting-started-with-shahrah.md
│   │   │   ├── announcing-new-models.md
│   │   │   └── understanding-rate-limits.md
│   │   ├── authors/                # Author profiles
│   │   │   ├── ali.md
│   │   │   └── sara.md
│   │   └── config.ts               # Content collection config
│   ├── components/
│   │   ├── BaseHead.astro          # SEO meta tags
│   │   ├── BlogPostCard.astro      # Post preview card
│   │   ├── BlogPostLayout.astro    # Post page layout
│   │   ├── AuthorCard.astro        # Author info
│   │   ├── TableOfContents.astro   # Auto-generated TOC
│   │   └── NewsletterSignup.astro  # Email subscription
│   ├── layouts/
│   │   └── BaseLayout.astro
│   ├── pages/
│   │   ├── index.astro             # Blog home
│   │   ├── blog/
│   │   │   ├── index.astro         # All posts
│   │   │   └── [slug].astro        # Single post
│   │   ├── rss.xml.ts              # RSS feed
│   │   └── sitemap.ts              # Sitemap
│   ├── styles/
│   │   └── global.css
│   └── utils/
│       └── formatDate.ts
├── public/
│   └── assets/
│       └── images/
├── astro.config.mjs
├── package.json
└── tsconfig.json
```

### 3.2 Content Collection Configuration

```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content'

const blog = defineCollection({
  type: 'content',
  schema: z.object({
    title: z.string(),
    description: z.string(),
    pubDate: z.date(),
    updatedDate: z.date().optional(),
    author: z.string(),  // Reference to authors collection
    tags: z.array(z.string()).default([]),
    category: z.enum([
      'announcement',
      'tutorial',
      'deep-dive',
      'case-study',
      'opinion',
    ]),
    coverImage: z.string().optional(),
    draft: z.boolean().default(false),
    readingTime: z.number().optional(),  // Auto-calculated
  }),
})

const authors = defineCollection({
  type: 'content',
  schema: z.object({
    name: z.string(),
    role: z.string(),
    avatar: z.string(),
    bio: z.string(),
    twitter: z.string().optional(),
    github: z.string().optional(),
    linkedin: z.string().optional(),
  }),
})

export const collections = { blog, authors }
```

## 4. Blog Post Template

### 4.1 Markdown Frontmatter

```markdown
---
title: "Getting Started with Shahrah: A Complete Guide"
description: "Learn how to set up and use Shahrah for your LLM integration needs."
pubDate: 2024-01-15
author: ali
tags: ["tutorial", "getting-started", "api"]
category: "tutorial"
coverImage: "/assets/images/getting-started-cover.png"
draft: false
---

# Getting Started with Shahrah: A Complete Guide

Welcome to Shahrah! In this comprehensive guide, we'll walk you through...

## Prerequisites

Before you begin, make sure you have:

- A Shahrah account
- An API key from your tenant dashboard
- Node.js 18+ or Python 3.10+

## Step 1: Create Your Tenant

Navigate to [sharah.ir](https://sharah.ir) and sign up...

\`\`\`bash
# Install the Shahrah SDK
npm install @shahrah/sdk
\`\`\`

## Step 2: Make Your First Request

\`\`\`typescript
import { ShahrahClient } from '@shahrah/sdk'

const client = new ShahrahClient({
  baseUrl: 'https://your-tenant.sharah.ir',
  apiKey: process.env.SHAHRAH_API_KEY,
})

const response = await client.chat.completions.create({
  model: 'gpt-4o',
  messages: [{ role: 'user', content: 'Hello, Shahrah!' }],
})

console.log(response.choices[0].message.content)
\`\`\`

## Conclusion

You're now ready to use Shahrah for your LLM needs!

## Next Steps

- Read our [API Reference](/docs/api-reference)
- Join our [Community](https://github.com/shahrah/discussions)
```

## 5. Key Features

### 5.1 SEO & Meta Tags

```astro
---
// src/components/BaseHead.astro
interface Props {
  title: string
  description: string
  image?: string
  url: string
}

const { title, description, image, url } = Astro.props
---

<head>
  <title>{title} | Shahrah Blog</title>
  <meta name="description" content={description} />
  <meta property="og:title" content={title} />
  <meta property="og:description" content={description} />
  {image && <meta property="og:image" content={image} />}
  <meta property="og:url" content={url} />
  <meta name="twitter:card" content="summary_large_image" />
  <link rel="canonical" href={url} />
</head>
```

### 5.2 RSS Feed

```typescript
// src/pages/rss.xml.ts
import rss from '@astrojs/rss'
import { getCollection } from 'astro:content'

export async function GET(context: any) {
  const posts = await getCollection('blog')
  
  return rss({
    title: 'Shahrah Blog',
    description: 'Technical blog posts about LLM integration, AI, and more',
    site: context.site,
    items: posts
      .filter((post) => !post.data.draft)
      .map((post) => ({
        title: post.data.title,
        pubDate: post.data.pubDate,
        description: post.data.description,
        link: `/blog/${post.slug}/`,
      })),
    customData: `<language>en-us</language>`,
  })
}
```

### 5.3 Newsletter Integration

```astro
---
// src/components/NewsletterSignup.astro
---

<div class="newsletter-signup">
  <h3>Stay Updated</h3>
  <p>Get the latest Shahrah news and tutorials delivered to your inbox.</p>
  
  <form action="https://email-service.com/subscribe" method="POST">
    <input 
      type="email" 
      name="email" 
      placeholder="your@email.com" 
      required 
    />
    <button type="submit">Subscribe</button>
  </form>
</div>

<style>
  .newsletter-signup {
    padding: 2rem;
    background: var(--color-surface);
    border-radius: 8px;
    margin: 2rem 0;
  }
</style>
```

## 6. Styling & Design

### 6.1 Design Tokens

```css
/* src/styles/global.css */
:root {
  --color-primary: #2563eb;
  --color-primary-dark: #1d4ed8;
  --color-text: #1f2937;
  --color-text-secondary: #6b7280;
  --color-surface: #f9fafb;
  --color-border: #e5e7eb;
  
  --font-sans: 'Inter', system-ui, sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
  
  --spacing-xs: 0.5rem;
  --spacing-sm: 1rem;
  --spacing-md: 1.5rem;
  --spacing-lg: 2rem;
  --spacing-xl: 3rem;
  
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 16px;
}
```

### 6.2 Blog Post Layout

```astro
---
// src/components/BlogPostLayout.astro
interface Props {
  title: string
  pubDate: Date
  author: any
  tags: string[]
  readingTime?: number
}

const { title, pubDate, author, tags, readingTime } = Astro.props
---

<article class="blog-post">
  <header>
    <h1>{title}</h1>
    <div class="meta">
      <time datetime={pubDate.toISOString()}>
        {pubDate.toLocaleDateString('en-US', { 
          year: 'numeric', 
          month: 'long', 
          day: 'numeric' 
        })}
      </time>
      <span class="author">{author.name}</span>
      {readingTime && <span>{readingTime} min read</span>}
    </div>
    {tags.length > 0 && (
      <div class="tags">
        {tags.map((tag) => <a href={`/blog/tags/${tag}/`}>#{tag}</a>)}
      </div>
    )}
  </header>
  
  <div class="content">
    <slot />
  </div>
  
  <footer>
    <AuthorCard author={author} />
    <NewsletterSignup />
  </footer>
</article>
```

## 7. Integration with Main Site

### 7.1 Nginx Configuration

```nginx
server {
    server_name sharah.ir;
    
    # Main application
    location / {
        proxy_pass http://shahrah-main:4000;
    }
    
    # Documentation
    location /docs/ {
        alias /var/www/shahrah/docs/;
        try_files $uri $uri/ /docs/index.html;
    }
    
    # Blog (static Astro build)
    location /blog/ {
        alias /var/www/shahrah/blog/;
        try_files $uri $uri/ /blog/index.html;
    }
    
    # API for blog comments (optional)
    location /blog/api/comments {
        proxy_pass http://shahrah-blog-api:3000;
    }
}
```

### 7.2 Shared Header/Footer

If using a shared layout across main site, docs, and blog:

```typescript
// Shared component via API or CDN
// Fetch header config from central config service
const headerConfig = await fetch('https://sharah.ir/api/shared-header')
```

## 8. Comments System

### 8.1 Option: Giscus (GitHub Discussions)

```astro
<!-- src/components/Comments.astro -->
<script is:inline>
  const giscusContainer = document.getElementById('giscus')
  const script = document.createElement('script')
  script.src = 'https://giscus.app/client.js'
  script.dataset.repo = 'shahrah/blog-comments'
  script.dataset.repoId = 'YOUR_REPO_ID'
  script.dataset.category = 'Comments'
  script.dataset.categoryId = 'YOUR_CATEGORY_ID'
  script.dataset.mapping = 'pathname'
  script.dataset.strict = '1'
  script.dataset.reactionsEnabled = '1'
  script.dataset.emitMetadata = '0'
  script.dataset.inputPosition = 'bottom'
  script.dataset.theme = 'light'
  script.crossOrigin = 'anonymous'
  script.async = true
  giscusContainer.appendChild(script)
</script>

<div id="giscus" />
```

### 8.2 Option: Custom Comments (Future)

- Build custom comment system with Shahrah backend
- Store comments in separate database
- Support moderation, threading, notifications

## 9. Deployment

### 9.1 Build Process

```json
// package.json
{
  "scripts": {
    "dev": "astro dev",
    "build": "astro build",
    "preview": "astro preview",
    "sync": "astro sync"
  }
}
```

### 9.2 CI/CD Pipeline

```yaml
# .github/workflows/deploy-blog.yml
name: Deploy Blog

on:
  push:
    branches: [main]
    paths: ['blog/**']
  schedule:
    - cron: '0 0 * * *'  # Daily rebuild

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '22'
          
      - name: Install dependencies
        run: cd blog && npm ci
        
      - name: Build blog
        run: cd blog && npm run build
        
      - name: Deploy to server
        run: |
          rsync -avz blog/dist/ server:/var/www/shahrah/blog/
```

## 10. Content Workflow

### 10.1 Writing Process

1. Create new markdown file in `src/content/blog/`
2. Add frontmatter with metadata
3. Write content using markdown + MDX
4. Preview locally with `npm run dev`
5. Submit PR for review
6. Merge triggers deployment

### 10.2 Content Types

| Type | Purpose | Example |
|------|---------|---------|
| Announcement | Product updates | "Announcing Shahrah v2.0" |
| Tutorial | How-to guides | "Function Calling with Shahrah" |
| Deep-Dive | Technical analysis | "Understanding the Silk Road Tunnel" |
| Case Study | Customer stories | "How Company X Reduced LLM Costs by 40%" |
| Opinion | Industry perspectives | "The Future of LLM Proxies" |

## 11. Analytics & Monitoring

### 11.1 Page Views

```typescript
// Add to BaseHead.astro
<script is:inline>
  // Simple analytics (Plausible, Umami, etc.)
  const script = document.createElement('script')
  script.defer = true
  script.src = 'https://analytics.sharah.ir/js/script.js'
  script.dataset.website = 'sharah.ir'
  document.head.appendChild(script)
</script>
```

### 11.2 Popular Posts

- Track post views via server-side or client-side analytics
- Display "Popular Posts" section on blog home
- Weekly/monthly trending posts

## 12. Future Enhancements

### 12.1 Persian Language Support

- RTL layout for Persian posts
- Multi-language content collections
- Persian date formatting

### 12.2 Interactive Demos

- Embed live API demos in blog posts
- Interactive code editors
- Real-time response viewers

### 12.3 Video Content

- Embed video tutorials
- YouTube/Vimeo integration
- Video transcriptions
