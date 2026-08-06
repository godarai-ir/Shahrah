export const socialLinks = [
  {
    label: "لینکدین",
    ariaLabel: "گدارAI در لینکدین",
    href: "https://www.linkedin.com/company/godarai/",
  },
  {
    label: "تلگرام",
    ariaLabel: "گدارAI در تلگرام",
    href: "https://t.me/godarai",
  },
  {
    label: "گیت‌هاب",
    ariaLabel: "گدارAI در گیت‌هاب",
    href: "https://github.com/godarai-ir",
  },
] as const;

export const socialLinkUrls = socialLinks.map(({ href }) => href);
