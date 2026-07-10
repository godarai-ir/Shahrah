import { BookOpen } from "lucide-react";
import { useTranslations } from "next-intl";
import * as React from "react";

import { cn } from "@/lib/utils";

type HelpLinkProps = React.AnchorHTMLAttributes<HTMLAnchorElement> & {
  href: string;
};

export function HelpLink({ href, className, target = "_blank", rel, children, ...props }: HelpLinkProps) {
  const t = useTranslations("common");
  const resolvedRel = target === "_blank" ? (rel ?? "noreferrer") : rel;

  return (
    <a
      href={href}
      target={target}
      rel={resolvedRel}
      className={cn(
        "inline-flex h-8 shrink-0 items-center gap-1.5 rounded-md border border-border bg-background px-2.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        className,
      )}
      {...props}
    >
      <BookOpen className="size-3.5" aria-hidden="true" />
      <span>{children ?? t("usageGuide")}</span>
    </a>
  );
}
