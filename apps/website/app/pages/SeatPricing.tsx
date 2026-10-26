import Link from 'next/link';
import { BadgeCheck, CalendarClock, Check, CreditCard, PowerOff } from 'lucide-react';
import { ClosingBand, FAQ, Section, StepList, TileGrid } from '@/app/components/patterns';
import { buttonPrimary, buttonSecondary, sectionCopy, sectionHeading } from '@/app/components/sectionType';

type SeatDefinition = {
  slug: 'free' | 'pro' | 'max';
  name: string;
  description: string;
  credits: number;
  creditsLabel: string;
  features: string[];
  highlighted?: boolean;
};

const seatDefinitions: SeatDefinition[] = [
  {
    slug: 'free',
    name: 'رایگان',
    description: 'برای شروع کار و آشنایی اعضای تیم با GodarChat',
    credits: 500,
    creditsLabel: 'اعتبار یک‌باره برای هر عضو',
    features: ['تا ۵ عضو با صندلی رایگان در هر سازمان', 'همه‌ی امکانات پلتفرم', 'اعتبار شخصی فقط برای خود عضو', 'هر عضو فقط یک‌بار صندلی رایگان می‌گیرد'],
  },
  {
    slug: 'pro',
    name: 'Pro',
    description: 'برای اعضایی که هر روز با هوش مصنوعی کار می‌کنند',
    credits: 2000,
    creditsLabel: 'اعتبار شخصی در هر ماه',
    features: ['همه‌ی امکانات پلتفرم', 'اعتبار هر ماه از نو داده می‌شود', 'پس از اعتبار شخصی، مصرف از استخر سازمان تا سقف تعیین‌شده', 'ارتقای فوری به Max با اعتبار کامل ماه'],
    highlighted: true,
  },
  {
    slug: 'max',
    name: 'Max',
    description: 'برای تحلیل‌گران، تولید محتوا و مصرف سنگین',
    credits: 8000,
    creditsLabel: 'اعتبار شخصی در هر ماه',
    features: ['همه‌ی امکانات پلتفرم', 'چهار برابر اعتبار Pro با قیمت کمتر برای هر اعتبار', 'پس از اعتبار شخصی، مصرف از استخر سازمان تا سقف تعیین‌شده', 'کاهش به Pro در پایان دوره'],
  },
];

const principles = [
  { icon: BadgeCheck, title: 'همه‌ی امکانات برای همه', text: 'پلن و قابلیت قفل‌شده وجود ندارد؛ هر سازمان از روز اول به همه‌ی امکانات گدارAI دسترسی دارد.' },
  { icon: CreditCard, title: 'پرداخت پیش از استفاده', text: 'صندلی‌ها و بسته‌های اعتبار از درگاه پرداخت و پیش از استفاده پرداخت می‌شوند. کارت شما ذخیره و خودکار شارژ نمی‌شود.' },
  { icon: CalendarClock, title: 'بدون تمدید خودکار', text: 'اشتراک خودکار تمدید نمی‌شود. از ۷ روز مانده به پایان دوره تا روز آخر می‌توانید آن را تمدید کنید.' },
  { icon: PowerOff, title: 'پایان اشتراک، فقط صندلی‌های پولی', text: 'اگر اشتراک تمدید نشود فقط صندلی‌های پولی غیرفعال می‌شوند؛ صندلی‌های رایگان، اعتبار بسته‌ها و API کار می‌کنند.' },
] as const;

const creditPackPoints = [
  'قیمت هر بسته از قیمت فهرست اعتبار ساخته می‌شود و بسته‌های بزرگ‌تر قیمت کمتری برای هر اعتبار دارند.',
  'خرید بسته به صندلی پولی یا اشتراک وابسته نیست و پیش از استفاده از درگاه پرداخت می‌شود.',
  'اعتبار بسته‌ها با پایان اشتراک صندلی‌ها از بین نمی‌رود و API به کار خود ادامه می‌دهد.',
  'اعضای Pro و Max پس از اعتبار شخصی، تا سقفی که مدیر تعیین می‌کند از استخر استفاده می‌کنند.',
];

const subscriptionSteps = [
  { title: 'خرید اولین صندلی پولی', text: 'اشتراک همان پرداخت صندلی‌های پولی است. با خرید اولین صندلی Pro یا Max اشتراک شروع می‌شود و دوره‌ی آن، ماهانه یا سالانه، همان‌جا انتخاب می‌شود.' },
  { title: 'افزودن صندلی در میانه‌ی دوره', text: 'صندلی‌هایی که بعداً می‌خرید هم‌دوره با اشتراک‌اند و فقط به نسبت روزهای باقی‌مانده‌ی دوره پرداخت می‌شوند. صندلی بلافاصله پس از پرداخت فعال می‌شود.' },
  { title: 'تمدید در هفته‌ی آخر', text: 'از ۷ روز مانده به پایان دوره تا روز آخر، تمدید در پنل مدیریت سازمان فعال است. پرداخت از درگاه انجام می‌شود و هیچ مبلغی خودکار کسر نمی‌شود.' },
  { title: 'اگر تمدید نکنید', text: 'در پایان دوره فقط صندلی‌های پولی غیرفعال می‌شوند و اعضای آن‌ها بدون صندلی می‌مانند. صندلی‌های رایگان، اعتبار بسته‌ها و کلیدهای API مثل قبل کار می‌کنند.' },
] as const;

