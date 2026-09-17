"use client";

import { useEffect, useMemo, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { usePermissions } from "@/components/permissions-context";
import { matchesSearch } from "@/lib/search-normalize";

import { buildPaletteCommands, type PaletteCommand } from "./commands";
import { pushRecentCommandID, readRecentCommandIDs, resolveRecentCommands } from "./recent-commands";

type CommandPaletteProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

function filterCommand(value: string, search: string, keywords?: string[]): number {
  return matchesSearch([value, ...(keywords ?? [])], search) ? 1 : 0;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const t = useTranslations("shell");
  const pathname = usePathname();
  const router = useRouter();
  const permissions = usePermissions();
  const [search, setSearch] = useState("");
  const [recentIDs, setRecentIDs] = useState<string[]>([]);

  const commands = useMemo(
    () => buildPaletteCommands(pathname, t, permissions ?? []),
    [pathname, t, permissions],
  );

  // Read storage after mount, on each open, so server and client render identical markup.
  useEffect(() => {
    if (!open) return;
    setSearch("");
    setRecentIDs(readRecentCommandIDs());
  }, [open]);

  const recentCommands = useMemo(() => resolveRecentCommands(recentIDs, commands.byId), [recentIDs, commands]);
  const showRecent = search.trim() === "" && recentCommands.length > 0;

  function runCommand(command: PaletteCommand) {
    pushRecentCommandID(command.id);
    onOpenChange(false);
    router.push(command.href);
  }

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("palette.dialogTitle")}
      description={t("palette.dialogDescription")}
      commandProps={{ filter: filterCommand, loop: true }}
    >
      <CommandInput value={search} onValueChange={setSearch} placeholder={t("palette.placeholder")} />
      <CommandList className="max-h-[min(24rem,60vh)]">
        <CommandEmpty>
          <p>{t("palette.empty")}</p>
          <p className="mt-1 text-xs text-muted-foreground">{t("palette.emptyHint")}</p>
        </CommandEmpty>
        {showRecent ? (
          <CommandGroup heading={t("palette.groups.recent")}>
            {recentCommands.map((command) => (
              <PaletteItem key={command.id} command={command} valuePrefix="recent" onSelect={runCommand} />
            ))}
          </CommandGroup>
        ) : null}
        {commands.navigate.length > 0 ? (
          <CommandGroup heading={t("palette.groups.navigate")}>
            {commands.navigate.map((command) => (
              <PaletteItem key={command.id} command={command} onSelect={runCommand} />
            ))}
          </CommandGroup>
        ) : null}
        {commands.action.length > 0 ? (
          <CommandGroup heading={t("palette.groups.action")}>
            {commands.action.map((command) => (
              <PaletteItem key={command.id} command={command} onSelect={runCommand} />
            ))}
          </CommandGroup>
        ) : null}
      </CommandList>
    </CommandDialog>
  );
}

function PaletteItem({
  command,
  valuePrefix,
  onSelect,
}: {
  command: PaletteCommand;
  /** cmdk tracks selection by value, so a command shown in two groups needs distinct values. */
  valuePrefix?: string;
  onSelect: (command: PaletteCommand) => void;
}) {
  const Icon = command.icon;
  const keywords = [command.title, ...(command.sectionLabel ? [command.sectionLabel] : []), ...(command.keywords ?? [])];

  return (
    <CommandItem
      value={valuePrefix ? `${valuePrefix}:${command.id}` : command.id}
      keywords={keywords}
      onSelect={() => onSelect(command)}
      data-command-id={command.id}
    >
      <Icon className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
      <span className="min-w-0 truncate">{command.title}</span>
      {command.sectionLabel ? (
        <span className="ms-auto shrink-0 truncate text-xs text-muted-foreground">{command.sectionLabel}</span>
      ) : null}
    </CommandItem>
  );
}
