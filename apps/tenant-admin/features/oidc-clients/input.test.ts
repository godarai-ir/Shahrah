import { describe, expect, it } from "vitest";

import { isValidRedirectURI, parseOIDCClientInput, splitURIList } from "./input";

describe("OIDC client input", () => {
  it("normalizes a create payload and defaults post-logout URIs", () => {
    const result = parseOIDCClientInput(
      { name: "  LibreChat ", redirect_uris: ["https://chat.example.com/cb", "https://chat.example.com/cb", " "] },
      { partial: false },
    );

    expect(result).toEqual({
      ok: true,
      value: { name: "LibreChat", redirect_uris: ["https://chat.example.com/cb"], post_logout_redirect_uris: [] },
    });
  });

  it("requires a name and at least one redirect URI on create", () => {
    expect(parseOIDCClientInput({ name: "", redirect_uris: ["https://a.example/cb"] }, { partial: false })).toMatchObject({
      ok: false,
      code: "name_required",
      field: "name",
    });
    expect(parseOIDCClientInput({ name: "App", redirect_uris: [] }, { partial: false })).toMatchObject({
      ok: false,
      code: "redirect_required",
      field: "redirect_uris",
    });
  });

  it("names the offending URI", () => {
    expect(
      parseOIDCClientInput({ name: "App", redirect_uris: ["https://ok.example/cb", "http://chat.example.com/cb"] }, { partial: false }),
    ).toMatchObject({ ok: false, code: "redirect_invalid", value: "http://chat.example.com/cb" });
  });

  it("accepts https everywhere and http only for development hosts", () => {
    expect(isValidRedirectURI("https://chat.example.com/oauth/openid/callback")).toBe(true);
    expect(isValidRedirectURI("http://localhost:3080/oauth/openid/callback")).toBe(true);
    expect(isValidRedirectURI("http://acme.localhost:3080/cb")).toBe(true);
    expect(isValidRedirectURI("http://127.0.0.1:3080/cb")).toBe(true);
    expect(isValidRedirectURI("http://chat.example.com/cb")).toBe(false);
    expect(isValidRedirectURI("https://chat.example.com/cb#frag")).toBe(false);
    expect(isValidRedirectURI("https://user:pass@chat.example.com/cb")).toBe(false);
    expect(isValidRedirectURI("/relative/cb")).toBe(false);
    expect(isValidRedirectURI("javascript:alert(1)")).toBe(false);
  });

  it("caps the number of URIs", () => {
    const uris = Array.from({ length: 11 }, (_, index) => `https://app${index}.example/cb`);
    expect(parseOIDCClientInput({ name: "App", redirect_uris: uris }, { partial: false })).toMatchObject({
      code: "redirect_too_many",
    });
  });

  it("keeps only present fields on update and validates the active flag", () => {
    expect(parseOIDCClientInput({ active: false }, { partial: true })).toEqual({ ok: true, value: { active: false } });
    expect(parseOIDCClientInput({ active: "no" }, { partial: true })).toMatchObject({ ok: false, code: "invalid_payload" });
    expect(parseOIDCClientInput({}, { partial: true })).toMatchObject({ ok: false, code: "invalid_payload" });
    expect(parseOIDCClientInput({ active: true }, { partial: false })).toMatchObject({ ok: false, code: "name_required" });
  });

  it("rejects non-object payloads and non-string URI lists", () => {
    expect(parseOIDCClientInput(null, { partial: false })).toMatchObject({ ok: false, code: "invalid_payload" });
    expect(parseOIDCClientInput({ name: "App", redirect_uris: "https://a.example/cb" }, { partial: false })).toMatchObject({
      ok: false,
      code: "invalid_payload",
    });
  });

  it("splits textarea input into unique trimmed lines", () => {
    expect(splitURIList(" https://a.example/cb \r\n\nhttps://b.example/cb\nhttps://a.example/cb")).toEqual([
      "https://a.example/cb",
      "https://b.example/cb",
    ]);
  });
});
