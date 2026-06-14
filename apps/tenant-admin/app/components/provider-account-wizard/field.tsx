"use client";

import type React from "react";

import { Label } from "@/components/ui/label";
export function Field(props: { label: string; id: string; children: React.ReactNode }) {
  return (
    <div className="grid gap-2">
      <Label htmlFor={props.id}>{props.label}</Label>
      {props.children}
    </div>
  );
}


