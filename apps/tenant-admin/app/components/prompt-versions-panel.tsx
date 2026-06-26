"use client";

import { Clipboard, FileCode2, History } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

import type { PromptTranslationFn, PromptVersion } from "./prompt-management-shared";

type PromptVersionsPanelProps = {
  isLoading: boolean;
  onCopyFqn: () => void;
  onCopySnippet: () => void;
  onSelectVersion: (version: PromptVersion) => void;
  selectedVersion: PromptVersion | null;
  selectedVersionID: string;
  t: PromptTranslationFn;
  versions: PromptVersion[];
};

export function PromptVersionsPanel(props: PromptVersionsPanelProps) {
  return (
    <Card className="shadow-none">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          <History className="size-4" />
          {props.t("versions.title")}
        </CardTitle>
        <CardDescription>{props.t("versions.description")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {props.isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : props.versions.length === 0 ? (
          <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">{props.t("versions.empty")}</div>
        ) : (
          <div className="space-y-2">
            {props.versions.map((version) => (
              <button
                key={version.id}
                type="button"
                className={cn(
                  "w-full rounded-md border p-3 text-start transition hover:bg-accent",
                  version.id === props.selectedVersionID && "border-primary/40 bg-primary/5",
                )}
                onClick={() => props.onSelectVersion(version)}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="font-medium">{props.t("versions.versionValue", { version: version.version })}</span>
                  <Badge variant="secondary">{version.model || version.manifest?.model || "-"}</Badge>
                </div>
                <div className="mt-1 truncate text-xs text-muted-foreground" dir="ltr">
                  {version.commit_message || version.fqn}
                </div>
              </button>
            ))}
          </div>
        )}
        {props.selectedVersion ? (
          <div className="rounded-lg border bg-muted/20 p-3">
            <div className="text-xs text-muted-foreground">{props.t("versions.selectedFqn")}</div>
            <div className="mt-1 font-mono text-xs" dir="ltr">
              {props.selectedVersion.fqn}
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button type="button" variant="outline" size="sm" onClick={props.onCopyFqn}>
                <Clipboard />
                {props.t("actions.copyFqn")}
              </Button>
              <Button type="button" variant="outline" size="sm" onClick={props.onCopySnippet}>
                <FileCode2 />
                {props.t("actions.copySnippet")}
              </Button>
            </div>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