const pricingFaqs = [
  {
    question: 'پلن‌های قبلی (توسعه‌دهنده، تیم، رشد و سازمانی) چه شدند؟',
    answer: 'پلن‌ها حذف شده‌اند و همه‌ی امکانات برای همه‌ی سازمان‌ها در دسترس است. هزینه فقط از دو جا می‌آید: صندلی‌های پولی اعضا و بسته‌های اعتبار برای استخر سازمان. اشتراک‌های پولی فعلی بلافاصله به مدل جدید منتقل شده‌اند.',
  },
  {
    question: 'دوره‌ی آزمایشی دارید؟',
    answer: 'دوره‌ی آزمایشی جداگانه نداریم، چون نیازی به آن نیست: تا ۵ عضو هر سازمان صندلی رایگان با ۵۰۰ اعتبار می‌گیرند و می‌توانند بدون پرداخت با همه‌ی امکانات کار کنند.',
  },
  {
    question: 'قیمت صندلی Pro و Max چطور حساب می‌شود؟',
    answer: 'قیمت هر صندلی برابر است با اعتبار ماهانه‌ی آن × قیمت فهرست فعال هر اعتبار، منهای تخفیف همان صندلی. صورتحساب سالانه تخفیف بیشتری دارد ولی اعتبار همچنان ماه‌به‌ماه داده می‌شود. مالیات بر ارزش افزوده جداگانه اضافه می‌شود. قیمت هر اعتبار و تاریخچه‌ی تغییراتش در صفحه‌ی تغییرات قیمت منتشر می‌شود.',
  },
  {
    question: 'استخر سازمان و بسته‌های اعتبار برای چیست؟',
    answer: 'استخر سازمان اعتبار مشترک سازمان است و فقط با خرید بسته‌های اعتبار پر می‌شود. مصرف API، کلیدهای دسترسی، triggerها و گاردریل‌ها از همین استخر کم می‌شود. اعضای Pro و Max پس از تمام شدن اعتبار شخصی‌شان می‌توانند تا سقفی که مدیر تعیین می‌کند از استخر هم استفاده کنند.',
  },
  {
    question: 'اگر عضوی را حذف کنیم، صندلی پولی او چه می‌شود؟',
    answer: 'صندلی به انبار صندلی‌های سازمان برمی‌گردد و تا پایان دوره‌ی پرداخت‌شده به عضو دیگری قابل تخصیص است. صندلی‌ها بازپرداخت ریالی ندارند.',
  },
  {
    question: 'یادم برود تمدید کنم چه می‌شود؟',
    answer: 'از ۷ روز مانده به پایان دوره در پنل یادآوری می‌بینید. اگر تمدید نکنید فقط صندلی‌های پولی غیرفعال می‌شوند؛ هر وقت دوباره صندلی بخرید، اشتراک تازه‌ای با دوره‌ی جدید شروع می‌شود.',
  },
  {
    question: 'برای پرسش‌های فنی یا استقرار اختصاصی از کجا کمک بگیریم؟',
    answer: 'از صفحه‌ی تماس با ما یا ایمیل support@godarai.ir با تیم گدارAI در ارتباط باشید. برای استقرار در محل یا ابر خصوصی، بررسی نیازها به‌صورت اختصاصی انجام می‌شود.',
  },
] as const;

/**
 * Public seat pricing. Seat prices are not hard-coded: they follow the active credit list price,
 * which the page receives from the public pricing changelog when it is available.
 */
