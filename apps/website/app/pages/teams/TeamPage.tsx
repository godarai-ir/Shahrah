import { ClosingBand, PageHero, PillarGrid, ProductStage, ProductWindow, Section, TileGrid, ToolSteps } from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';
import { pillars } from '@/app/lib/pillars';
import { teams } from '@/app/lib/teams';
import type { Team } from '@/app/lib/teams';
import { CheckCircle2 } from 'lucide-react';

// One template for the six team pages; content comes from lib/teams.ts.
export function TeamPage({ team }: { team: Team }) {
  const related = pillars.filter((pillar) => team.pillars.includes(pillar.id));
  const others = teams.filter((item) => item.id !== team.id);

  return (
    <main className="bg-site-ground">
      <PageHero
        title={team.title}
        lead={team.lead}
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <div className="relative w-full max-w-sm">
              <ProductWindow title={`دستیار ${team.label}`} meta="نمونه" label={`نمونه‌ای از کار دستیار تیم ${team.label}: ${team.result.title}.`}>
                <ToolSteps steps={team.steps.map((step) => ({ icon: step.icon, name: step.title, text: step.text }))} result={team.result} />
              </ProductWindow>
            </div>
          </ProductStage>
        }
      />

      <Section id="tasks" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>سه کاری که اول سپرده می‌شود.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>کارهای تکراری با دستیار؛ تصمیم با همکارانتان.</p>
        </div>
        <TileGrid columns={3} items={team.tasks.map((task) => ({ key: task.title, icon: CheckCircle2, title: task.title, text: task.text }))} />
      </Section>

      <Section id="tools" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>{`ابزارهایی که دستیار ${team.label} می‌شناسد.`}</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>همان ابزارهایی که تیم هر روز باز می‌کند.</p>
        </div>
        <TileGrid
          columns={team.tools.length >= 3 ? 3 : 2}
          size="compact"
          items={team.tools.map((tool) => ({ key: tool.title, icon: tool.icon, title: tool.title, text: tool.text, href: '/integrations' }))}
        />
      </Section>

      <Section id="controls" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>کنترل‌هایی که این تیم لازم دارد.</h2>
        </div>
        <PillarGrid items={related} columns={2} size="compact" />
      </Section>

      <Section id="other-teams" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>تیم‌های دیگر.</h2>
        </div>
        <TileGrid
          columns={3}
          size="compact"
          items={others.map((item) => ({ key: item.id, icon: item.icon, title: `تیم ${item.label}`, text: item.tasks[0].text, href: `/teams/${item.id}` }))}
        />
      </Section>

      <ClosingBand
        title={`دستیار تیم ${team.label} را امروز بسازید.`}
        titleMuted="فردا کار سبک‌تر است."
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />
    </main>
  );
}
