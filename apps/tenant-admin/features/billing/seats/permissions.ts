import type { Permission } from "@/lib/permissions";

/** Any of these lets a member open the seats pages; buying and renewing also needs billing:manage. */
export const SEAT_READ_PERMISSIONS: Permission[] = ["billing:read", "billing:manage", "settings:list", "user:list", "user:manage"];
