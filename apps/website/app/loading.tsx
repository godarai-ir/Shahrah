export default function Loading() {
  return (
    <div className="min-h-[40vh] w-full flex items-center justify-center px-4 py-12">
      <div className="text-center">
        <div className="mx-auto h-10 w-10 rounded-full border-4 border-slate-300 border-t-indigo-500 animate-spin" />
        <p className="mt-4 text-sm text-slate-600 dark:text-slate-400">
          در حال بارگذاری...
        </p>
      </div>
    </div>
  );
}