export function SeatPricing({ creditPriceToman }: { creditPriceToman: number | null }) {
  return (
    <main className="bg-site-ground">
      <section className="px-4 pb-10 pt-12 sm:px-6 lg:pb-20 lg:pt-20">
        <div className="mx-auto flex w-full max-w-content flex-col gap-10">
          <header className="mx-auto flex max-w-3xl flex-col items-center gap-4 text-center">
            <h1 className="m-0 text-balance text-h1 font-extrabold text-site-ink lg:text-display">برای هر عضو یک صندلی. برای API استخر سازمان.</h1>
            <p className={`m-0 max-w-2xl ${sectionCopy}`}>
              پلنی برای انتخاب وجود ندارد. هر عضو با صندلی خودش اعتبار شخصی برای GodarChat دارد و مصرف API از استخر مشترک سازمان پرداخت می‌شود که با بسته‌های اعتبار پر می‌شود.
            </p>
          </header>

          <div className="grid items-stretch gap-6 md:grid-cols-3">
            {seatDefinitions.map((seat) => (
              <SeatCard key={seat.slug} seat={seat} creditPriceToman={creditPriceToman} />
            ))}
          </div>

          <p className="m-0 text-center text-label text-site-ink-soft">
            {creditPriceToman === null
              ? 'قیمت فهرست هر اعتبار موقتاً در دسترس نیست؛ قیمت دقیق هر صندلی هنگام خرید در پنل مدیریت سازمان نمایش داده می‌شود.'
              : `قیمت فهرست فعلی هر اعتبار ${formatNumber(creditPriceToman)} تومان است. قیمت نهایی هر صندلی با تخفیف همان صندلی و مالیات، هنگام خرید در پنل مدیریت سازمان نمایش داده می‌شود.`}
          </p>
        </div>
      </section>

      <Section id="principles" tone="panel" layout="flex flex-col gap-8">
        <h2 className={`max-w-3xl ${sectionHeading}`}>قواعد پرداخت. کوتاه و روشن.</h2>
        <TileGrid size="compact" items={principles.map(({ icon, title, text }) => ({ key: title, icon, title, text }))} />
      </Section>

      <Section id="credit-packs" layout="grid gap-10 lg:grid-cols-2 lg:items-center">
        <div className="flex flex-col gap-4">
          <h2 className={sectionHeading}>استخر مشترک سازمان برای API</h2>
          <p className={`m-0 ${sectionCopy}`}>
            بسته‌های اعتبار استخر مشترک سازمان را پر می‌کنند. مصرف API، کلیدهای دسترسی، triggerها و گاردریل‌ها فقط از این استخر کم می‌شود و هرگز اعتبار شخصی اعضا را مصرف نمی‌کند.
          </p>
        </div>
        <ul className="m-0 flex list-none flex-col gap-4 rounded-card border border-site-line bg-site-surface p-6 lg:p-8">
          {creditPackPoints.map((item) => (
            <li key={item} className="flex items-start gap-3">
              <Check aria-hidden="true" className="mt-1 size-4 shrink-0 text-site-ink" />
              <span className="text-body text-site-ink">{item}</span>
            </li>
          ))}
        </ul>
      </Section>

      <Section id="subscription" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>اشتراک یعنی پرداخت صندلی‌های پولی.</h2>
          <p className={`m-0 max-w-2xl ${sectionCopy}`}>تا وقتی صندلی پولی نخریده‌اید اشتراکی ندارید و چیزی پرداخت نمی‌کنید.</p>
        </div>
        <StepList steps={[...subscriptionSteps]} />
      </Section>

      <FAQ id="pricing-faq" title="پیش از خرید صندلی بدانید." items={[...pricingFaqs]} />

      <ClosingBand
        title="با صندلی‌های رایگان شروع کنید."
        titleMuted="بعد تصمیم بگیرید."
        copy="سازمان را بسازید، تا ۵ عضو را رایگان دعوت کنید و هر وقت لازم شد صندلی پولی یا بسته‌ی اعتبار بخرید."
        primary={{ href: '/onboarding', label: 'ساخت سازمان رایگان' }}
        secondary={{ href: '/pricing/changelog', label: 'تغییرات قیمت' }}
      />
    </main>
  );
}

function SeatCard({ seat, creditPriceToman }: { seat: SeatDefinition; creditPriceToman: number | null }) {
  const isFree = seat.slug === 'free';

  return (
    <article className={`flex flex-col gap-5 rounded-card border bg-site-surface p-6 lg:p-8 ${seat.highlighted ? 'border-site-ink ring-2 ring-site-ink' : 'border-site-line'}`}>
      <div className="flex items-start justify-between gap-3">
        <div className="flex flex-col gap-1">
          <h2 className="m-0 text-h3 font-semibold text-site-ink">{seat.name}</h2>
          <p className="m-0 text-label text-site-ink-soft">{seat.description}</p>
        </div>
        {seat.highlighted ? (
          <span className="shrink-0 rounded-full bg-site-accent px-3 py-1 text-caption font-medium text-site-on-accent">پرکاربردترین</span>
        ) : null}
      </div>

      <div className="flex flex-col gap-1 border-y border-site-line py-4">
        <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
          <span className="text-h2 font-extrabold text-site-ink">{formatNumber(seat.credits)}</span>
          <span className="text-label text-site-ink-soft">{seat.creditsLabel}</span>
        </div>
        <p className="m-0 text-label text-site-ink-soft">
          {isFree
            ? 'رایگان، بدون نیاز به پرداخت'
            : creditPriceToman === null
              ? 'قیمت ماهانه = اعتبار ماهانه × قیمت هر اعتبار، منهای تخفیف صندلی'
              : `پیش از تخفیف صندلی: ${formatNumber(seat.credits * creditPriceToman)} تومان در ماه`}
        </p>
      </div>

      <ul className="m-0 flex flex-1 list-none flex-col gap-3 p-0">
        {seat.features.map((feature) => (
          <li key={feature} className="flex items-start gap-3">
            <Check aria-hidden="true" className="mt-1 size-4 shrink-0 text-site-ink" />
            <span className="text-body text-site-ink">{feature}</span>
          </li>
        ))}
      </ul>

      <Link href="/onboarding" className={seat.highlighted ? buttonPrimary : buttonSecondary}>
        {isFree ? 'شروع رایگان' : 'خرید صندلی در پنل'}
      </Link>
    </article>
  );
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('fa-IR').format(value);
}
