# GodarAI Documentation System - Docusaurus

## Overview

This document describes the documentation system for Shahrah, built using Docusaurus. The documentation will be accessible at `https://sharah.ir/docs/` and serves as the primary knowledge base for developers and technical users.

## 1. Technology Stack

### 1.1 Core Technologies

| Technology | Purpose | Version |
|------------|---------|---------|
| Docusaurus | Static site generator | 3.x |
| React | UI framework | 18.x |
| MDX | Markdown with JSX | 3.x |
| TypeScript | Type safety | 5.x |
| Vercel/Netlify | Hosting (optional) | - |

### 1.2 Why Docusaurus?

1. **Developer-Friendly**: Built for technical documentation
2. **MDX Support**: Write interactive docs with React components
3. **Versioning**: Support multiple versions of documentation
4. **Search**: Built-in Algolia DocSearch integration
5. **i18n**: Multi-language support (useful for Persian translations)
6. **Plugin Ecosystem**: Easy to extend with plugins

## 2. Project Structure

### 2.1 Directory Layout

```
shahrah-docs/
├── docs/                           # Documentation markdown
│   ├── getting-started/
│   │   ├── _category_.json
│   │   ├── introduction.md
│   │   ├── quickstart.md
│   │   └── installation.md
│   ├── core-concepts/
│   │   ├── _category_.json
│   │   ├── tenants.md
│   │   ├── api-keys.md
│   │   ├── models.md
│   │   └── routing.md
│   ├── api-reference/
│   │   ├── _category_.json
│   │   ├── chat-completions.md
│   │   ├── completions.md
│   │   ├── embeddings.md
│   │   └── models.md
│   ├── sdks/
│   │   ├── _category_.json
│   │   ├── typescript.md
│   │   ├── python.md
│   │   └── other-languages.md
│   ├── guides/
│   │   ├── _category_.json
│   │   ├── migration-from-openai.md
│   │   ├── function-calling.md
│   │   ├── streaming.md
│   │   └── error-handling.md
│   ├── billing/
│   │   ├── _category_.json
│   │   ├── wallet.md
│   │   ├── pricing.md
│   │   └── invoices.md
│   └── troubleshooting/
│       ├── _category_.json
│       ├── faq.md
│       └── common-errors.md
├── blog/                           # Blog posts (if enabled)
├── src/
│   ├── components/                 # Custom React components
│   │   ├── ApiExample/
│   │   ├── CodeBlock/
│   │   ├── ModelComparison/
│   │   └── PricingCalculator/
│   ├── css/
│   │   └── custom.css
│   ├── pages/                      # Custom pages
│   │   ├── index.tsx
│   │   └── pricing.tsx
│   └── theme/                      # Theme customization
│       └── Root.tsx
├── static/
│   ├── img/                        # Static images
│   └── assets/                     # Downloadable assets
├── versioned_docs/                 # Versioned documentation
│   ├── version-1.0.0/
│   └── version-1.1.0/
├── versioned_sidebars/
│   ├── version-1.0.0-sidebars.json
│   └── version-1.1.0-sidebars.json
├── docusaurus.config.ts
├── sidebars.ts
├── package.json
└── tsconfig.json
```

## 3. Configuration

### 3.1 Docusaurus Config

