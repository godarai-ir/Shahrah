"use client";

import {
  createContext,
  useCallback,
  useContext,
  useLayoutEffect,
  useMemo,
  useState,
} from "react";

import type { ShellPageMetadata } from "@/components/app-ui/ShellPageHeading";

type ContentPadding = "default" | "none";

type AdminShellPageState = {
  contentPadding: ContentPadding;
  pageHeader?: ShellPageMetadata;
};

type AdminShellPageContextValue = AdminShellPageState & {
  setPageState: (state: AdminShellPageState) => void;
};

const DEFAULT_PAGE_STATE: AdminShellPageState = { contentPadding: "default" };
const AdminShellPageContext = createContext<AdminShellPageContextValue | null>(null);

export function AdminShellPageProvider({ children }: { children: React.ReactNode }) {
  const [pageState, setPageStateValue] = useState<AdminShellPageState>(DEFAULT_PAGE_STATE);
  const setPageState = useCallback((state: AdminShellPageState) => setPageStateValue(state), []);
  const value = useMemo(
    () => ({ ...pageState, setPageState }),
    [pageState, setPageState],
  );

  return (
    <AdminShellPageContext.Provider value={value}>
      {children}
    </AdminShellPageContext.Provider>
  );
}

export function AdminShellPageBridge({
  children,
  pageHeader,
  contentPadding = "default",
}: {
  children: React.ReactNode;
  contentPadding?: ContentPadding;
  pageHeader?: ShellPageMetadata;
}) {
  const { setPageState } = useAdminShellPage();
  const title = pageHeader?.title;
  const description = pageHeader?.description;
  const helpHref = pageHeader?.helpHref;

  useLayoutEffect(() => {
    setPageState({
      contentPadding,
      pageHeader: title ? { title, description, helpHref } : undefined,
    });

    return () => setPageState(DEFAULT_PAGE_STATE);
  }, [contentPadding, description, helpHref, setPageState, title]);

  return children;
}

export function useAdminShellPage() {
  const context = useContext(AdminShellPageContext);
  if (!context) throw new Error("Admin shell page context is unavailable.");
  return context;
}
