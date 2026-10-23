import { Banknote, BookOpen, Briefcase, Code2, Database, FileText, GitBranch, Headphones, KanbanSquare, Landmark, MessageCircle, NotebookText, Scale, ScrollText, UserRound, Users } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { HomeMediaId } from './home-media';
import type { PillarId } from './pillars';

export type TeamTask = { title: string; text: string };
export type TeamTool = { icon: LucideIcon; title: string; text: string };

export type Team = {
  id: 'sales' | 'support' | 'finance' | 'legal' | 'hr' | 'engineering';
  label: string;
  icon: LucideIcon;
  media: HomeMediaId;
  tasks: TeamTask[];
  /** Team page hero: a two-clause headline with a turn, and a two-sentence lead. */
  title: string;
  lead: string;
  /** The hero window: the tool rows the assistant works through, then the result. */
  steps: TeamTool[];
  result: { title: string; text: string };
  tools: TeamTool[];
  pillars: PillarId[];
  seo: string;
};

// One source for the home page tabs and the /teams pages. Every task and tool is a confirmed capability
// (PRODUCT.md); messengers and CRM use the fixed phrasings from docs/site-architecture/04-open-questions.md.
export const teams: Team[] = [
  {
    id: 'sales',
    label: 'فروش',
    icon: Briefcase,
    media: 'teamSales',
    tasks: [
      { title: 'آماده‌سازی جلسه', text: 'خلاصه‌ی مذاکره‌های قبلی و قول‌های داده‌شده را از CRM بیرون می‌کشد.' },
      { title: 'ثبت بعد از تماس', text: 'یادداشت‌ها را مرتب می‌کند و کار بعدی را در CRM ثبت می‌کند.' },
      { title: 'پیگیری مشتری', text: 'برای مشتری‌هایی که مدتی جواب نداده‌اند پیش‌نویس پیام آماده می‌کند.' },
    ],
    title: 'دستیار تیم فروش. با سابقه‌ی کامل هر مشتری.',
    lead: 'پیش از هر جلسه، خلاصه‌ی مذاکره‌ها و قول‌های داده‌شده آماده است. بعد از تماس، کار بعدی در CRM ثبت می‌شود.',
    steps: [
      { icon: Database, title: 'جست‌وجو در CRM', text: 'سابقه‌ی مذاکره‌ها' },
      { icon: ScrollText, title: 'قول‌های داده‌شده', text: 'از یادداشت‌های قبلی' },
      { icon: MessageCircle, title: 'آخرین پیام‌ها', text: 'از پیام‌رسان‌ها' },
    ],
    result: { title: 'خلاصه‌ی پیش از جلسه آماده است', text: 'برای کارشناس فروش' },
    tools: [
      { icon: Database, title: 'CRM از راه MCP', text: 'سابقه‌ی مشتری، فرصت‌ها و کارهای بعدی' },
      { icon: MessageCircle, title: 'پیام‌رسان‌ها', text: 'مانند تلگرام، واتساپ، بله و …' },
    ],
    pillars: ['govern', 'cost'],
    seo: 'دستیار هوش مصنوعی تیم فروش: خلاصه‌ی پیش از جلسه از CRM، ثبت بعد از تماس و پیش‌نویس پیگیری مشتری.',
  },
  {
    id: 'support',
    label: 'پشتیبانی',
    icon: Headphones,
    media: 'teamSupport',
    tasks: [
      { title: 'پیش‌نویس پاسخ', text: 'از روی مستندات و سابقه‌ی خود مشتری جواب تیکت را می‌نویسد.' },
      { title: 'دسته‌بندی پیام‌ها', text: 'پیام‌های رسیده در پیام‌رسان‌ها مانند تلگرام، واتساپ، بله و … را دسته‌بندی می‌کند و فوری‌ها را جلو می‌اندازد.' },
      { title: 'خلاصه‌ی تیکت‌های باز', text: 'تیکت‌های باز هفته را خلاصه می‌کند تا سرپرست تیم ببیند کار کجا مانده است.' },
    ],
    title: 'دستیار تیم پشتیبانی. پاسخ آماده، پیش از اینکه مشتری دوباره بپرسد.',
    lead: 'از روی مستندات و سابقه‌ی خود مشتری پیش‌نویس پاسخ می‌نویسد و پیام‌های فوری را جلو می‌اندازد.',
    steps: [
      { icon: Database, title: 'جست‌وجو در CRM', text: 'سفارش و فاکتور مشتری' },
      { icon: MessageCircle, title: 'تیکت‌های قبلی', text: 'گفتگوهای پیشین همین مشتری' },
      { icon: BookOpen, title: 'پایگاه دانش', text: 'روال‌های پشتیبانی' },
    ],
    result: { title: 'پیش‌نویس پاسخ آماده است', text: 'همکار پشتیبانی می‌خواند و می‌فرستد' },
    tools: [
      { icon: Database, title: 'CRM از راه MCP', text: 'سابقه‌ی سفارش و گفتگوی مشتری' },
      { icon: MessageCircle, title: 'پیام‌رسان‌ها', text: 'مانند تلگرام، واتساپ، بله و …' },
      { icon: BookOpen, title: 'پایگاه دانش', text: 'راهنمای محصول و روال‌ها' },
    ],
    pillars: ['guard', 'observe'],
    seo: 'دستیار هوش مصنوعی تیم پشتیبانی: پیش‌نویس پاسخ از مستندات و سابقه‌ی مشتری، و دسته‌بندی پیام‌ها.',
  },
  {
    id: 'finance',
    label: 'مالی',
    icon: Banknote,
    media: 'teamFinance',
    tasks: [
      { title: 'تطبیق فاکتور', text: 'فاکتورها را با سفارش‌ها مقایسه می‌کند و هر جا نخواند علامت می‌زند.' },
      { title: 'گزارش آخر ماه', text: 'از روی جدول‌های خودتان پیش‌نویس گزارش را آماده می‌کند.' },
      { title: 'کنترل مصرف', text: 'مصرف هوش مصنوعی هر تیم را کنار سقفش نشان می‌دهد.' },
    ],
    title: 'دستیار تیم مالی. فاکتورها خودشان تطبیق نمی‌خورند.',
    lead: 'فاکتورها را با سفارش‌ها مقایسه می‌کند و از روی جدول‌های خودتان پیش‌نویس گزارش آخر ماه را می‌نویسد.',
    steps: [
      { icon: FileText, title: 'فاکتورهای ماه', text: 'از فایل‌های تیم' },
      { icon: Database, title: 'سفارش‌ها', text: 'CRM از راه MCP' },
      { icon: Scale, title: 'تطبیق', text: 'ردیف به ردیف' },
    ],
    result: { title: 'دو مغایرت پیدا شد', text: 'برای بررسی کارشناس مالی علامت خورد' },
    tools: [
      { icon: FileText, title: 'فایل‌ها و پایگاه دانش', text: 'جدول‌ها و فاکتورهای تیم' },
      { icon: Database, title: 'CRM از راه MCP', text: 'سفارش‌ها و مشتری‌ها' },
    ],
    pillars: ['cost', 'govern'],
    seo: 'دستیار هوش مصنوعی تیم مالی: تطبیق فاکتور با سفارش، پیش‌نویس گزارش آخر ماه و کنترل مصرف هوش مصنوعی.',
  },
  {
    id: 'legal',
    label: 'حقوقی',
    icon: Landmark,
    media: 'teamLegal',
    tasks: [
      { title: 'مرور اولیه', text: 'بندهایی را که با سیاست شرکت نمی‌خواند علامت می‌زند و دلیلش را می‌نویسد.' },
      { title: 'مقایسه نسخه‌ها', text: 'تفاوت نسخه‌های مختلف یک قرارداد را فهرست می‌کند.' },
      { title: 'دستورالعمل خودتان', text: 'مرور بر اساس چک‌لیست و قاعده‌های خودتان انجام می‌شود.' },
    ],
    title: 'دستیار تیم حقوقی. با چک‌لیست خودتان.',
    lead: 'بندهایی را که با سیاست شرکت نمی‌خواند علامت می‌زند و تفاوت نسخه‌های یک قرارداد را فهرست می‌کند.',
    steps: [
      { icon: FileText, title: 'قرارداد جدید', text: 'فایل بارگذاری‌شده' },
      { icon: BookOpen, title: 'چک‌لیست شرکت', text: 'از پایگاه دانش تیم' },
      { icon: Scale, title: 'مرور بندها', text: 'بند به بند' },
    ],
    result: { title: 'سه بند علامت خورد', text: 'با دلیل، برای کارشناس حقوقی' },
    tools: [
      { icon: BookOpen, title: 'پایگاه دانش', text: 'چک‌لیست‌ها و قراردادهای نمونه' },
      { icon: FileText, title: 'فایل‌ها', text: 'قراردادهای بارگذاری‌شده' },
    ],
    pillars: ['guard', 'govern'],
    seo: 'دستیار هوش مصنوعی تیم حقوقی: مرور اولیه‌ی قرارداد با چک‌لیست شرکت و مقایسه‌ی نسخه‌ها.',
  },
  {
    id: 'hr',
    label: 'منابع انسانی',
    icon: UserRound,
    media: 'teamHr',
    tasks: [
      { title: 'سؤال‌های تکراری', text: 'درباره‌ی مرخصی، بیمه و ماموریت از روی آیین‌نامه‌ی خودتان جواب می‌دهد.' },
      { title: 'ورود کارمند جدید', text: 'فهرست کارهای هفته‌ی اول را برای هر کارمند آماده می‌کند.' },
      { title: 'دسترسی محدود', text: 'خودتان تعیین می‌کنید هر کس به چه چیزی دسترسی داشته باشد.' },
    ],
    title: 'دستیار منابع انسانی. آیین‌نامه را از بر است.',
    lead: 'به پرسش‌های تکراری درباره‌ی مرخصی، بیمه و ماموریت از روی آیین‌نامه‌ی خودتان جواب می‌دهد و فهرست هفته‌ی اول کارمند جدید را آماده می‌کند.',
    steps: [
      { icon: MessageCircle, title: 'پرسش همکار', text: 'مانده‌ی مرخصی' },
      { icon: BookOpen, title: 'آیین‌نامه', text: 'از پایگاه دانش تیم' },
      { icon: Users, title: 'دسترسی', text: 'فقط اطلاعات عمومی' },
    ],
    result: { title: 'پاسخ با ارجاع به آیین‌نامه', text: 'بدون نیاز به تیکت' },
    tools: [{ icon: BookOpen, title: 'پایگاه دانش', text: 'آیین‌نامه‌ها و راهنماها' }],
    pillars: ['govern', 'guard'],
    seo: 'دستیار هوش مصنوعی منابع انسانی: پاسخ پرسش‌های تکراری از آیین‌نامه و آماده‌سازی ورود کارمند جدید.',
  },
  {
    id: 'engineering',
    label: 'فنی',
    icon: Code2,
    media: 'teamEngineering',
    tasks: [
      { title: 'تسک تا کد', text: 'تسک‌های Jira را با تغییرهای Git و سندهای Notion کنار هم می‌گذارد.' },
      { title: 'ریشه‌یابی خطا', text: 'گزارش رخدادها و مستندات داخلی را می‌گردد و علت احتمالی را پیشنهاد می‌دهد.' },
      { title: 'API سازگار با OpenAI', text: 'اگر خودتان محصول می‌سازید، همان درگاه را مستقیم هم می‌توانید بردارید.' },
    ],
    title: 'دستیار تیم فنی. Jira، Git و Notion کنار هم.',
    lead: 'تسک‌ها را با تغییرهای کد و سندهای تیم کنار هم می‌گذارد و برای ریشه‌یابی خطا، گزارش رخدادها را می‌گردد.',
    steps: [
      { icon: KanbanSquare, title: 'Jira', text: 'تسک باز این هفته' },
      { icon: GitBranch, title: 'Git', text: 'تغییرهای مرتبط' },
      { icon: NotebookText, title: 'Notion', text: 'سند طراحی' },
    ],
    result: { title: 'خلاصه‌ی وضعیت تسک آماده است', text: 'با پیوند به تغییرها و سند' },
    tools: [
      { icon: KanbanSquare, title: 'Jira', text: 'تسک‌ها و پروژه‌ها' },
      { icon: GitBranch, title: 'Git', text: 'تغییرهای کد' },
      { icon: NotebookText, title: 'Notion', text: 'سندهای تیم' },
    ],
    pillars: ['observe', 'cost'],
    seo: 'دستیار هوش مصنوعی تیم فنی: کنار هم گذاشتن Jira، Git و Notion، ریشه‌یابی خطا و درگاه API سازگار با OpenAI.',
  },
];

export function getTeam(id: string): Team | undefined {
  return teams.find((team) => team.id === id);
}
