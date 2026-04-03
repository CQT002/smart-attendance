"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { useBranches, useCreateBranch, useUpdateBranch, useDeleteBranch } from "@/hooks/use-branches";
import { useCurrentUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { DataTableSkeleton } from "@/components/shared/data-table-skeleton";
import { ActiveBadge } from "@/components/shared/status-badge";
import { Pagination } from "@/components/shared/pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Plus, Search, Pencil, Trash2, MapPin, Wifi, Eye } from "lucide-react";
import { BranchFilter, Branch, CreateBranchRequest, UpdateBranchRequest } from "@/types/branch";
import { BranchFormDialog } from "@/components/branches/branch-form-dialog";
import { WifiConfigDialog } from "@/components/branches/wifi-config-dialog";
import { BranchDetailDialog } from "@/components/branches/branch-detail-dialog";
import { formatDate } from "@/lib/utils";

export default function BranchesPage() {
  const { data: currentUser } = useCurrentUser();
  const isAdmin = currentUser?.role === "admin";

  const [filter, setFilter] = useState<BranchFilter>({ page: 1, limit: 10 });
  const [search, setSearch] = useState("");
  const [editingBranch, setEditingBranch] = useState<Branch | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [wifiBranch, setWifiBranch] = useState<Branch | null>(null);
  const [detailBranch, setDetailBranch] = useState<Branch | null>(null);

  // Manager: backend tự filter theo chi nhánh qua JWT
  const { data, isLoading } = useBranches(filter);
  const createBranch = useCreateBranch();
  const updateBranch = useUpdateBranch();
  const deleteBranch = useDeleteBranch();

  const handleSearch = () => {
    setFilter((f) => ({ ...f, search, page: 1 }));
  };

  return (
    <div>
      <Header title="Quản lý Chi nhánh" />
      <div className="p-6 space-y-4">
        {/* Toolbar */}
        <div className="flex items-center gap-3">
          <div className="flex flex-1 items-center gap-2">
            <Input
              placeholder="Tìm kiếm chi nhánh..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              className="max-w-sm"
            />
            <Button variant="outline" size="icon" onClick={handleSearch}>
              <Search className="h-4 w-4" />
            </Button>
          </div>
          {isAdmin && (
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="h-4 w-4" />
              Thêm chi nhánh
            </Button>
          )}
        </div>

        {/* Table */}
        <Card>
          <CardContent className="p-0">
            {isLoading ? (
              <DataTableSkeleton columns={6} />
            ) : (
              <>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Mã / Tên</TableHead>
                      <TableHead>Địa chỉ</TableHead>
                      <TableHead>Liên hệ</TableHead>
                      <TableHead>GPS</TableHead>
                      <TableHead>WiFi</TableHead>
                      <TableHead>Trạng thái</TableHead>
                      <TableHead>Ngày tạo</TableHead>
                      <TableHead className="text-right">Thao tác</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data?.data.map((branch) => (
                      <TableRow key={branch.id}>
                        <TableCell>
                          <div className="font-medium">{branch.name}</div>
                          <div className="text-xs text-muted-foreground">{branch.code}</div>
                        </TableCell>
                        <TableCell className="max-w-[200px] truncate text-sm">
                          {branch.address}
                        </TableCell>
                        <TableCell className="text-sm">
                          <div>{branch.phone}</div>
                          <div className="text-xs text-muted-foreground">{branch.email}</div>
                        </TableCell>
                        <TableCell>
                          {branch.latitude && branch.longitude ? (
                            <div className="flex items-center gap-1 text-xs text-green-600">
                              <MapPin className="h-3 w-3" />
                              {branch.latitude.toFixed(4)}, {branch.longitude.toFixed(4)}
                            </div>
                          ) : (
                            <span className="text-xs text-muted-foreground">Chưa cấu hình</span>
                          )}
                        </TableCell>
                        <TableCell>
                          {branch.wifi_count > 0 ? (
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => setWifiBranch(branch)}
                              title="Xem danh sách WiFi"
                            >
                              <Wifi className="h-4 w-4 text-green-600" />
                            </Button>
                          ) : (
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              disabled
                              title="Chưa cấu hình WiFi"
                            >
                              <Wifi className="h-4 w-4 text-muted-foreground" />
                            </Button>
                          )}
                        </TableCell>
                        <TableCell>
                          <ActiveBadge isActive={branch.is_active} />
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(branch.created_at)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => setDetailBranch(branch)}
                              title="Xem chi tiết"
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                            {isAdmin && (
                              <>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => setEditingBranch(branch)}
                                  title="Chỉnh sửa"
                                >
                                  <Pencil className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="text-destructive hover:text-destructive"
                                  onClick={() => {
                                    if (confirm(`Vô hiệu hoá chi nhánh "${branch.name}"?`)) {
                                      deleteBranch.mutate(branch.id);
                                    }
                                  }}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                    {data?.data.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                          Không tìm thấy chi nhánh nào
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
                {data?.meta && data.meta.total_pages > 1 && (
                  <Pagination
                    meta={data.meta}
                    onPageChange={(p) => setFilter((f) => ({ ...f, page: p }))}
                  />
                )}
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Create dialog */}
      <BranchFormDialog
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onSubmit={(data) =>
          createBranch.mutateAsync(data as CreateBranchRequest).then(() => setShowCreate(false))
        }
        loading={createBranch.isPending}
      />

      {/* Edit dialog */}
      {editingBranch && (
        <BranchFormDialog
          open={!!editingBranch}
          defaultValues={editingBranch}
          onClose={() => setEditingBranch(null)}
          onSubmit={(data) =>
            updateBranch
              .mutateAsync({ id: editingBranch.id, data: data as UpdateBranchRequest })
              .then(() => setEditingBranch(null))
          }
          loading={updateBranch.isPending}
        />
      )}

      {/* WiFi config dialog */}
      {wifiBranch && (
        <WifiConfigDialog
          branch={wifiBranch}
          open={!!wifiBranch}
          onClose={() => setWifiBranch(null)}
        />
      )}

      {/* Detail dialog */}
      {detailBranch && (
        <BranchDetailDialog
          branch={detailBranch}
          open={!!detailBranch}
          onClose={() => setDetailBranch(null)}
        />
      )}
    </div>
  );
}
