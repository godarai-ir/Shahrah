import fs from 'node:fs';
import path from 'node:path';
import {themes as prismThemes} from 'prism-react-renderer';
import type {Config, LoadContext, Plugin} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import {normalizeForSearch} from './src/search/normalize';
import type {SearchData, SearchEntry} from './src/search/types';

const localSearchPluginName = 'shahrah-local-search';

function shahrahLocalSearchPlugin(context: LoadContext): Plugin {
  const docsDirectory = path.join(context.siteDir, 'docs');

  return {
    name: localSearchPluginName,
    getPathsToWatch() {
      return [path.join(docsDirectory, '**/*.{md,mdx}')];
    },
    loadContent() {
      return {
        entries: readSearchEntries(docsDirectory),
      };
    },
    contentLoaded({content, actions}) {
      actions.setGlobalData(content as SearchData);
    },
  };
}

function readSearchEntries(docsDirectory: string): SearchEntry[] {
  if (!fs.existsSync(docsDirectory)) return [];

  return findMarkdownFiles(docsDirectory)
    .map(filePath => createSearchEntry(docsDirectory, filePath))
    .filter((entry): entry is SearchEntry => Boolean(entry))
    .sort((a, b) => a.path.localeCompare(b.path));
}

function findMarkdownFiles(directory: string): string[] {
  const entries = fs.readdirSync(directory, {withFileTypes: true});

  return entries.flatMap(entry => {
    const entryPath = path.join(directory, entry.name);

    if (entry.isDirectory()) return findMarkdownFiles(entryPath);
    if (entry.isFile() && /\.(md|mdx)$/i.test(entry.name)) return [entryPath];

    return [];
  });
}

function createSearchEntry(
  docsDirectory: string,
  filePath: string,
): SearchEntry | null {
  const raw = fs.readFileSync(filePath, 'utf8').replace(/^\uFEFF/, '');
  const {frontMatter, body} = parseFrontMatter(raw);

  if (frontMatter.draft === 'true') return null;

  const headings = extractHeadings(body);
  const content = markdownToPlainText(body);
  const title = frontMatter.title || headings[0] || path.basename(filePath, path.extname(filePath));
  const description = frontMatter.description || createDescription(content);
  const id = createDocId(docsDirectory, filePath);

  return {
    id,
    title,
    path: createDocPath(id, frontMatter.slug),
    description,
    headings,
    content,
    normalized: {
      title: normalizeForSearch(title),
      description: normalizeForSearch(description),
      headings: normalizeForSearch(headings.join(' ')),
      content: normalizeForSearch(content),
    },
  };
}

function parseFrontMatter(raw: string): {
  frontMatter: Record<string, string>;
  body: string;
} {
  const match = raw.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n?/);

  if (!match) {
    return {
      frontMatter: {},
      body: raw,
    };
  }

  const frontMatter = Object.fromEntries(
    match[1]
      .split(/\r?\n/)
      .map(line => line.match(/^([A-Za-z0-9_-]+):\s*(.*)$/))
      .filter((line): line is RegExpMatchArray => Boolean(line))
      .map(line => [line[1], cleanFrontMatterValue(line[2])]),
  );

  return {
    frontMatter,
    body: raw.slice(match[0].length),
  };
}

