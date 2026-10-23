import { readFileSync } from "node:fs";
import { join } from "node:path";

import { MarkdownDocument } from "@/app/components/MarkdownDocument";

const privacyContent = readFileSync(
  join(process.cwd(), "privacy-policy.md"),
  "utf8",
);

export function Privacy() {
  return (
    <main className="bg-site-ground px-4 py-10 sm:px-6 lg:py-20">
      <article dir="rtl" className="mx-auto flex max-w-2xl flex-col gap-8 text-right">
        <h1 className="m-0 text-balance text-h2 font-extrabold text-site-ink lg:text-h1">
          سیاست حفظ حریم خصوصی
        </h1>
        <div>
          <MarkdownDocument content={privacyContent} />
        </div>
      </article>
    </main>
  );
}
