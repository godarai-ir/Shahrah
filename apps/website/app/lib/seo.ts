import type { Metadata } from "next";

const base = process.env.NEXT_PUBLIC_SITE_URL || "https://godarai.ir";

export const siteUrl = base.replace(/\/+$/, "");
export const defaultSocialImage = "/brand/social/og-image-light-1200x630.png";

export function absoluteUrl(path = "/"): string {
  if (/^https?:\/\//i.test(path)) {
    return new URL(path).toString();
  }
  return `${siteUrl}${path.startsWith("/") ? path : `/${path}`}`;
}

type SEOInput = {
  title: string;
  description: string;
  path: string;
  keywords?: string[];
  noIndex?: boolean;
  image?: string;
};

export function buildMetadata(input: SEOInput): Metadata {
  const url = absoluteUrl(input.path);
  const image = absoluteUrl(input.image ?? defaultSocialImage);
  return {
    title: input.title,
    description: input.description,
    keywords: input.keywords,
    alternates: {
      canonical: url,
    },
    openGraph: {
      title: input.title,
      description: input.description,
      url,
      siteName: "گدارAI",
      locale: "fa_IR",
      type: "website",
      images: [
        {
          url: image,
          width: 1200,
          height: 630,
          type: "image/png",
          alt: "لوگوی گدارAI",
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: input.title,
      description: input.description,
      images: [image],
    },
    robots: input.noIndex
      ? {
          index: false,
          follow: false,
        }
      : {
          index: true,
          follow: true,
        },
  };
}
