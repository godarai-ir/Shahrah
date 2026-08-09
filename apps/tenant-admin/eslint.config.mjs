import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    files: ["**/*.ts", "**/*.tsx"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "@/components/app-ui",
              message: "Import from the specific app-ui module to keep route chunks isolated.",
            },
          ],
          patterns: [
            {
              group: ["radix-ui", "radix-ui/*", "@radix-ui/*", "@base-ui/react", "@base-ui/react/*"],
              message: "Feature and product code must consume primitives through components/ui.",
            },
          ],
        },
      ],
    },
  },
  {
    files: ["features/**/*.ts", "features/**/*.tsx"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "@/components/app-ui",
              message: "Import from the specific app-ui module to keep route chunks isolated.",
            },
          ],
          patterns: [
            {
              group: ["@/app/**"],
              message: "Feature code cannot depend on the Next.js app layer.",
            },
            {
              group: ["radix-ui", "radix-ui/*", "@radix-ui/*", "@base-ui/react", "@base-ui/react/*"],
              message: "Feature code must consume primitives through components/ui.",
            },
          ],
        },
      ],
    },
  },
  {
    files: ["components/ui/**/*.ts", "components/ui/**/*.tsx"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "@/components/app-ui",
              message: "UI primitives cannot depend on product-level components.",
            },
          ],
          patterns: [
            {
              group: ["@/app/**", "@/features/**"],
              message: "UI primitives cannot depend on app or feature code.",
            },
          ],
        },
      ],
    },
  },
  {
    files: ["components/app-ui/**/*.ts", "components/app-ui/**/*.tsx", "hooks/**/*.ts", "hooks/**/*.tsx", "lib/**/*.ts", "lib/**/*.tsx"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "@/components/app-ui",
              message: "Import from the specific app-ui module to keep route chunks isolated.",
            },
          ],
          patterns: [
            {
              group: ["@/app/**", "@/features/**"],
              message: "Shared code cannot depend on app or feature code.",
            },
            {
              group: ["radix-ui", "radix-ui/*", "@radix-ui/*", "@base-ui/react", "@base-ui/react/*"],
              message: "Product-level shared code must consume primitives through components/ui.",
            },
          ],
        },
      ],
    },
  },
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
