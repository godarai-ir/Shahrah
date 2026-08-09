"use client";

import * as React from "react";

import type { CodeLanguage } from "@/components/app-ui/CodeBlock";
import { CodePanel, type CodePanelProps } from "@/components/app-ui/CodePanel";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

export type CodePanelTab = {
  value: string;
  label: React.ReactNode;
  code: string;
  language: CodeLanguage;
  className?: string;
};

export type TabbedCodePanelProps = Omit<
  CodePanelProps,
  "className" | "code" | "displayLanguage" | "language"
> & {
  tabs: readonly CodePanelTab[];
  tabsLabel: string;
  defaultValue?: string;
};

export function TabbedCodePanel({
  tabs,
  tabsLabel,
  defaultValue,
  actions,
  ...panelProps
}: TabbedCodePanelProps) {
  const [selectedValue, setSelectedValue] = React.useState(defaultValue ?? tabs[0]?.value ?? "");
  const activeTab = tabs.find((tab) => tab.value === selectedValue) ?? tabs[0];

  if (!activeTab) return null;

  return (
    <Tabs value={activeTab.value} onValueChange={setSelectedValue} className="min-w-0">
      <CodePanel
        {...panelProps}
        code={activeTab.code}
        language={activeTab.language}
        className={activeTab.className}
        actions={
          <>
            <TabsList aria-label={tabsLabel}>
              {tabs.map((tab) => (
                <TabsTrigger key={tab.value} value={tab.value}>
                  {tab.label}
                </TabsTrigger>
              ))}
            </TabsList>
            {actions}
          </>
        }
      />
    </Tabs>
  );
}
