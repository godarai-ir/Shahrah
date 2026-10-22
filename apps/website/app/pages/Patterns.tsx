import { Database, KanbanSquare, NotebookText } from 'lucide-react';
import {
  ClosingBand,
  ComparisonTable,
  ControlChecklist,
  FAQ,
  FeatureRow,
  PageHero,
  PillarCrossLinks,
  PillarGrid,
  ProductStage,
  ProductWindow,
  Section,
  StepList,
  ToolSteps,
  TrustStrip,
} from '@/app/components/patterns';
import { sectionHeading } from '@/app/components/sectionType';
import { pillars } from '@/app/lib/pillars';

const sampleSteps = [
  { icon: Database, name: 'CRM', text: 'سابقه‌ی مشتری' },
  { icon: KanbanSquare, name: 'Jira', text: 'تسک‌های باز' },
  { icon: NotebookText, name: 'Notion', text: 'سندهای تیم' },
];

function SampleWindow() {
  return (
    <ProductWindow title="دستیار نمونه" meta="نمونه" label="نمونه‌ی یک پنجره‌ی محصول با سه ابزار و یک نتیجه.">
      <ToolSteps steps={sampleSteps} result={{ title: 'کار انجام شد', text: 'نتیجه‌ی نمونه' }} />
    </ProductWindow>
  );
}

function PatternLabel({ name }: { name: string }) {
  return (
    <div className="mx-auto w-full max-w-6xl px-6 pt-16">
      <p dir="ltr" className="m-0 w-fit rounded-full bg-site-panel-deep px-3 py-1 font-mono text-xs text-site-ink-soft">
        {name}
      </p>
    </div>
  );
}

// Every shared pattern on one page, with each differentiator tone. See docs/site-architecture/03-shared-patterns.md.
export function Patterns() {
  return (
    <main className="bg-site-ground">
      <PatternLabel name="PageHero + ProductStage + ProductWindow + ToolSteps" />
      <PageHero
        title="عنوان نمونه‌ی صفحه"
        lead="یک یا دو جمله‌ی کوتاه که پرسش اصلی بازدیدکننده را جواب می‌دهد."
        primary={{ href: '/onboarding', label: 'اقدام اصلی' }}
        secondary={{ href: '/demo', label: 'اقدام دوم' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <SampleWindow />
          </ProductStage>
        }
      />

      <PatternLabel name="ProductStage tones" />
      <Section layout="grid grid-cols-1 gap-6 md:grid-cols-2">
        {pillars.map((pillar) => (
          <ProductStage key={pillar.id} tone={pillar.id}>
            <ProductWindow title={pillar.title} meta="نمونه" media="teamSupport" />
          </ProductStage>
        ))}
      </Section>

      <PatternLabel name="FeatureRow (tone=guard, reverse)" />
      <FeatureRow
        tone="guard"
        reverse
        title="عنوان یک ایده"
        copy="متن کوتاه که همان یک ایده را توضیح می‌دهد."
        bullets={['نکته‌ی اول', 'نکته‌ی دوم']}
        link={{ href: '/features/guardrails', label: 'پیوند به صفحه‌ی مرتبط' }}
        visual={
          <ProductStage tone="ground" layout="w-full max-w-lg p-8">
            <SampleWindow />
          </ProductStage>
        }
      />

      <PatternLabel name="PillarGrid (numbered) · PillarGrid compact" />
      <Section layout="flex flex-col gap-10">
        <PillarGrid numbered />
        <PillarGrid columns={2} size="compact" />
      </Section>

      <PatternLabel name="ControlChecklist" />
      <Section>
        <ControlChecklist
          groups={[
            { title: 'امنیت', surface: 'guard', items: [{ label: 'گاردریل معنایی', href: '/features/guardrails' }, { label: 'بررسی فایل‌ها' }] },
            { title: 'حاکمیت', surface: 'govern', items: [{ label: 'نقش و دسترسی تیمی', href: '/features/governance' }, { label: 'ثبت تغییر تنظیمات' }] },
            { title: 'هزینه', surface: 'cost', items: [{ label: 'سقف برای هر تیم', href: '/features/cost-control' }, { label: 'پرداخت تومانی' }] },
          ]}
        />
      </Section>

      <PatternLabel name="StepList · TrustStrip" />
      <Section layout="flex flex-col gap-16">
        <StepList
          steps={[
            { title: 'قدم اول', text: 'توضیح کوتاه قدم اول.' },
            { title: 'قدم دوم', text: 'توضیح کوتاه قدم دوم.' },
            { title: 'قدم سوم', text: 'توضیح کوتاه قدم سوم.' },
          ]}
        />
        <TrustStrip />
      </Section>

      <PatternLabel name="ComparisonTable" />
      <Section layout="flex flex-col gap-8">
        <h2 className={sectionHeading}>مقایسه‌ی نمونه</h2>
        <ComparisonTable
          caption="مقایسه‌ی نمونه"
          columns={['گزینه‌ی الف', 'گدارAI']}
          highlight={1}
          rows={[
            { label: 'پرداخت', values: ['ارزی', 'تومانی'] },
            { label: 'دسترسی تیمی', values: ['ندارد', 'دارد'], note: 'یادداشت کوتاه برای این ردیف' },
          ]}
        />
      </Section>

      <PatternLabel name="FAQ (groups)" />
      <FAQ
        title="پرسش‌های نمونه"
        id="patterns-faq"
        groups={[
          { title: 'امنیت', items: [{ question: 'پرسش نمونه؟', answer: 'پاسخ نمونه.' }] },
          { title: 'هزینه', items: [{ question: 'پرسش نمونه‌ی دوم؟', answer: 'پاسخ نمونه‌ی دوم.' }] },
        ]}
      />

      <PatternLabel name="PillarCrossLinks (current=guard)" />
      <PillarCrossLinks current="guard" />

      <PatternLabel name="ClosingBand" />
      <ClosingBand
        title="عنوان پایانی"
        titleMuted="ادامه‌ی کم‌رنگ عنوان."
        copy="یک جمله‌ی کوتاه پیش از اقدام."
        primary={{ href: '/onboarding', label: 'اقدام اصلی' }}
        secondary={{ href: '/demo', label: 'اقدام دوم' }}
      />
    </main>
  );
}
