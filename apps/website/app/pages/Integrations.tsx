import { Database, GitBranch, KanbanSquare, KeyRound, MessageCircle, NotebookText, Server, Wrench } from 'lucide-react';
import {
  ClosingBand,
  FeatureRow,
  MockRows,
  PageHero,
  ProductStage,
  ProductWindow,
  Section,
  TileGrid,
  ToolSteps,
} from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

// Only connections that work today, in the fixed phrasings of docs/site-architecture/04-open-questions.md:
// messengers as one group (Eitaa not named separately), CRM without product names. No "coming soon".
export function Integrations() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="دستیار، همان‌جا که تیم کار می‌کند. نه یک پنجره‌ی دیگر."
        lead="دستیارها به ابزارهای تیم و سرورهای MCP سازمان وصل می‌شوند. شما تعیین می‌کنید کدام تیم به کدام ابزار برسد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <div className="relative w-full max-w-sm">
              <ProductWindow title="دستیار فنی" meta="نمونه" label="نمونه‌ای از دستیاری که Jira، Git و Notion را فرامی‌خواند و خلاصه‌ی وضعیت را آماده می‌کند.">
                <ToolSteps
                  steps={[
                    { icon: KanbanSquare, name: 'Jira', text: 'تسک باز این هفته' },
                    { icon: GitBranch, name: 'Git', text: 'تغییرهای مرتبط' },
                    { icon: NotebookText, name: 'Notion', text: 'سند طراحی' },
                  ]}
                  result={{ title: 'خلاصه‌ی وضعیت آماده است', text: 'با پیوند به تغییرها و سند' }}
                />
              </ProductWindow>
            </div>
          </ProductStage>
        }
      />

      <Section id="catalog" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>آنچه امروز وصل می‌شود.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>ابزارهای کار تیمی، کد، CRM و پیام‌رسان‌ها؛ و هر چیز دیگری که سرور MCP دارد.</p>
        </div>
        <TileGrid
          columns={3}
          size="compact"
          items={[
            { key: 'messengers', icon: MessageCircle, title: 'پیام‌رسان‌ها', text: 'ارسال و خواندن پیام در پیام‌رسان‌ها مانند تلگرام، واتساپ، بله و …' },
            { key: 'jira', icon: KanbanSquare, title: 'Jira', text: 'تسک‌ها و پروژه‌ها' },
            { key: 'notion', icon: NotebookText, title: 'Notion', text: 'سندها و دانش تیم' },
            { key: 'git', icon: GitBranch, title: 'Git', text: 'تغییرهای کد' },
            { key: 'crm', icon: Database, title: 'CRM از راه MCP', text: 'سابقه‌ی مشتری، سفارش‌ها و فرصت‌ها' },
            { key: 'mcp', icon: Server, title: 'سرورهای MCP سازمان', text: 'هر ابزار داخلی که سرور MCP دارد' },
          ]}
        />
      </Section>

      <FeatureRow
        id="own-mcp"
        title="ابزار داخلی دارید؟ وصلش کنید."
        copy="ابزار داخلی را به‌صورت یک سرور MCP اضافه کنید تا دستیارها از آن استفاده کنند؛ مثل هر ابزار دیگر، با دسترسی کنترل‌شده."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="سرورهای MCP" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از فهرست سرورهای MCP سازمان: CRM و سامانه‌ی انبار.">
              <MockRows
                rows={[
                  { icon: Server, title: 'CRM', text: 'سه ابزار', badge: { label: 'وصل', tone: 'success' } },
                  { icon: Server, title: 'سامانه‌ی انبار (داخلی)', text: 'دو ابزار', badge: { label: 'وصل', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="access"
        tone="panel"
        reverse
        title="کلید ابزار پیش ما می‌ماند. دسترسی دست شماست."
        copy="اعتبارنامه‌ی هر ابزار در گدارAI می‌ماند و به کاربر داده نمی‌شود. دسترسی در سطح سرور و ابزار تعیین می‌شود."
        link={{ href: '/features/mcp-gateway', label: 'درگاه MCP' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="دسترسی به CRM" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از دسترسی تیم‌ها به ابزارهای یک سرور MCP.">
              <MockRows
                rows={[
                  { icon: Wrench, title: 'خواندن مشتری', text: 'تیم فروش و پشتیبانی', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Wrench, title: 'ثبت کار بعدی', text: 'فقط تیم فروش', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: KeyRound, title: 'اعتبارنامه', text: 'فقط در گدارAI', badge: { label: 'پنهان', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <ClosingBand
        title="ابزارهای تیمتان را وصل کنید."
        titleMuted="دستیار بقیه را می‌داند."
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />
    </main>
  );
}
