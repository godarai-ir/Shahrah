import { Search } from "lucide-react";
import * as React from "react";

import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group";
import { cn } from "@/lib/utils";

type SearchFieldProps = React.ComponentProps<typeof InputGroupInput> & {
  containerClassName?: string;
};

export function SearchField({ containerClassName, "aria-label": ariaLabel, placeholder, ...props }: SearchFieldProps) {
  return (
    <InputGroup className={cn("h-10", containerClassName)}>
      <InputGroupInput type="search" aria-label={ariaLabel ?? placeholder} placeholder={placeholder} {...props} />
      <InputGroupAddon align="inline-end">
        <Search aria-hidden="true" />
      </InputGroupAddon>
    </InputGroup>
  );
}
