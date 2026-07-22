export default function ModelDetailLoading() {
  return (
    <div className="animate-pulse space-y-4" aria-hidden="true">
      <div className="rounded-xl border border-card-border bg-card p-4">
        <div className="h-8 w-72 rounded bg-muted" />
        <div className="mt-3 h-4 w-full max-w-3xl rounded bg-muted" />
        <div className="mt-2 h-4 w-2/3 rounded bg-muted" />
        <div className="mt-4 grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="rounded-lg border border-border bg-background p-3"
            >
              <div className="h-3 w-24 rounded bg-muted" />
              <div className="mt-2 h-5 w-28 rounded bg-muted" />
            </div>
          ))}
        </div>
      </div>
      <div className="grid gap-4 xl:grid-cols-2">
        {Array.from({ length: 8 }).map((_, i) => (
          <div
            key={i}
            className="rounded-xl border border-card-border bg-card p-4"
          >
            <div className="h-5 w-40 rounded bg-muted" />
            <div className="mt-3 space-y-2">
              <div className="h-4 w-full rounded bg-muted" />
              <div className="h-4 w-11/12 rounded bg-muted" />
              <div className="h-4 w-9/12 rounded bg-muted" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}


