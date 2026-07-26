import Link from "next/link";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export function CatalogBreadcrumb({
  providersLabel,
  providerSlug,
  providerLabel,
  modelLabel,
}: {
  providersLabel: string;
  providerSlug: string;
  providerLabel: string;
  modelLabel?: string;
}) {
  return (
    <Breadcrumb dir="rtl">
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink asChild>
            <Link href="/providers">{providersLabel}</Link>
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          {modelLabel ? (
            <BreadcrumbLink asChild>
              <Link href={`/providers/${providerSlug}`}>
                <bdi dir="auto">{providerLabel}</bdi>
              </Link>
            </BreadcrumbLink>
          ) : (
            <BreadcrumbPage>
              <bdi dir="auto">{providerLabel}</bdi>
            </BreadcrumbPage>
          )}
        </BreadcrumbItem>
        {modelLabel ? (
          <>
            <BreadcrumbSeparator />
            <BreadcrumbItem className="min-w-0">
              <BreadcrumbPage className="max-w-64 truncate">
                <bdi dir="auto">{modelLabel}</bdi>
              </BreadcrumbPage>
            </BreadcrumbItem>
          </>
        ) : null}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
