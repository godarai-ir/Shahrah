import path from "node:path";
import { configDefaults, defineConfig } from "vitest/config";

const testExclude = [
  ...configDefaults.exclude,
  "lib/auth/**/*.test.ts",
  ".next/**",
];

export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(__dirname),
    },
  },
  test: {
    exclude: testExclude,
    maxWorkers: 3,
    pool: "threads",
    projects: [
      {
        extends: true,
        test: {
          name: "node",
          environment: "node",
          include: ["**/*.test.ts"],
        },
      },
      {
        extends: true,
        test: {
          name: "dom",
          environment: "jsdom",
          include: ["**/*.test.tsx"],
          setupFiles: ["./vitest.setup.ts"],
        },
      },
    ],
  },
});
