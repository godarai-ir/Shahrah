export default function Home() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-950 px-6 text-slate-50">
      <div className="w-full max-w-3xl rounded-3xl border border-slate-800 bg-slate-900 p-10 shadow-2xl shadow-slate-950/40">
        <p className="text-sm uppercase tracking-[0.3em] text-cyan-300">
          Shahrah
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-tight">
          Tenant Admin
        </h1>
        <p className="mt-4 max-w-2xl text-base leading-7 text-slate-300">
          This app will become the tenant-facing dashboard for API keys, usage,
          billing, routing, and settings at{" "}
          <code className="rounded bg-slate-800 px-2 py-1 text-sm">
            &lt;tenant&gt;.app.sharah.ir
          </code>
          .
        </p>
      </div>
    </main>
  );
}
