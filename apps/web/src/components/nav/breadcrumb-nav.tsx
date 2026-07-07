import { Fragment } from "react";
import Link from "next/link";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

interface BreadcrumbTrailItem {
  label: string;
  href?: string;
}

export function BreadcrumbNav({ items }: { items: BreadcrumbTrailItem[] }) {
  return (
    <Breadcrumb data-testid="breadcrumb">
      <BreadcrumbList>
        {items.map((item, index) => (
          <Fragment key={`${item.label}-${index}`}>
            {index > 0 && <BreadcrumbSeparator />}
            <BreadcrumbItem>
              {item.href ? (
                <BreadcrumbLink asChild className="max-w-48 truncate">
                  <Link href={item.href}>{item.label}</Link>
                </BreadcrumbLink>
              ) : (
                <BreadcrumbPage className="max-w-48 truncate">
                  {item.label}
                </BreadcrumbPage>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
