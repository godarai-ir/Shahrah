import { redirect } from "next/navigation";

// IRR reports and documents ended with the credit cutover; credit usage lives on the billing overview.
export default function RetiredIRRReportPage() {
  redirect("/billing");
}
