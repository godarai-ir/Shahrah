// @ts-check
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_SITE_URL || "https://shahrah.ir/blog";

export default defineConfig({
  site,
  trailingSlash: "never",
});
