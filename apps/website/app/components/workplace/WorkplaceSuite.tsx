import { MediaSlot } from '../MediaSlot';
import { Reveal } from '../Reveal';
import { Carousel } from './Carousel';
import { AgentRunMock, ChatMock, ConnectionsMock } from './mocks';
import { CircleLink, sectionPad, TwoTone } from './parts';
import type { HomeMediaId } from '@/app/lib/home-media';
import type { ReactNode } from 'react';

type SuiteCard = {
  title: string;
  text: string;
  href: string;
  media: HomeMediaId;
  tone: string;
  light: boolean;
  mock?: ReactNode;
};

const cards: SuiteCard[] = [
  {
    title: 'گفتگو',
    text: 'فضای گفتگوی سازمان با هر مدل. برای همه‌ی همکارها.',
    href: '/features/ai-gateway',
    media: 'suiteChat',
    tone: 'bg-stone-800',
    light: true,
    mock: <ChatMock />,
  },
  {
    title: 'دستیارها',
    text: 'برای هر تیم دستیار خودش را بسازید.',
    href: '/features/agents',
    media: 'suiteAgents',
    tone: 'bg-sky-200',
    light: false,
    mock: <AgentRunMock />,
  },
  {
    title: 'اتصال‌ها',
    text: 'CRM، بله، ایتا، Jira، Notion، Git و سرورهای MCP.',
    href: '/features/mcp-gateway',
    media: 'suiteConnections',
    tone: 'bg-amber-200',
    light: false,
    mock: <ConnectionsMock />,
  },
  {
    title: 'درگاه API',
    text: 'برای تیم فنی: یک نشانی برای همه‌ی مدل‌ها.',
    href: '/docs',
    media: 'suiteApi',
    tone: 'bg-emerald-200',
    light: false,
  },
  {
    title: 'کنترل',
    text: 'گاردریل، سقف مصرف و گزارش رخدادها.',
    href: '/features/guardrails',
    media: 'suiteControl',
    tone: 'bg-rose-200',
    light: false,
  },
];

export function WorkplaceSuite() {
  return (
    <section className="overflow-hidden">
      <div className={`${sectionPad} pb-20 md:pb-24 lg:pb-30`}>
        <div className="mx-auto w-full max-w-7xl">
          <Reveal className="mb-8 md:mb-12 lg:mb-16">
            <TwoTone lead="هر آنچه سازمان برای شروع لازم دارد." muted="برای همکار عادی ساده، برای تیم فنی کامل." />
          </Reveal>

          <Carousel label="محصولات گدارAI">
            {cards.map((card) => (
              <SuiteCardView key={card.title} card={card} />
            ))}
          </Carousel>
        </div>
      </div>
    </section>
  );
}

function SuiteCardView({ card }: { card: SuiteCard }) {
  return (
    <article
      className={`relative flex h-128 w-80 shrink-0 snap-start flex-col overflow-hidden rounded-xl p-8 sm:w-96 sm:p-10 ${card.tone} ${
        card.light ? 'text-white' : 'text-zinc-900'
      }`}
    >
      <h3 className="mb-2 text-lg font-medium">{card.title}</h3>
      <p className="mb-6 text-base leading-8 opacity-75">{card.text}</p>
      <div className="mb-6">
        <CircleLink href={card.href} tone={card.light ? 'light' : 'ink'}>
          بیشتر بدانید
        </CircleLink>
      </div>
      <div className="mt-auto min-h-0 flex-1">
        <MediaSlot id={card.media} className="h-full" fallback={card.mock} />
      </div>
    </article>
  );
}
