"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { PencilIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { CreateOrgDialog } from "@/components/orgs/create-org-dialog";
import { RenameOrgDialog } from "@/components/orgs/rename-org-dialog";
import { DeleteDialog } from "@/components/shared/delete-dialog";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { roleAtLeast } from "@/lib/roles";
import { deleteOrganization } from "@/services/organizations";
import type { OrganizationListItem } from "@/types/organization";

export function OrgList({ orgs }: { orgs: OrganizationListItem[] }) {
  const router = useRouter();

  return (
    <div className="grid gap-4">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-sm">Organizações</h1>
        <CreateOrgDialog />
      </div>
      {orgs.length === 0 ? (
        <p className="text-muted-foreground">
          Você ainda não tem organizações. Crie a primeira para começar.
        </p>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {orgs.map((org) => (
            <Card key={org.id} data-testid="org-card">
              <CardHeader>
                <div className="flex items-start justify-between gap-2">
                  <Link href={`/orgs/${org.id}`} className="min-w-0">
                    <CardTitle className="truncate hover:text-primary">
                      {org.name}
                    </CardTitle>
                    <CardDescription>{org.slug}</CardDescription>
                  </Link>
                  {roleAtLeast(org.role, "ADMIN") && (
                    <div className="flex shrink-0 gap-1">
                      <RenameOrgDialog
                        trigger={
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Renomear ${org.name}`}
                          >
                            <PencilIcon />
                          </Button>
                        }
                        orgId={org.id}
                        currentName={org.name}
                      />
                      <DeleteDialog
                        trigger={
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Deletar ${org.name}`}
                          >
                            <Trash2Icon />
                          </Button>
                        }
                        entityLabel="a organização"
                        name={org.name}
                        onConfirm={async () => {
                          await deleteOrganization(org.id);
                          toast.success("Organização deletada");
                          router.refresh();
                        }}
                      />
                    </div>
                  )}
                </div>
              </CardHeader>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
