import * as React from "react";

import { AppCard } from "@/components/app-ui/AppCard";
import { cn } from "@/lib/utils";

type SettingsCardProps = {
  title?: React.ReactNode;
  description?: React.ReactNode;
  actions?: React.ReactNode;
  footer?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
};

export function SettingsCard({ title, description, actions, footer, children, className }: SettingsCardProps) {
  return (
    <AppCard title={title} description={description} actions={actions} footer={footer} className={cn("max-w-4xl", className)}>
      <div className="flex flex-col gap-4">{children}</div>
    </AppCard>
  );
}
