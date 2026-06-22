import type { Metadata } from "next";
import "./globals.css";
import { absoluteUrl } from "@/app/lib/seo";

export const metadata: Metadata = {
  metadataBase: new URL(absoluteUrl("/")),
  title: {
    default: "GodarAI | AI Gateway for Iranian Teams",
    template: "%s | GodarAI",
  },
  description:
    "Shahrah is an AI gateway for Iranian teams to access, manage, and monitor leading language models with stable local infrastructure.",
  applicationName: "GodarAI",
  manifest: "/site.webmanifest",
  icons: {
    icon: [
      { url: "/favicon.ico", sizes: "any" },
      { url: "/favicon-32x32.png", type: "image/png", sizes: "32x32" },
      { url: "/favicon-16x16.png", type: "image/png", sizes: "16x16" },
    ],
    apple: [{ url: "/apple-touch-icon.png", sizes: "180x180" }],
  },
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "GodarAI | AI Gateway for Iranian Teams",
    description:
      "Route, monitor, and secure AI model usage with Shahrah's LLM gateway built for Iranian developers and businesses.",
    url: "/",
    siteName: "Shahrah",
    locale: "fa_IR",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "GodarAI | AI Gateway for Iranian Teams",
    description:
      "AI gateway for local payments, observability, and scalable model access in Iran.",
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="fa" dir="rtl" suppressHydrationWarning className="h-full antialiased" style={{ fontFamily: 'Vazirmatn, sans-serif' }}>
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
