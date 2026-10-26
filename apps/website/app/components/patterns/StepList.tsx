export type Step = { title: string; text: string };

const columnClass = { 3: 'md:grid-cols-3', 4: 'md:grid-cols-2 lg:grid-cols-4' } as const;

// A numbered sequence where order carries meaning; the one place section numbering is allowed.
export function StepList({ steps }: { steps: Step[] }) {
  const columns = steps.length >= 4 ? 4 : 3;

  return (
    <ol className={`m-0 grid list-none gap-6 p-0 sm:gap-8 ${columnClass[columns]}`}>
      {steps.map((step, index) => (
        <li key={step.title}>
          <p className="mb-2 text-h2 font-extrabold text-site-accent-text">{(index + 1).toLocaleString('fa-IR')}</p>
          <h3 className="mb-2 text-h3 font-semibold text-site-ink">{step.title}</h3>
          <p className="text-body text-site-ink-soft">{step.text}</p>
        </li>
      ))}
    </ol>
  );
}
