import {
  BookOpen,
  Briefcase,
  Code2,
  FileStack,
  Headphones,
  Landmark,
  LayoutGrid,
  Library,
  MessagesSquare,
  Network,
  Newspaper,
  Plug,
  Scale,
  Server,
  ShieldHalf,
  SquareTerminal,
  TerminalSquare,
  UserRound,
  Wallet,
  Database,
  Bot,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { pillars } from './pillars';
import type { PillarId } from './pillars';
import { solutionGroups, solutionHref, solutions } from './solutions';

export type NavLink = {
  label: string;
  href: string;
  description?: string;
  icon?: LucideIcon;
  /** Served by another app (docs, catalog, blog), so it needs a full page load rather than client navigation. */
  external?: boolean;
  /** False while the page does not exist yet; hidden everywhere until its page prompt turns it on. */
  ready?: boolean;
  /** Draws the icon well on this differentiator's tone. */
  pillar?: PillarId;
};

export type NavColumn = { title?: string; links: NavLink[] };

export type NavMenu = {
  id: string;
  label: string;
  /** Columns of the desktop mega menu; the mobile sheet lists them in order. */
  columns: NavColumn[];
};

// Pages that do not exist yet. Each page prompt removes its route from this list.
const pendingRoutes = new Set<string>([]);

const link = (item: NavLink): NavLink => ({ ...item, ready: !pendingRoutes.has(item.href) });

/** Whether a route's page exists yet, so content can link to it without producing a 404. */
export function isReady(href: string): boolean {
  return !pendingRoutes.has(href);
}

export function visibleLinks(links: NavLink[]): NavLink[] {
  return links.filter((item) => item.ready !== false);
}

export function visibleColumns(columns: NavColumn[]): NavColumn[] {
  return columns
    .map((column) => ({ ...column, links: visibleLinks(column.links) }))
    .filter((column) => column.links.length > 0);
}

export const teamLinks: NavLink[] = [
  { label: 'فروش', href: '/teams/sales', icon: Briefcase },
  { label: 'پشتیبانی', href: '/teams/support', icon: Headphones },
  { label: 'مالی', href: '/teams/finance', icon: Wallet },
  { label: 'حقوقی', href: '/teams/legal', icon: Landmark },
  { label: 'منابع انسانی', href: '/teams/hr', icon: UserRound },
  { label: 'فنی', href: '/teams/engineering', icon: Code2 },
].map(link);

const pillarLinks: NavLink[] = pillars.map((pillar) =>
  link({ label: pillar.title, href: pillar.href, description: pillar.summary, icon: pillar.icon, pillar: pillar.id }),
);

const solutionColumns: NavColumn[] = solutionGroups.map((group) => ({
  title: group.label,
  links: solutions
    .filter((solution) => solution.group === group.id)
    .map((solution) =>
      link({
        label: solution.navName,
        href: solutionHref(solution.slug),
        description: solution.navDescription,
        icon: solution.icon,
      }),
    ),
}));

export const navMenus: NavMenu[] = [
  {
    id: 'workplace',
    label: 'فضای کاری',
    columns: [
      {
        links: [
          link({ label: 'نمای کلی', href: '/workplace', description: 'گفتگو، دستیار و اتصال برای همه‌ی تیم‌ها', icon: LayoutGrid }),
          link({ label: 'دستیارها و دانش', href: '/features/agents', description: 'برای هر تیم دستیار خودش', icon: Bot }),
          link({ label: 'اتصال‌ها', href: '/integrations', description: 'MCP، پیام‌رسان‌ها و ابزارهای تیم', icon: Plug }),
        ],
      },
      { title: 'برای تیم‌ها', links: teamLinks },
    ],
  },
  {
    id: 'control',
    label: 'کنترل',
    columns: [
      { links: pillarLinks },
      {
        links: [
          link({ label: 'امنیت و اعتماد', href: '/features/security', description: 'مسیر داده، محل اجرا و کنترل دسترسی', icon: ShieldHalf }),
          link({ label: 'استقرار در سازمان', href: '/on-premise', description: 'گدارAI روی سرورهای خودتان', icon: Server }),
        ],
      },
    ],
  },
  {
    id: 'gateway',
    label: 'درگاه API',
    columns: [
      {
        links: [
          link({ label: 'درگاه مدل‌های زبانی', href: '/features/ai-gateway', description: 'یک نشانی برای ۱۲۰۰+ مدل از ۲۱ ارائه‌دهنده', icon: Network }),
          link({ label: 'کشف مدل‌ها', href: '/features/model-discovery', description: 'جست‌وجو و مقایسه‌ی مدل‌ها', icon: Database }),
          link({ label: 'رجیستری پرامپت', href: '/features/prompt-engineering', description: 'ساخت، نسخه‌بندی و اجرای پرامپت', icon: TerminalSquare }),
          link({ label: 'درگاه MCP', href: '/features/mcp-gateway', description: 'یک مسیر برای سرورهای MCP', icon: SquareTerminal }),
        ],
      },
      {
        title: 'منابع',
        links: [
          link({ label: 'کاتالوگ مدل‌ها', href: '/catalog', icon: Library, external: true }),
          link({ label: 'مستندات', href: '/docs', icon: BookOpen, external: true }),
          link({ label: 'بلاگ', href: '/blog', icon: Newspaper, external: true }),
          link({ label: 'گدارAI و LiteLLM', href: '/vs/litellm', icon: Scale }),
          link({ label: 'گدارAI و Portkey', href: '/vs/portkey', icon: Scale }),
        ],
      },
    ],
  },
  {
    id: 'solutions',
    label: 'راهکارها',
    columns: [
      ...solutionColumns,
      {
        links: [
          link({
            label: 'گدارAI یا ChatGPT کنترل‌نشده',
            href: '/vs/chatgpt',
            description: 'جایگزین سازمانی برای استفاده‌ی پراکنده',
            icon: MessagesSquare,
          }),
        ],
      },
    ],
  },
];

export const pricingLink: NavLink = link({ label: 'قیمت‌گذاری', href: '/pricing', icon: FileStack });

export const footerColumns: NavColumn[] = [
  {
    title: 'فضای کاری',
    links: [
      link({ label: 'نمای کلی', href: '/workplace' }),
      link({ label: 'دستیارها و دانش', href: '/features/agents' }),
      link({ label: 'اتصال‌ها', href: '/integrations' }),
      link({ label: 'قیمت‌گذاری', href: '/pricing' }),
    ],
  },
  {
    title: 'کنترل',
    links: [
      ...pillars.map((pillar) => link({ label: pillar.title, href: pillar.href })),
      link({ label: 'امنیت و اعتماد', href: '/features/security' }),
      link({ label: 'استقرار در سازمان', href: '/on-premise' }),
    ],
  },
  {
    title: 'درگاه API',
    links: [
      link({ label: 'درگاه مدل‌های زبانی', href: '/features/ai-gateway' }),
      link({ label: 'کشف مدل‌ها', href: '/features/model-discovery' }),
      link({ label: 'رجیستری پرامپت', href: '/features/prompt-engineering' }),
      link({ label: 'درگاه MCP', href: '/features/mcp-gateway' }),
      link({ label: 'داشبورد', href: '/dashboard' }),
      link({ label: 'کاتالوگ مدل‌ها', href: '/catalog', external: true }),
      link({ label: 'مستندات', href: '/docs', external: true }),
      link({ label: 'گدارAI و LiteLLM', href: '/vs/litellm' }),
      link({ label: 'گدارAI و Portkey', href: '/vs/portkey' }),
    ],
  },
  {
    title: 'راهکارها',
    links: [
      ...teamLinks.map((team) => ({ ...team, label: `تیم ${team.label}`, icon: undefined })),
      ...solutions.map((solution) => link({ label: solution.navName, href: solutionHref(solution.slug) })),
      link({ label: 'گدارAI یا ChatGPT', href: '/vs/chatgpt' }),
    ],
  },
  {
    title: 'شرکت',
    links: [
      link({ label: 'درباره‌ی ما', href: '/about' }),
      link({ label: 'تماس با ما', href: '/contact' }),
      link({ label: 'درخواست دمو', href: '/demo' }),
      link({ label: 'بلاگ', href: '/blog', external: true }),
      link({ label: 'قوانین و مقررات', href: '/terms' }),
      link({ label: 'حریم خصوصی', href: '/privacy' }),
    ],
  },
];
