import { createTranslator } from "next-intl";
import { describe, expect, it } from "vitest";

import en from "@/messages/modules/oidcClients/en.json";
import fa from "@/messages/modules/oidcClients/fa.json";

type MessageTree = { [key: string]: string | MessageTree };

function leafKeys(tree: MessageTree, prefix = ""): Array<[string, string]> {
  return Object.entries(tree).flatMap(([key, value]) =>
    typeof value === "string" ? [[`${prefix}${key}`, value] as [string, string]] : leafKeys(value, `${prefix}${key}.`),
  );
}

describe("oidcClients messages", () => {
  // Component tests mock useTranslations, so they cannot catch ICU syntax errors such as an
  // angle-bracketed placeholder being parsed as an unclosed rich-text tag.
  for (const [locale, messages] of [["en", en], ["fa", fa]] as const) {
    it(`formats every ${locale} message with the real translator`, () => {
      const errors: string[] = [];
      const t = createTranslator({
        locale,
        messages: { oidcClients: messages },
        namespace: "oidcClients",
        onError: (error) => errors.push(`${error.code}: ${error.message}`),
      });

      for (const [key, source] of leafKeys(messages as MessageTree)) {
        const values = Object.fromEntries([...source.matchAll(/\{(\w+)\}/g)].map(([, name]) => [name, "x"]));
        (t as (key: string, values?: Record<string, string>) => string)(key, values);
      }

      expect(errors).toEqual([]);
    });
  }

  it("keeps both locales structurally aligned", () => {
    expect(leafKeys(fa as MessageTree).map(([key]) => key)).toEqual(leafKeys(en as MessageTree).map(([key]) => key));
  });
});
