"use client";

import { createContext, useContext } from "react";

const PermissionsContext = createContext<string[] | null>(null);

export function PermissionsProvider({
  permissions,
  children,
}: {
  permissions: string[];
  children: React.ReactNode;
}) {
  return (
    <PermissionsContext.Provider value={permissions}>
      {children}
    </PermissionsContext.Provider>
  );
}

export function usePermissions(): string[] | null {
  return useContext(PermissionsContext);
}