function cleanFrontMatterValue(value: string): string {
  return value.trim().replace(/^['"]|['"]$/g, '');
}

function extractHeadings(markdown: string): string[] {
  return [...markdown.matchAll(/^#{1,4}\s+(.+)$/gm)].map(match =>
    cleanMarkdownInline(match[1]),
  );
}

function markdownToPlainText(markdown: string): string {
  return cleanMarkdownInline(
    markdown
      .replace(/```[\s\S]*?```/g, ' ')
      .replace(/^import\s+.+$/gm, ' ')
      .replace(/^export\s+.+$/gm, ' ')
      .replace(/<[A-Z][\s\S]*?\/>/g, ' ')
      .replace(/<\/?[A-Za-z][^>]*>/g, ' ')
      .replace(/^#{1,6}\s+/gm, '')
      .replace(/^\s*:::\w+.*$/gm, ' ')
      .replace(/^\s*:::\s*$/gm, ' '),
  );
}

function cleanMarkdownInline(value: string): string {
  return value
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/[`*_>#|]/g, ' ')
    .replace(/\{#[^}]+\}/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

function createDescription(content: string): string {
  const description = content.replace(/\s+/g, ' ').trim();

  if (description.length <= 180) return description;

  return `${description.slice(0, 177).trim()}...`;
}

function createDocId(docsDirectory: string, filePath: string): string {
  return path
    .relative(docsDirectory, filePath)
    .replace(/\\/g, '/')
    .replace(/\.(md|mdx)$/i, '');
}

function createDocPath(id: string, slug?: string): string {
  if (slug) return slug.startsWith('/') ? slug : `/${slug}`;

  const routePath = id.endsWith('/index') ? id.slice(0, -'/index'.length) : id;

  return `/${routePath}`;
}

const config: Config = {
  title: 'مستندات گدارAI',
  tagline: 'مستندات پلتفرم گیت‌وی مدل زبانی گدارAI',
  favicon: 'brand/icons/favicon.ico',
  future: {
    v4: true,
  },
  url: 'https://godarai.ir',
  baseUrl: '/docs/',
  organizationName: 'Delpak',
  projectName: 'shahrah',
  onBrokenLinks: 'throw',
  i18n: {
    defaultLocale: 'fa',
    locales: ['fa'],
    localeConfigs: {
      fa: {
        direction: 'rtl',
        htmlLang: 'fa-IR',
        label: 'فارسی',
      },
    },
  },
  headTags: [
    {
      tagName: 'link',
      attributes: {
        rel: 'icon',
        href: '/docs/brand/icons/favicon.svg',
        type: 'image/svg+xml',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'icon',
        href: '/docs/brand/icons/favicon-light.ico',
        sizes: 'any',
        media: '(prefers-color-scheme: light)',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'icon',
        href: '/docs/brand/icons/favicon-dark.ico',
        sizes: 'any',
        media: '(prefers-color-scheme: dark)',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'apple-touch-icon',
        href: '/docs/brand/icons/apple-touch-icon.png',
        sizes: '180x180',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'mask-icon',
        href: '/docs/brand/svg/safari-pinned-tab.svg',
        color: '#00C9EB',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'manifest',
        href: '/docs/brand/web/site.webmanifest',
      },
    },
    {
      tagName: 'meta',
      attributes: {
        name: 'theme-color',
        content: '#FFFFFF',
        media: '(prefers-color-scheme: light)',
      },
    },
    {
      tagName: 'meta',
      attributes: {
        name: 'theme-color',
        content: '#06182D',
        media: '(prefers-color-scheme: dark)',
      },
    },
    {
      tagName: 'meta',
      attributes: {
        name: 'msapplication-config',
        content: '/docs/brand/web/browserconfig.xml',
      },
    },
    {
      tagName: 'script',
      attributes: {
        type: 'application/ld+json',
      },
      innerHTML: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'Organization',
        name: 'گدارAI',
        alternateName: 'GodarAI',
        url: 'https://godarai.ir',
        logo: 'https://godarai.ir/docs/brand/svg/godarai-mark-light.svg',
      }),
    },
    {
      tagName: 'script',
      attributes: {
        type: 'application/ld+json',
      },
      innerHTML: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'WebSite',
        name: 'مستندات گدارAI',
        url: 'https://godarai.ir/docs/',
        inLanguage: 'fa-IR',
      }),
    },
  ],
  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: '/',
          editUrl: 'https://github.com/Delpak/shahrah/tree/main/apps/docs/',
        },
        blog: false,
        sitemap: {
          changefreq: 'weekly',
          priority: 0.7,
          ignorePatterns: ['/search'],
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],
  plugins: [shahrahLocalSearchPlugin],
  themeConfig: {
    image: 'brand/social/og-image-light-1200x630.png',
    metadata: [
      {
        name: 'keywords',
        content:
          'گدارAI, GodarAI, LLM Gateway, درگاه مدل‌های زبانی, API هوش مصنوعی, مستندات API, OpenAI compatible API',
      },
      {name: 'twitter:card', content: 'summary_large_image'},
      {property: 'og:locale', content: 'fa_IR'},
      {property: 'og:site_name', content: 'مستندات گدارAI'},
      {property: 'og:image:width', content: '1200'},
      {property: 'og:image:height', content: '630'},
      {property: 'og:image:type', content: 'image/png'},
      {property: 'og:image:alt', content: 'لوگوی گدارAI — مستندات'},
      {name: 'twitter:image:alt', content: 'لوگوی گدارAI — مستندات'},
    ],
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'گدارAI',
      logo: {
        alt: 'لوگوی گدارAI',
        src: 'brand/svg/godarai-mark-light.svg',
        srcDark: 'brand/svg/godarai-mark-dark.svg',
        width: 34,
        height: 25,
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'مستندات',
        },
        {
          to: '/search',
          position: 'right',
          label: 'جستجو',
        },
        {
          href: 'https://github.com/Delpak/shahrah',
          label: 'گیت‌هاب',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'مستندات',
          items: [
            {
              label: 'شروع',
              to: '/intro',
            },
          ],
        },
        {
          title: 'پلتفرم',
          items: [
            {
              label: 'سایت اصلی',
              href: 'https://godarai.ir',
            },
          ],
        },
        {
          title: 'لینک‌های کمکی',
          items: [
            {
              label: 'بلاگ',
              href: 'https://godarai.ir/blog',
            },
            {
              label: 'گیت‌هاب',
              href: 'https://github.com/Delpak/shahrah',
            },
          ],
        },
      ],
      copyright: 'GodarAI | 1405 | Built by Developers for Developers',
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;

