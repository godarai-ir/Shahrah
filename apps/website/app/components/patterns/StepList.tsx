export type Step = { title: string; text: string };

const columnClass = { 3: 'md:grid-cols-3', 4: 'md:grid-cols-2 lg:grid-cols-4' } as const;

// A numbered sequence where order carries meaning; the one place section numbering is allowed.
export function StepList({ steps }: { steps: Step[] }) {
  const columns = steps.length >= 4 ? 4 : 3;

  return (
    <ol className={`m-0 grid list-none gap-10 p-0 ${columnClass[columns]}`}>
      {steps.map((step, index) => (
        <li key={step.title}>
          <p className="mb-3 text-2xl font-medium text-site-accent">{(index + 1).toLocaleString('fa-IR')}</p>
          <h3 className="mb-2 text-xl font-semibold text-site-ink">{step.title}</h3>
          <p className="text-base leading-8 text-site-ink-soft">{step.text}</p>
        </li>
      ))}
    </ol>
  );
}
