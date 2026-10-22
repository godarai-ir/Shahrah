import { BookOpen, Database, MessagesSquare } from 'lucide-react';
import { FeatureRow, ProductStage, ProductWindow, ToolSteps } from './patterns';
import type { ToolStep } from './patterns';

const steps: ToolStep[] = [
  { icon: Database, name: 'جست‌وجو در CRM', text: 'سفارش ۱۴۸۲ و فاکتورش' },
  { icon: MessagesSquare, name: 'تیکت‌های قبلی', text: 'دو گفتگوی پیشین همین مشتری' },
  { icon: BookOpen, name: 'پایگاه دانش', text: 'روال اصلاح فاکتور' },
];

// One task end to end. It replaces the former collaboration section, whose point (people and the assistant hand
// work back and forth inside the team's own tools) now lives in the copy here.
export function Scenario() {
  return (
    <FeatureRow
      id="scenario"
      reverse
      title="از سابقه‌ی مشتری تا پیش‌نویس پاسخ"
      copy="دستیار پشتیبانی سابقه‌ی مشتری را می‌خواند و پیش‌نویس پاسخ را آماده می‌کند. همکارتان می‌خواند، اصلاح می‌کند و می‌فرستد؛ از همان ابزارهایی که تیم هر روز باز می‌کند."
      visual={
        <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
          <ProductWindow
            title="دستیار پشتیبانی"
            meta="نمونه"
            size="m-auto max-w-md"
            label="نمونه‌ای از کار یک دستیار پشتیبانی: سفارش مشتری را در CRM پیدا می‌کند، تیکت‌های قبلی و پایگاه دانش را می‌خواند و پیش‌نویس پاسخ را آماده می‌کند."
          >
            <div className="border-b border-site-line p-5">
              <p className="mb-1 text-xs text-site-ink-soft">پیام مشتری</p>
              <p className="m-0 text-sm leading-7 text-site-ink">سلام، مبلغ فاکتور سفارش ۱۴۸۲ اشتباه است. لطفاً بررسی کنید.</p>
            </div>
            <ToolSteps steps={steps} result={{ title: 'پیش‌نویس پاسخ آماده است', text: 'برای همکار پشتیبانی، در صف تیکت‌ها' }} />
          </ProductWindow>
        </ProductStage>
      }
    />
  );
}
