import { ArrowUp, Check, Cpu, Database, GitBranch, KanbanSquare, Loader2, MessageCircle, NotebookText, Plug, Plus, X } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { ReactNode } from 'react';

// Product-UI fragments drawn in code. They stand in for real screenshots on the
// `/workplace` page and are replaced automatically once a slot gets a `src`.

function Describe({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div role="img" aria-label={label} className="h-full w-full">
      {children}
    </div>
  );
}

function FileChip({ tone, letter, name }: { tone: string; letter: string; name: string }) {
  return (
    <span className="inline-flex items-center gap-2 rounded-lg bg-zinc-100 px-2 py-1.5 text-xs text-zinc-800">
      <span className={`flex size-5 items-center justify-center rounded text-[0.625rem] font-bold text-white ${tone}`}>{letter}</span>
      <span dir="ltr">{name}</span>
    </span>
  );
}

export function ChatMock() {
  return (
    <Describe label="کادر پیام در گفتگوی گدارAI با فایل پیوست، انتخاب مدل و دکمه‌ی ارسال">
      <div className="flex h-full items-center justify-center">
        <div className="w-full max-w-sm rounded-2xl bg-white p-3 text-zinc-900 shadow-xl ring-4 ring-white/15">
          <div className="mb-1 flex items-center gap-2">
            <FileChip tone="bg-emerald-600" letter="X" name="sales-report.xlsx" />
            <X className="size-3.5 text-zinc-400" />
          </div>
          <p className="py-3 text-sm text-zinc-400">از دستیار بپرسید...</p>
          <div className="flex items-center gap-2">
            <span className="flex size-8 items-center justify-center rounded-full border border-zinc-200">
              <Plus className="size-4 text-zinc-600" />
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-full border border-zinc-200 px-3 py-1.5 text-xs text-zinc-700">
              <Cpu className="size-3.5" />
              انتخاب مدل
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-full border border-zinc-200 px-3 py-1.5 text-xs text-zinc-700">
              <Plug className="size-3.5" />
              اتصال‌ها
            </span>
            <span className="ms-auto flex size-8 items-center justify-center rounded-full bg-blue-500 text-white">
              <ArrowUp className="size-4" />
            </span>
          </div>
        </div>
      </div>
    </Describe>
  );
}

export function AgentRunMock() {
  return (
    <Describe label="مراحل کار یک دستیار پشتیبانی: جست‌وجو در CRM، خواندن دو سند و نوشتن پیش‌نویس پاسخ">
      <div className="flex h-full items-center justify-center">
        <ol className="relative flex w-full max-w-xs flex-col gap-7 text-sm text-zinc-800">
          <li className="relative flex gap-3">
            <span className="relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full bg-zinc-700 text-white">
              <Check className="size-3.5" />
            </span>
            <div className="flex flex-col gap-2">
              <span>در CRM جست‌وجو کرد</span>
              <span className="inline-flex w-fit items-center gap-2 rounded-lg bg-black/5 px-2 py-1 text-xs">
                <Database className="size-3.5 text-blue-600" />
                سفارش ۱۴۸۲
              </span>
            </div>
          </li>
          <li className="relative flex gap-3">
            <span className="relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full bg-zinc-700 text-white">
              <Check className="size-3.5" />
            </span>
            <div className="flex flex-col gap-2">
              <span>۲ سند را خواند</span>
              <div className="flex flex-wrap gap-2">
                <FileChip tone="bg-rose-600" letter="P" name="invoice-1482.pdf" />
                <FileChip tone="bg-blue-600" letter="W" name="contract.docx" />
              </div>
            </div>
          </li>
          <li className="relative flex gap-3 text-zinc-500">
            <span className="relative z-10 flex size-6 shrink-0 items-center justify-center rounded-full bg-black/5">
              <Loader2 className="size-3.5 animate-spin" />
            </span>
            <span>در حال نوشتن پیش‌نویس پاسخ...</span>
          </li>
          <span aria-hidden="true" className="absolute start-3 top-3 bottom-3 w-px bg-zinc-800/30" />
        </ol>
      </div>
    </Describe>
  );
}

const connections: { icon: LucideIcon; name: string; text: string; tone: string }[] = [
  { icon: Database, name: 'CRM', text: 'سابقه‌ی مشتری و سفارش‌ها', tone: 'bg-blue-50 text-blue-600' },
  { icon: MessageCircle, name: 'بله و ایتا', text: 'پیام‌رسان‌های داخلی', tone: 'bg-emerald-50 text-emerald-600' },
  { icon: KanbanSquare, name: 'Jira', text: 'تسک‌ها و پروژه‌ها', tone: 'bg-indigo-50 text-indigo-600' },
  { icon: NotebookText, name: 'Notion', text: 'سندها و دانش تیم', tone: 'bg-zinc-100 text-zinc-700' },
  { icon: GitBranch, name: 'Git', text: 'کد و تاریخچه‌ی تغییرها', tone: 'bg-orange-50 text-orange-600' },
  { icon: Plug, name: 'MCP', text: 'هر ابزار دلخواه', tone: 'bg-violet-50 text-violet-600' },
];

export function ConnectionsMock() {
  return (
    <Describe label="منوی افزودن اتصال با CRM، پیام‌رسان‌ها، Jira، Notion، Git و سرور MCP">
      <div className="flex h-full items-start justify-center overflow-hidden">
        <div className="w-full max-w-xs rounded-xl bg-white p-4 text-zinc-900 shadow-xl ring-1 ring-black/5">
          <div className="mb-3 flex items-center gap-2 text-sm font-medium">
            <Plus className="size-4" />
            افزودن اتصال
            <X className="ms-auto size-3.5 text-zinc-400" />
          </div>
          <p className="mb-1 text-[0.6875rem] text-zinc-400">ابزارها</p>
          <ul className="flex flex-col gap-1">
            {connections.map((item, index) => (
              <li key={item.name} className={`flex items-center gap-3 rounded-lg px-2 py-2 ${index === 0 ? 'bg-zinc-100' : ''}`}>
                <span className={`flex size-8 shrink-0 items-center justify-center rounded-lg ${item.tone}`}>
                  <item.icon className="size-4" />
                </span>
                <span className="flex flex-col">
                  <span className="text-xs font-medium">{item.name}</span>
                  <span className="text-[0.6875rem] text-zinc-500">{item.text}</span>
                </span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </Describe>
  );
}
