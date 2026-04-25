export default function Home() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-stone-950 px-6 text-stone-50">
      <div className="w-full max-w-3xl rounded-3xl border border-stone-800 bg-stone-900 p-10 shadow-2xl shadow-stone-950/40">
        <p className="text-sm uppercase tracking-[0.3em] text-amber-300">
          Shahrah
        </p>
        <h1 className="mt-4 text-4xl font-semibold tracking-tight">
          Support Admin
        </h1>
        <p className="mt-4 max-w-2xl text-base leading-7 text-stone-300">
          This app will become the internal operations panel for support,
          customer triage, billing review, and synchronized tenant data.
        </p>
      </div>
    </main>
  );
}
