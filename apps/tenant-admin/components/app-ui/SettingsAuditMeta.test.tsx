import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SettingsAuditMeta } from "./SettingsAuditMeta";

describe("SettingsAuditMeta", () => { it("renders update metadata with fallbacks", () => { render(<SettingsAuditMeta updatedAt={null} updatedBy="Admin" updatedAtLabel="Updated" updatedByLabel="By" fallback="Not set" />); expect(screen.getByText("Not set")).toBeInTheDocument(); expect(screen.getByText("Admin")).toBeInTheDocument(); }); });
