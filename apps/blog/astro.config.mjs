// @ts-check
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_SITE_URL || "https://godarai.ir/blog";
const base = process.env.PUBLIC_BASE_PATH || "/blog";
const allowedHosts = (process.env.PUBLIC_ALLOWED_HOSTS || "godarai.ir,localhost,127.0.0.1")
  .split(",")
  .map((host) => host.trim())
  .filter(Boolean);

export default defineConfig({
  site,
  base,
  build: {
    assetsPrefix: base,
  },
  trailingSlash: "never",
  vite: {
    plugins: [tailwindcss()],
    server: {
      allowedHosts,
    },
    preview: {
      allowedHosts,
    },
  },
});
