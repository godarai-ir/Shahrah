import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'مستندات گدارAI',
  tagline: 'مستندات پلتفرم گیت‌وی مدل زبانی گدارAI',
  favicon: 'img/favicon.ico',
  future: {
    v4: true,
  },
  url: 'https://sharah.ir',
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
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],
  themeConfig: {
    image: 'img/docusaurus-social-card.jpg',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'گدارAI',
      logo: {
        alt: 'لوگوی گدارAI',
        src: 'img/mini_logo.svg',
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

