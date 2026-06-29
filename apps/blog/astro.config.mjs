// @ts-check
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_SITE_URL || "https://godarai.ir/blog";
const allowedHosts = (process.env.PUBLIC_ALLOWED_HOSTS || "godarai.ir,localhost,127.0.0.1")
  .split(",")
  .map((host) => host.trim())
  .filter(Boolean);

export default defineConfig({
  site,
  trailingSlash: "never",
  vite: {
    server: {
      allowedHosts,
    },
    preview: {
      allowedHosts,
    },
  },
});
