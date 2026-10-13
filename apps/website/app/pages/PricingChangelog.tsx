import Link from "next/link";

export type PublicPricingChangelog = {
  rate_cards: {
    version: number;
    label: string;
    status: "scheduled" | "published" | "superseded" | string;
    effective_from: string;
    changelog_fa: string;
    changelog_en: string;
  }[];
  credit_prices: {
    price_toman_per_credit: number;
    status: "active" | "superseded" | string;
    effective_from: string;
    plan_effective_from: string;
  }[];
};

const dateFormatter = new Intl.DateTimeFormat("fa-IR-u-ca-persian", { dateStyle: "long" });
const numberFormatter = new Intl.NumberFormat("fa-IR");

function formatDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : dateFormatter.format(date);
}

const rateCardStatus: Record<string, string> = {
  scheduled: "زمان‌بندی‌شده",
  published: "فعال",
  superseded: "قبلی",
};

const priceStatus: Record<string, string> = {
  active: "فعال",
  superseded: "قبلی",
};

export function PricingChangelog({ changelog }: { changelog: PublicPricingChangelog | null }) {
  return (
    <main dir="rtl" className="px-4 py-24 text-right sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-4xl flex-col gap-10">
        <header className="flex flex-col gap-3">
          <h1 className="text-3xl font-bold tracking-tight text-foreground md:text-4xl">تغییرات قیمت و نرخ‌نامه</h1>
          <p className="leading-8 text-muted-foreground">
            هر تغییر در قیمت هر کردیت یا نرخ‌نامه‌ی مدل‌ها همراه با تاریخ اعمال آن در این صفحه ثبت می‌شود. کردیت‌هایی که
            قبلاً خریده‌اید هیچ‌وقت ارزش مصرفشان را از دست نمی‌دهند. افزایش قیمت پلن‌ها و افزایش نرخ مدل‌ها دست‌کم ۳۰ روز
            پیش از اعمال اعلام می‌شود؛ کاهش‌ها فوراً اعمال می‌شوند.
          </p>
          <Link href="/pricing" className="text-sm font-medium text-primary hover:underline">بازگشت به قیمت‌گذاری</Link>
        </header>

        {changelog === null ? (
          <p role="status" className="rounded-md border border-border bg-card p-4 text-muted-foreground">
            تاریخچه‌ی قیمت‌ها در حال حاضر در دسترس نیست. کمی بعد دوباره سر بزنید.
          </p>
        ) : (
          <>
            <section aria-labelledby="credit-prices" className="flex flex-col gap-4">
              <h2 id="credit-prices" className="text-xl font-semibold text-foreground">قیمت هر کردیت</h2>
              <ol className="flex flex-col divide-y divide-border rounded-md border border-border bg-card">
                {changelog.credit_prices.map((price) => (
                  <li key={price.effective_from} className="flex flex-col gap-1 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="text-lg font-semibold text-foreground">
                        {numberFormatter.format(price.price_toman_per_credit)} تومان
                      </span>
                      <span className="rounded-full border border-border px-3 py-0.5 text-xs text-muted-foreground">
                        {priceStatus[price.status] ?? price.status}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      بسته‌های کردیت از {formatDate(price.effective_from)} · پلن‌ها از دوره‌ی بعد از {formatDate(price.plan_effective_from)}
                    </p>
                  </li>
                ))}
                {changelog.credit_prices.length === 0 ? <li className="p-4 text-muted-foreground">هنوز تغییری ثبت نشده است.</li> : null}
              </ol>
            </section>

            <section aria-labelledby="rate-cards" className="flex flex-col gap-4">
              <h2 id="rate-cards" className="text-xl font-semibold text-foreground">نسخه‌های نرخ‌نامه</h2>
              <ol className="flex flex-col divide-y divide-border rounded-md border border-border bg-card">
                {changelog.rate_cards.map((card) => (
                  <li key={card.version} className="flex flex-col gap-1 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="font-semibold text-foreground" dir="ltr">{card.label}</span>
                      <span className="rounded-full border border-border px-3 py-0.5 text-xs text-muted-foreground">
                        {rateCardStatus[card.status] ?? card.status}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground">اعمال از {formatDate(card.effective_from)}</p>
                    {card.changelog_fa ? <p className="leading-7 text-foreground">{card.changelog_fa}</p> : null}
                  </li>
                ))}
                {changelog.rate_cards.length === 0 ? <li className="p-4 text-muted-foreground">هنوز نسخه‌ای منتشر نشده است.</li> : null}
              </ol>
            </section>
          </>
        )}
      </div>
    </main>
  );
}
