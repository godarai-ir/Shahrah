import * as React from "react";

import { CodeBlock, type CodeBlockProps } from "@/components/app-ui/CodeBlock";
import { CopyAction, type CopyActionProps } from "@/components/app-ui/CopyAction";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export type CodePanelProps = CodeBlockProps & {
  title: React.ReactNode;
  description?: React.ReactNode;
  displayLanguage?: React.ReactNode;
  actions?: React.ReactNode;
  copy?: Omit<CopyActionProps, "value">;
  cardClassName?: string;
  contentClassName?: string;
};

export function CodePanel({ title, description, displayLanguage, actions, copy, cardClassName, contentClassName, ...codeBlockProps }: CodePanelProps) {
  return (
    <Card className={cn("shadow-none", cardClassName)}>
      <CardHeader className="border-b">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex min-w-0 flex-col gap-1">
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <CardTitle className="text-base">{title}</CardTitle>
              {displayLanguage ? <Badge variant="outline" className="font-mono">{displayLanguage}</Badge> : null}
            </div>
            {description ? <CardDescription>{description}</CardDescription> : null}
          </div>
          {copy || actions ? <div className="flex flex-wrap items-center gap-2">{copy ? <CopyAction value={codeBlockProps.code} showLabel {...copy} /> : null}{actions}</div> : null}
        </div>
      </CardHeader>
      <CardContent className={cn("p-4", contentClassName)}>
        <CodeBlock {...codeBlockProps} maxHeightClassName={codeBlockProps.maxHeightClassName ?? "max-h-[min(34rem,60dvh)]"} />
      </CardContent>
    </Card>
  );
}