```typescript
// docusaurus.config.ts
import { themes as prismThemes } from 'prism-react-renderer'
import type { Config } from '@docusaurus/types'

const config: Config = {
  title: 'GodarAI Documentation',
  tagline: 'Unified LLM Gateway for Iranian Developers',
  favicon: 'img/favicon.ico',
  url: 'https://godarai.ir',
  baseUrl: '/docs/',
  trailingSlash: false,
  
  organizationName: 'shahrah',
  projectName: 'shahrah-docs',
  
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
    // Future: Add Persian locale
    // locales: ['en', 'fa'],
  },
  
  presets: [
    'classic',
    {
      docs: {
        sidebarPath: './sidebars.ts',
        editUrl: 'https://github.com/shahrah/docs/tree/main/',
        routeBasePath: '/',  // Serve docs at /docs/
        versions: {
          current: {
            label: 'Next 🚧',
            path: 'next',
          },
        },
      },
      blog: false,  // Separate blog service
      theme: {
        customCss: './src/css/custom.css',
      },
    },
  ],
  
  themeConfig: {
    image: 'img/shahrah-social-card.png',
    navbar: {
      title: 'Shahrah',
      logo: {
        alt: 'GodarAI Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          to: '/api-reference',
          label: 'API Reference',
          position: 'left',
        },
        {
          href: 'https://godarai.ir',
          label: 'Main Site',
          position: 'right',
        },
        {
          href: 'https://github.com/shahrah',
          label: 'GitHub',
          position: 'right',
        },
        {
          type: 'docsVersionDropdown',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            { label: 'Getting Started', to: '/docs/getting-started/introduction' },
            { label: 'API Reference', to: '/docs/api-reference/chat-completions' },
          ],
        },
        {
          title: 'Community',
          items: [
            { label: 'GitHub Discussions', href: 'https://github.com/shahrah/discussions' },
          ],
        },
        {
          title: 'More',
          items: [
            { label: 'Blog', href: 'https://sharah.ir/blog' },
            { label: 'GitHub', href: 'https://github.com/shahrah' },
          ],
        },
      ],
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'typescript', 'python', 'json', 'yaml'],
    },
    algolia: {
      appId: 'YOUR_ALGOLIA_APP_ID',
      apiKey: 'YOUR_ALGOLIA_SEARCH_KEY',
      indexName: 'shahrah-docs',
      contextualSearch: true,
    },
  },
  
  // Plugins
  plugins: [
    // OpenAPI/Swagger integration
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: 'api',
        docsPlugin: '@docusaurus/plugin-content-docs',
        specPath: './openapi/shahrah-api.yaml',
        outputDir: 'docs/api-reference',
        sidebarOptions: {
          groupPathsBy: 'tag',
        },
      },
    ],
  ],
  
  markdown: {
    mermaid: true,  // Enable Mermaid diagrams
  },
  themes: ['@docusaurus/theme-mermaid'],
}

export default config
```

### 3.2 Sidebars Configuration

```typescript
// sidebars.ts
import type { SidebarsConfig } from '@docusaurus/plugin-content-docs'

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      link: { type: 'generated-index' },
      items: [
        'getting-started/introduction',
        'getting-started/quickstart',
        'getting-started/installation',
        'getting-started/authentication',
      ],
    },
    {
      type: 'category',
      label: 'Core Concepts',
      link: { type: 'generated-index' },
      items: [
        'core-concepts/tenants',
        'core-concepts/api-keys',
        'core-concepts/models',
        'core-concepts/routing',
        'core-concepts/rate-limits',
      ],
    },
    {
      type: 'category',
      label: 'API Reference',
      link: { type: 'generated-index' },
      items: [
        'api-reference/chat-completions',
        'api-reference/completions',
        'api-reference/embeddings',
        'api-reference/images',
        'api-reference/models',
        'api-reference/audio',
        'api-reference/batch',
      ],
    },
    {
      type: 'category',
      label: 'SDKs',
      link: { type: 'generated-index' },
      items: [
        'sdks/typescript',
        'sdks/python',
        'sdks/go',
        'sdks/rust',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      link: { type: 'generated-index' },
      items: [
        'guides/migration-from-openai',
        'guides/function-calling',
        'guides/streaming',
        'guides/error-handling',
        'guides/retry-fallback',
        'guides/caching',
      ],
    },
    {
      type: 'category',
      label: 'Billing & Pricing',
      link: { type: 'generated-index' },
      items: [
        'billing/wallet',
        'billing/pricing',
        'billing/invoices',
        'billing/payment-methods',
      ],
    },
    {
      type: 'category',
      label: 'Troubleshooting',
      link: { type: 'generated-index' },
      items: [
        'troubleshooting/faq',
        'troubleshooting/common-errors',
        'troubleshooting/rate-limits',
      ],
    },
  ],
}

export default sidebars
```

## 4. Custom Components

### 4.1 API Example Component

