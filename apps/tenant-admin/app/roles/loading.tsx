import { getTranslations } from "next-intl/server";

import { PageContainer } from "@/components/app-ui";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export default async function RolesLoading() {
  const t = await getTranslations("accessManagement");

  return (
    <PageContainer
      padding="none"
      aria-busy="true"
      aria-label={t("rolesPage.loading")}
    >
      <div className="flex flex-col gap-3">
        <Skeleton className="h-4 w-28" />
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-5 w-full max-w-xl" />
      </div>

      <Skeleton className="h-10 w-full" />

      <Card>
        <CardHeader className="gap-4">
          <div className="flex flex-col gap-2">
            <Skeleton className="h-6 w-36" />
            <Skeleton className="h-5 w-52" />
          </div>
          <div className="flex flex-col gap-3 sm:flex-row">
            <Skeleton className="h-10 w-full sm:max-w-md" />
            <Skeleton className="h-10 w-full sm:w-48" />
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <Skeleton key={index} className="h-12 w-full" />
          ))}
        </CardContent>
      </Card>
    </PageContainer>
  );
}
