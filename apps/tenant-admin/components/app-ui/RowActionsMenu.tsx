"use client";

import * as React from "react";
import { MoreVertical } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

type RowAction = {
  label: React.ReactNode;
  onSelect?: () => void;
  disabled?: boolean;
  destructive?: boolean;
  icon?: React.ReactNode;
};

type RowActionsMenuProps = {
  actions: RowAction[];
  label: string;
  align?: "start" | "center" | "end";
  className?: string;
  disabled?: boolean;
};

export function RowActionsMenu({ actions, label, align = "end", className, disabled }: RowActionsMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button type="button" variant="ghost" size="icon" aria-label={label} className={className} disabled={disabled}>
          <MoreVertical aria-hidden="true" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align={align} className="min-w-40">
        {actions.map((action, index) => (
          <DropdownMenuItem
            key={index}
            disabled={action.disabled}
            onSelect={action.onSelect}
            className={cn("gap-2", action.destructive && "text-destructive focus:text-destructive")}
          >
            {action.icon ? <span className="flex size-4 items-center justify-center [&_svg]:size-4" aria-hidden="true">{action.icon}</span> : null}
            <span>{action.label}</span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