```tsx
// src/components/ApiExample/index.tsx
import { Tabs, TabItem } from '@theme/Tabs'
import CodeBlock from '@theme/CodeBlock'

interface ApiExampleProps {
  endpoint: string
  method: string
  curl: string
  python: string
  typescript: string
}

export default function ApiExample({
  endpoint,
  method,
  curl,
  python,
  typescript,
}: ApiExampleProps) {
  return (
    <div className="api-example">
      <div className="api-example-header">
        <span className={`method-badge method-${method.toLowerCase()}`}>
          {method}
        </span>
        <code>{endpoint}</code>
      </div>
      
      <Tabs groupId="language" defaultValue="curl" values={[
        { label: 'cURL', value: 'curl' },
        { label: 'Python', value: 'python' },
        { label: 'TypeScript', value: 'typescript' },
      ]}>
        <TabItem value="curl">
          <CodeBlock language="bash">{curl}</CodeBlock>
        </TabItem>
        <TabItem value="python">
          <CodeBlock language="python">{python}</CodeBlock>
        </TabItem>
        <TabItem value="typescript">
          <CodeBlock language="typescript">{typescript}</CodeBlock>
        </TabItem>
      </Tabs>
    </div>
  )
}
```

### 4.2 Model Comparison Table

```tsx
// src/components/ModelComparison/index.tsx
interface ModelInfo {
  name: string
  provider: string
  inputCost: number  // per 1M tokens USD
  outputCost: number
  maxTokens: number
  supportsStreaming: boolean
  supportsFunctionCalling: boolean
}

const models: ModelInfo[] = [
  { name: 'gpt-4o', provider: 'OpenAI', inputCost: 2.5, outputCost: 10, maxTokens: 128000, supportsStreaming: true, supportsFunctionCalling: true },
  { name: 'claude-sonnet-4', provider: 'Anthropic', inputCost: 3, outputCost: 15, maxTokens: 200000, supportsStreaming: true, supportsFunctionCalling: true },
  // More models...
]

export default function ModelComparison() {
  return (
    <table>
      <thead>
        <tr>
          <th>Model</th>
          <th>Provider</th>
          <th>Input Cost</th>
          <th>Output Cost</th>
          <th>Max Tokens</th>
          <th>Streaming</th>
          <th>Function Calling</th>
        </tr>
      </thead>
      <tbody>
        {models.map((model) => (
          <tr key={model.name}>
            <td><code>{model.name}</code></td>
            <td>{model.provider}</td>
            <td>${model.inputCost}/1M</td>
            <td>${model.outputCost}/1M</td>
            <td>{model.maxTokens.toLocaleString()}</td>
            <td>{model.supportsStreaming ? '✅' : '❌'}</td>
            <td>{model.supportsFunctionCalling ? '✅' : '❌'}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
```

### 4.3 Pricing Calculator

```tsx
// src/components/PricingCalculator/index.tsx
import { useState } from 'react'

export default function PricingCalculator() {
  const [inputTokens, setInputTokens] = useState(1000000)
  const [outputTokens, setOutputTokens] = useState(500000)
  const [exchangeRate, setExchangeRate] = useState(60000) // IRR per USD
  
  const inputCostUSD = inputTokens * 2.5 / 1000000
  const outputCostUSD = outputTokens * 10 / 1000000
  const totalCostUSD = inputCostUSD + outputCostUSD
  const totalCostIRR = totalCostUSD * exchangeRate
  
  return (
    <div className="pricing-calculator">
      <h3>Pricing Calculator (GPT-4o)</h3>
      
      <div className="calculator-inputs">
        <label>
          Input Tokens:
          <input
            type="number"
            value={inputTokens}
            onChange={(e) => setInputTokens(Number(e.target.value))}
          />
        </label>
        
        <label>
          Output Tokens:
          <input
            type="number"
            value={outputTokens}
            onChange={(e) => setOutputTokens(Number(e.target.value))}
          />
        </label>
        
        <label>
          Exchange Rate (IRR/USD):
          <input
            type="number"
            value={exchangeRate}
            onChange={(e) => setExchangeRate(Number(e.target.value))}
          />
        </label>
      </div>
      
      <div className="calculator-results">
        <div className="result">
          <span>Input Cost:</span>
          <strong>${inputCostUSD.toFixed(4)}</strong>
        </div>
        <div className="result">
          <span>Output Cost:</span>
          <strong>${outputCostUSD.toFixed(4)}</strong>
        </div>
        <div className="result total">
          <span>Total Cost:</span>
          <strong>${totalCostUSD.toFixed(4)} / {totalCostIRR.toLocaleString()} IRR</strong>
        </div>
      </div>
    </div>
  )
}
```

