import { MediaSlot } from '../MediaSlot';
import { Reveal } from '../Reveal';
import { CircleLink, sectionPad } from './parts';
import type { HomeMediaId } from '@/app/lib/home-media';

type ValueCard = {
  title: string;
  text: string;
  href: string;
  media: HomeMediaId;
  tone: string;
  light: boolean;
};

const rows: { columns: string; cards: ValueCard[] }[] = [
  {
    columns: 'md:grid-cols-[2fr_1fr]',
    cards: [
      {
        title: 'مدل را خودتان انتخاب می‌کنید',
        text: 'بیش از ۱۲۰۰ مدل از ۲۱ ارائه‌دهنده. به هیچ مدل و ارائه‌دهنده‌ای بسته نیستید.',
        href: '/features/model-discovery',
        media: 'valModels',
        tone: 'bg-sky-200',
        light: false,
      },
      {
        title: 'امنیت از روز اول',
        text: 'گاردریل‌ها داده‌ی حساس را قبل از رسیدن به مدل می‌گیرند و هر درخواست ثبت می‌شود.',
        href: '/features/guardrails',
        media: 'valSecurity',
        tone: 'bg-stone-800',
        light: true,
      },
    ],
  },
  {
    columns: 'md:grid-cols-[1fr_2fr]',
    cards: [
      {
        title: 'هر تیم دستیار خودش را می‌سازد',
        text: 'دستورالعمل، اتصال‌ها و دسترسی‌ها را برای هر تیم جدا تنظیم می‌کنید.',
        href: '/features/agents',
        media: 'valCustom',
        tone: 'bg-amber-200',
        light: false,
      },
      {
        title: 'آماده‌ی سازمان',
        text: 'سقف مصرف هر تیم، کیف پول و گزارش مالی تومانی، و کنترل اینکه چه کسی به چه چیزی دسترسی دارد.',
        href: '/pricing',
        media: 'valEnterprise',
        tone: 'bg-emerald-200',
        light: false,
      },
    ],
  },
  {
    columns: 'md:grid-cols-2',
    cards: [
      {
        title: 'هر جا که لازم است اجرا می‌شود',
        text: 'روی ابر گدارAI یا داخل سرور خودتان، برای داده‌ای که نباید از شبکه بیرون برود.',
        href: '/on-premise',
        media: 'valDeploy',
        tone: 'bg-rose-200',
        light: false,
      },
      {
        title: 'به ابزارهای شما وصل است',
        text: 'CRM، بله، ایتا، واتساپ، تلگرام، Jira، Notion، Git و سرورهای MCP.',
        href: '/features/mcp-gateway',
        media: 'valConnect',
        tone: 'bg-violet-200',
        light: false,
      },
    ],
  },
];

export function WorkplaceValues() {
  return (
    <section>
      <div className={`${sectionPad} pb-20 md:pb-24 lg:pb-40`}>
        <div className="mx-auto w-full max-w-7xl">
          <Reveal className="mx-auto mb-8 w-full max-w-lg md:mb-12 lg:mb-16">
            <h2 className="text-center text-2xl font-medium leading-snug text-zinc-900 dark:text-white md:text-4xl lg:leading-tight">
              هر چیزی که برای پذیرش هوش مصنوعی در سازمان لازم است، یک‌جا.
            </h2>
          </Reveal>

          <div className="flex flex-col gap-8 lg:gap-10">
            {rows.map((row) => (
              <div key={row.columns} className={`grid grid-cols-1 gap-8 lg:gap-10 ${row.columns}`}>
                {row.cards.map((card) => (
                  <Reveal key={card.title} className="flex">
                    <article
                      className={`relative flex min-h-100 w-full flex-col overflow-hidden rounded-xl p-8 sm:p-10 md:h-128 ${card.tone} ${
                        card.light ? 'text-white' : 'text-zinc-900'
                      }`}
                    >
                      <h3 className="mb-2 text-lg font-medium">{card.title}</h3>
                      <p className="mb-6 max-w-md text-base leading-8 opacity-75">{card.text}</p>
                      <div className="mb-6">
                        <CircleLink href={card.href} tone={card.light ? 'light' : 'ink'}>
                          بیشتر بدانید
                        </CircleLink>
                      </div>
                      <div className="min-h-0 flex-1">
                        <MediaSlot id={card.media} className="h-full" />
                      </div>
                    </article>
                  </Reveal>
                ))}
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
