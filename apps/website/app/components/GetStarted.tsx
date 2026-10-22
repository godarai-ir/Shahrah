import Link from 'next/link';
import { Section, StepList } from './patterns';
import { sectionHeading } from './sectionType';

const steps = [
  { title: 'فضای کاری بسازید', text: 'ثبت‌نام می‌کنید و فضای کاری سازمانتان همان لحظه آماده است.' },
  { title: 'یک دستیار و یک اتصال تعریف کنید', text: 'مثلاً دستیار پشتیبانی که به CRM شما وصل است. برای شروع همین کافی است.' },
  { title: 'تیم را دعوت کنید', text: 'سقف مصرف را بگذارید و بگذارید همکارها امتحان کنند.' },
];

const code = `from openai import OpenAI

client = OpenAI(
    api_key="YOUR_GODARAI_TOKEN",
    base_url="https://YOUR_WORKSPACE.godarai.ir/v1",
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "یک پیام کوتاه برای پیگیری سفارش بنویس."}],
)`;

const textLink =
  'font-medium text-site-accent underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-accent';

export function GetStarted() {
  return (
    <Section id="start">
      <div>
        <h2 className={`mb-14 max-w-3xl ${sectionHeading}`}>سه قدم تا اولین دستیار</h2>
      </div>

      <div className="mb-6">
        <StepList steps={steps} />
      </div>
      <p className="mb-20 text-sm leading-7 text-site-ink-soft">
        اگر وسط راه گیر کردید، به ما بگویید.{' '}
        <Link href="/contact" className={textLink}>
          تماس با ما
        </Link>
      </p>

      <div className="grid items-center gap-10 rounded-3xl bg-site-panel p-8 sm:p-12 lg:grid-cols-2 lg:gap-16">
        <div>
          <h3 className="mb-3 text-2xl font-semibold text-site-ink sm:text-3xl">برای تیم فنی</h3>
          <p className="mb-5 text-base leading-8 text-site-ink-soft">
            اگر خودتان محصول می‌سازید، همان درگاه را مستقیم هم می‌توانید بردارید. API سازگار با OpenAI است؛ فقط نشانی و کلید را عوض می‌کنید.
          </p>
          <Link href="/docs" className={`text-sm ${textLink}`}>
            مشاهده‌ی مستندات
          </Link>
        </div>
        <div dir="ltr" className="min-w-0 overflow-hidden rounded-2xl bg-site-night">
          <pre className="overflow-x-auto p-5 text-xs leading-6 text-site-night-soft sm:p-6 sm:text-sm">
            <code>{code}</code>
          </pre>
        </div>
      </div>
    </Section>
  );
}
