import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin();
const isProductionBuild = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  output: "standalone",
  typescript: {
    tsconfigPath: isProductionBuild ? "tsconfig.build.json" : "tsconfig.json",
  },
  experimental: {
    authInterrupts: true,
    turbopackFileSystemCacheForBuild: true,
  },
};

export default withNextIntl(nextConfig);
