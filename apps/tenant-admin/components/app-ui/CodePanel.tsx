import * as React from "react";

import type { CodeBlockProps } from "@/components/app-ui/CodeBlock";
import { CodeViewer } from "@/components/app-ui/CodeViewer";
import { CopyAction, type CopyActionProps } from "@/components/app-ui/CopyAction";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type CodePanelProps = CodeBlockProps & {
  title: React.ReactNode;
  titleHidden?: boolean;
  description?: React.ReactNode;
  displayLanguage?: React.ReactNode;
  actions?: React.ReactNode;
  copy?: Omit<CopyActionProps, "value">;
  cardClassName?: string;
  contentClassName?: string;
};

export function CodePanel({ title, titleHidden = false, description, displayLanguage, actions, copy, cardClassName, contentClassName, ...codeBlockProps }: CodePanelProps) {
  return (
    <Card className={cn("shadow-none", cardClassName)}>
      <CardHeader className="border-b">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className={cn("flex min-w-0 flex-col gap-1", titleHidden && !description && !displayLanguage && "sr-only")}>
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <CardTitle className={cn("text-base", titleHidden && "sr-only")}>{title}</CardTitle>
              {displayLanguage ? <Badge variant="outline" className="font-mono">{displayLanguage}</Badge> : null}
            </div>
            {description ? <CardDescription>{description}</CardDescription> : null}
          </div>
          {copy || actions ? <div className="flex flex-wrap items-center gap-2">{copy ? <CopyAction value={codeBlockProps.code} showLabel {...copy} /> : null}{actions}</div> : null}
        </div>
      </CardHeader>
      <CardContent className={cn("p-4", contentClassName)}>
        <CodeViewer
          value={codeBlockProps.code}
          language={codeBlockProps.language}
          containerClassName={codeBlockProps.className}
          aria-label={typeof title === "string" ? title : undefined}
        />
      </CardContent>
    </Card>
  );
}
