import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Shahrah Docs',
  tagline: 'Documentation for the Shahrah LLM gateway platform',
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
    defaultLocale: 'en',
    locales: ['en'],
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
      title: 'Shahrah',
      logo: {
        alt: 'Shahrah Logo',
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
          href: 'https://github.com/Delpak/shahrah',
          label: 'GitHub',
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
            {
              label: 'Introduction',
              to: '/intro',
            },
          ],
        },
        {
          title: 'Platform',
          items: [
            {
              label: 'Main Site',
              href: 'https://sharah.ir',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'Blog',
              href: 'https://sharah.ir/blog',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/Delpak/shahrah',
            },
          ],
        },
      ],
      copyright: `Copyright (c) ${new Date().getFullYear()} Shahrah. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;

