"use client";

import * as React from "react";
import Link from "next/link";
import { MoreVertical } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

export type RowAction = {
  key?: React.Key;
  label: React.ReactNode;
  onSelect?: () => void;
  href?: string;
  external?: boolean;
  disabled?: boolean;
  pending?: boolean;
  pendingLabel?: React.ReactNode;
  destructive?: boolean;
  separatorBefore?: boolean;
  icon?: React.ReactNode;
};

export type RowActionsMenuProps = {
  actions: RowAction[];
  label: string;
  align?: "start" | "center" | "end";
  className?: string;
  disabled?: boolean;
  dir?: "rtl" | "ltr";
  triggerSize?: "icon" | "icon-sm" | "icon-xs";
  stopPropagation?: boolean;
};

export function RowActionsMenu({
  actions,
  label,
  align = "end",
  className,
  disabled,
  dir = "rtl",
  triggerSize = "icon",
  stopPropagation = false,
}: RowActionsMenuProps) {
  const groups = groupActions(actions);

  return (
    <DropdownMenu dir={dir}>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size={triggerSize}
          aria-label={label}
          className={className}
          disabled={disabled}
          onClick={(event) => {
            if (stopPropagation) event.stopPropagation();
          }}
        >
          <MoreVertical aria-hidden="true" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align={align}
        className="min-w-40 text-xs"
        onClick={(event) => {
          if (stopPropagation) event.stopPropagation();
        }}
      >
        {groups.map((group, groupIndex) => (
          <React.Fragment key={groupIndex}>
            {groupIndex > 0 ? <DropdownMenuSeparator /> : null}
            <DropdownMenuGroup>
              {group.map((action, actionIndex) => (
                <RowActionItem key={action.key ?? action.href ?? actionIndex} action={action} />
              ))}
            </DropdownMenuGroup>
          </React.Fragment>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function RowActionItem({ action }: { action: RowAction }) {
  const isDisabled = Boolean(action.disabled || action.pending);
  const content = (
    <>
      {action.icon ? <span aria-hidden="true">{action.icon}</span> : null}
      <span>{action.pending && action.pendingLabel ? action.pendingLabel : action.label}</span>
    </>
  );
  const className = cn(
    "justify-start gap-2 text-start text-xs [&_svg]:size-3.5",
    action.destructive && "text-destructive focus:text-destructive",
  );

  if (action.href && !isDisabled) {
    return (
      <DropdownMenuItem asChild className={className}>
        {action.external ? (
          <a href={action.href} target="_blank" rel="noreferrer" onClick={action.onSelect}>
            {content}
          </a>
        ) : (
          <Link href={action.href} onClick={action.onSelect}>
            {content}
          </Link>
        )}
      </DropdownMenuItem>
    );
  }

  return (
    <DropdownMenuItem disabled={isDisabled} onSelect={action.onSelect} className={className}>
      {content}
    </DropdownMenuItem>
  );
}

function groupActions(actions: RowAction[]) {
  return actions.reduce<RowAction[][]>((groups, action) => {
    if (groups.length === 0 || (action.separatorBefore && groups.at(-1)?.length)) {
      groups.push([]);
    }
    groups.at(-1)?.push(action);
    return groups;
  }, []);
}