## 5. Content Guidelines

### 5.1 Documentation Structure Template

```markdown
---
sidebar_label: "Page Title"
sidebar_position: 1
title: "Full Page Title"
description: "Meta description for SEO"
---

# Full Page Title

Brief introduction paragraph explaining what this page covers.

## Prerequisites

- List of things needed before starting

## Step 1: First Step

Explain the first step with code examples:

\`\`\`bash
curl https://your-tenant.sharah.ir/v1/chat/completions \
  -H "Authorization: Bearer sk-your-key" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello"}]}'
\`\`\`

## Step 2: Second Step

More instructions...

## Next Steps

- [Link to related doc 1](./related-1)
- [Link to related doc 2](./related-2)
```

### 5.2 Code Block Standards

```markdown
<!-- Always specify language -->
\`\`\`python
from GodarAI import ShahrahClient

client = ShahrahClient(
    base_url="https://your-tenant.sharah.ir",
    api_key="sk-your-key"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello"}]
)
\`\`\`

<!-- Use diff for showing changes -->
\`\`\`diff
- model: "gpt-3.5-turbo"
+ model: "gpt-4o"
\`\`\`
```

## 6. Deployment

### 6.1 Build & Deploy Process

```yaml
# .github/workflows/deploy-docs.yml
name: Deploy Documentation

on:
  push:
    branches: [main]
    paths: ['docs/**']

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
        run: cd docs && npm ci
        
      - name: Build docs
        run: cd docs && npm run build
        
      - name: Deploy to server
        run: |
          rsync -avz docs/build/ server:/var/www/shahrah/docs/
```

### 6.2 Integration with Main GodarAI Deployment

The docs are served as a subpath of the main domain:

```nginx
# Main nginx config
server {
    server_name sharah.ir;
    
    # Main application
    location / {
        proxy_pass http://shahrah-main:4000;
    }
    
    # Documentation (static files)
    location /docs/ {
        alias /var/www/shahrah/docs/;
        try_files $uri $uri/ /docs/index.html;
    }
}
```

## 7. Search Integration

### 7.1 Algolia DocSearch

```typescript
// docusaurus.config.ts
{
  themeConfig: {
    algolia: {
      appId: 'YOUR_APP_ID',
      apiKey: 'YOUR_SEARCH_ONLY_API_KEY',
      indexName: 'shahrah-docs',
      contextualSearch: true,
      searchParameters: {
        facetFilters: ['language:en'],
      },
    },
  },
}
```

### 7.2 Crawler Configuration

```json
// algolia-crawler-config.json
{
  "index_name": "shahrah-docs",
  "start_urls": ["https://sharah.ir/docs/"],
  "sitemap_urls": ["https://sharah.ir/docs/sitemap.xml"],
  "selectors": {
    "lvl0": {
      "selector": "(//ul[contains(@class,'menu__list')]//a[contains(@class, 'menu__link menu__link--sublist menu__link--active')]/text() | //nav[contains(@class, 'navbar')]//a[contains(@class, 'navbar__link--active')]/text())[last()]",
      "type": "xpath",
      "global": true,
      "default_value": "Documentation"
    },
    "lvl1": "article h1",
    "lvl2": "article h2",
    "lvl3": "article h3",
    "lvl4": "article h4",
    "lvl5": "article h5",
    "text": "article p, article li"
  }
}
```

## 8. Future Enhancements

### 8.1 Persian Language Support

```typescript
// docusaurus.config.ts
{
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'fa'],
    localeConfigs: {
      en: { label: 'English', direction: 'ltr' },
      fa: { label: 'فارسی', direction: 'rtl' },
    },
  },
}
```

### 8.2 Interactive API Playground

- Embed Swagger UI for testing APIs directly from docs
- Allow authenticated users to make test requests
- Show real-time responses

### 8.3 Version Comparison

- Show diffs between API versions
- Migration guides for breaking changes
