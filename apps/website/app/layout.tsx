import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Shahrah",
  description: "Shahrah AI Gateway website",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="fa" dir="rtl" suppressHydrationWarning className="h-full antialiased">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
