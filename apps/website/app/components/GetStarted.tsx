import Link from 'next/link';

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

export function GetStarted() {
  return (
    <section id="start" className="py-14 lg:py-24">
      <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
        <h2 className="mb-12 max-w-3xl text-balance text-3xl font-bold leading-tight text-slate-950 dark:text-white sm:mb-16 sm:text-4xl lg:text-5xl lg:leading-tight">
          سه قدم تا اولین دستیار
        </h2>

        <ol className="mb-6 grid gap-8 md:grid-cols-3">
          {steps.map((step, index) => (
            <li key={step.title}>
              <p className="mb-3 text-3xl font-bold text-blue-500">{(index + 1).toLocaleString('fa-IR')}</p>
              <h3 className="mb-2 text-xl font-bold text-slate-950 dark:text-white">{step.title}</h3>
              <p className="text-base leading-8 text-slate-600 dark:text-gray-400">{step.text}</p>
            </li>
          ))}
        </ol>
        <p className="mb-16 text-sm leading-7 text-slate-600 dark:text-gray-400">
          اگر وسط راه گیر کردید، به ما بگویید.{' '}
          <Link href="/contact" className="font-medium text-blue-600 hover:underline dark:text-blue-300">
            تماس با ما
          </Link>
        </p>

        <div className="grid items-center gap-8 rounded-3xl bg-blue-50 p-6 dark:bg-blue-500/10 sm:p-10 lg:grid-cols-2 lg:gap-14">
          <div>
            <h3 className="mb-3 text-2xl font-bold text-slate-950 dark:text-white sm:text-3xl">برای تیم فنی</h3>
            <p className="mb-4 text-base leading-8 text-slate-600 dark:text-gray-400">
              اگر خودتان محصول می‌سازید، همان درگاه را مستقیم هم می‌توانید بردارید. API سازگار با OpenAI است؛ فقط نشانی
              و کلید را عوض می‌کنید.
            </p>
            <Link href="/docs" className="text-sm font-medium text-blue-600 hover:underline dark:text-blue-300">
              مشاهده‌ی مستندات
            </Link>
          </div>
          <div dir="ltr" className="min-w-0 overflow-hidden rounded-2xl bg-slate-950">
            <pre className="overflow-x-auto p-4 text-xs leading-6 text-gray-300 sm:p-6 sm:text-sm">
              <code>{code}</code>
            </pre>
          </div>
        </div>
      </div>
    </section>
  );
}
