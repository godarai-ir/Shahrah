import { readFileSync } from "node:fs";
import { join } from "node:path";

import { MarkdownDocument } from "@/app/components/MarkdownDocument";
import { ScrollArea } from "@/app/components/ui/scroll-area";

const privacyContent = readFileSync(
  join(process.cwd(), "privacy-policy.md"),
  "utf8",
);

export function Privacy() {
  return (
    <main className="px-4 py-24 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-4xl flex-col gap-8">
        <h1 className="text-3xl font-bold tracking-tight text-foreground md:text-4xl">
          سیاست حفظ حریم خصوصی
        </h1>

        <ScrollArea
          dir="rtl"
          className="h-[72dvh] rounded-md border border-border bg-card"
        >
          <article dir="rtl" className="p-6 text-right">
            <MarkdownDocument content={privacyContent} />
          </article>
        </ScrollArea>
      </div>
    </main>
  );
}
