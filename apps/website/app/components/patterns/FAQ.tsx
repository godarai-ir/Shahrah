import { ChevronDown } from 'lucide-react';
import { sectionHeading } from '../sectionType';
import { Section } from './Section';

export type FAQItem = { question: string; answer: string };
export type FAQGroup = { title: string; items: FAQItem[] };

function FAQList({ items }: { items: FAQItem[] }) {
  return (
    <div className="divide-y divide-site-line border-y border-site-line">
      {items.map((item) => (
        <details key={item.question} className="group py-1">
          <summary className="flex min-h-14 cursor-pointer list-none items-center justify-between gap-4 rounded-xl py-3 text-lg font-medium text-site-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-accent [&::-webkit-details-marker]:hidden">
            {item.question}
            <ChevronDown className="size-5 shrink-0 text-site-ink-soft transition-transform duration-500 group-open:rotate-180" />
          </summary>
          <p className="pb-5 pe-9 text-base leading-8 text-site-ink-soft">{item.answer}</p>
        </details>
      ))}
    </div>
  );
}

type FAQProps = {
  title: string;
  id?: string;
} & ({ items: FAQItem[]; groups?: never } | { groups: FAQGroup[]; items?: never });

// Native disclosure widgets: keyboard and screen-reader support without script. Groups split fears from features.
export function FAQ({ title, id = 'faq', items, groups }: FAQProps) {
  return (
    <Section id={id} width="narrow">
      <div>
        <h2 className={`mb-12 ${sectionHeading}`}>{title}</h2>
      </div>
      {groups ? (
        <div className="flex flex-col gap-12">
          {groups.map((group) => (
            <div key={group.title} className="flex flex-col gap-4">
              <h3 className="text-xl font-semibold text-site-ink">{group.title}</h3>
              <FAQList items={group.items} />
            </div>
          ))}
        </div>
      ) : (
        <FAQList items={items} />
      )}
    </Section>
  );
}
