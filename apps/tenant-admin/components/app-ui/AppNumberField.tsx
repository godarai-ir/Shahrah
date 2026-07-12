import { AppTextField, type AppTextFieldProps } from "@/components/app-ui/AppTextField";
import { cn } from "@/lib/utils";

export type AppNumberFieldProps = Omit<AppTextFieldProps, "type">;

export function AppNumberField({ inputMode = "decimal", dir = "ltr", className, ...props }: AppNumberFieldProps) {
  return <AppTextField type="number" inputMode={inputMode} dir={dir} className={cn("text-left tabular-nums", className)} {...props} />;
}
