import { toNextJsHandler } from "better-auth/next-js";

import { socialAuth } from "@/app/lib/social-auth/config";

export const dynamic = "force-dynamic";

export const { GET, POST } = toNextJsHandler(socialAuth);

