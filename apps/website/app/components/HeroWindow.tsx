import { Database, KanbanSquare, MessageCircle, NotebookText } from 'lucide-react';
import { ProductWindow, ToolSteps } from './patterns';
import type { ToolStep } from './patterns';

const tools: ToolStep[] = [
  { icon: Database, name: 'CRM', text: 'سابقه‌ی مشتری و سفارش‌ها' },
  { icon: MessageCircle, name: 'بله و ایتا', text: 'پیام‌رسان‌های داخلی' },
  { icon: KanbanSquare, name: 'Jira', text: 'تسک‌ها و پروژه‌ها' },
  { icon: NotebookText, name: 'Notion', text: 'سندها و دانش تیم' },
];

// A drawn product window: an assistant connected to the team's tools finishes one task.
export function HeroWindow() {
  return (
    <ProductWindow
      title="دستیار پشتیبانی"
      meta="نمونه"
      label="نمونه‌ای از یک دستیار پشتیبانی که به CRM، پیام‌رسان‌ها، Jira و Notion وصل است و پیش‌نویس پاسخ یک سفارش را آماده می‌کند."
    >
      <ToolSteps
        steps={tools}
        result={{ title: 'پیش‌نویس پاسخ سفارش ۱۴۸۲', text: 'آماده‌ی تأیید همکار پشتیبانی' }}
      />
    </ProductWindow>
  );
}
