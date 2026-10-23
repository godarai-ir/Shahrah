// Monospace fonts have no Persian glyphs, so each line keeps the site font and its own direction: Persian comments
// read right to left inside left-to-right code.
const lineStyle = { unicodeBidi: 'plaintext', textAlign: 'left' } as const;

type CodeWindowProps = {
  title: string;
  /** Right of the title bar, such as the language or file name. */
  meta?: string;
  code: string;
};

// A code sample framed like a product window: a surface title bar over a navy body.
export function CodeWindow({ title, meta, code }: CodeWindowProps) {
  return (
    <div className="w-full max-w-md overflow-hidden rounded-card border border-site-line bg-site-surface">
      <div className="flex items-center justify-between gap-3 border-b border-site-line px-5 py-4">
        <span className="truncate text-label font-semibold text-site-ink">{title}</span>
        {meta ? <span dir="ltr" className="shrink-0 text-caption text-site-ink-soft">{meta}</span> : null}
      </div>
      <pre dir="ltr" className="m-0 overflow-x-auto bg-site-night p-5 text-caption text-site-night-soft sm:text-label">
        <code>
          {code.split('\n').map((line, index) => (
            // Lines are static and never reorder, so the index is a stable key; a blank line keeps its height.
            <span key={index} className="block whitespace-pre" style={lineStyle}>
              {line || ' '}
            </span>
          ))}
        </code>
      </pre>
    </div>
  );
}
