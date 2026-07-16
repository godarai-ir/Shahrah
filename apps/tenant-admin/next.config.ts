import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin();
const isProductionBuild = process.env.NODE_ENV === "production";

export function buildContentSecurityPolicy(isProduction: boolean) {
  return [
    "default-src 'self'",
    "base-uri 'self'",
    `connect-src 'self'${isProduction ? "" : " http: https: ws: wss:"}`,
    "font-src 'self' data:",
    "form-action 'self'",
    "frame-ancestors 'none'",
    "img-src 'self' data: blob: https:",
    "object-src 'none'",
    `script-src 'self' 'unsafe-inline'${isProduction ? "" : " 'unsafe-eval'"}`,
    "style-src 'self' 'unsafe-inline'",
    "worker-src 'self' blob:",
  ].join("; ");
}

const contentSecurityPolicy = buildContentSecurityPolicy(isProductionBuild);

const browserSecurityHeaders = [
  { key: "Content-Security-Policy", value: contentSecurityPolicy },
  {
    key: "Strict-Transport-Security",
    value: "max-age=63072000; includeSubDomains; preload",
  },
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "X-Frame-Options", value: "DENY" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  {
    key: "Permissions-Policy",
    value: "camera=(), geolocation=(), microphone=(), payment=(), usb=()",
  },
];

const nextConfig: NextConfig = {
  output: "standalone",
  async headers() {
    return [
      {
        source: "/:path*",
        headers: browserSecurityHeaders,
      },
      {
        source: "/api/:path*",
        headers: [{ key: "Cache-Control", value: "no-store" }],
      },
    ];
  },
  typescript: {
    tsconfigPath: isProductionBuild ? "tsconfig.build.json" : "tsconfig.json",
  },
  experimental: {
    authInterrupts: true,
    turbopackFileSystemCacheForBuild: true,
  },
};

export default withNextIntl(nextConfig);
