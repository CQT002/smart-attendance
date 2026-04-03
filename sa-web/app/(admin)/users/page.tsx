"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from "@/hooks/use-users";
import { useActiveBranches } from "@/hooks/use-branches";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { DataTableSkeleton } from "@/components/shared/data-table-skeleton";
import { ActiveBadge, RoleBadge } from "@/components/shared/status-badge";
import { Pagination } from "@/components/shared/pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Search, Pencil, Trash2, KeyRound } from "lucide-react";
import { UserFilter, User, CreateUserRequest, UpdateUserRequest } from "@/types/user";
import { UserFormDialog } from "@/components/users/user-form-dialog";
import { ResetPasswordDialog } from "@/components/users/reset-password-dialog";
import { formatDate } from "@/lib/utils";

export default function UsersPage() {
  const [filter, setFilter] = useState<UserFilter>({ page: 1, limit: 10 });
  const [search, setSearch] = useState("");
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [resetUser, setResetUser] = useState<User | null>(null);
  const [showCreate, setShowCreate] = useState(false);

  const { data, isLoading } = useUsers(filter);
  const { data: branches } = useActiveBranches();
  const createUser = useCreateUser();
  const updateUser = useUpdateUser();
  const deleteUser = useDeleteUser();

  const handleSearch = () => {
    setFilter((f) => ({ ...f, search, page: 1 }));
  };

  return (
    <div>
      <Header title="Quản lý Nhân viên" />
      <div className="p-6 space-y-4">
        {/* Toolbar */}
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex flex-1 items-center gap-2 min-w-0">
            <Input
              placeholder="Tìm theo tên, email, mã NV..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              className="max-w-xs"
            />
            <Button variant="outline" size="icon" onClick={handleSearch}>
              <Search className="h-4 w-4" />
            </Button>
          </div>
          <Select
            value={filter.role ?? "all"}
            onValueChange={(v) =>
              setFilter((f) => ({ ...f, role: v === "all" ? undefined : (v as any), page: 1 }))
            }
          >
            <SelectTrigger className="w-36">
              <SelectValue placeholder="Chức vụ" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Tất cả chức vụ</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
              <SelectItem value="manager">Quản lý</SelectItem>
              <SelectItem value="employee">Nhân viên</SelectItem>
            </SelectContent>
          </Select>
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4" />
            Thêm nhân viên
          </Button>
        </div>

        {/* Stats */}
        {data?.meta && (
          <p className="text-sm text-muted-foreground">
            Tổng cộng <strong>{data.meta.total}</strong> nhân viên
          </p>
        )}

        {/* Table */}
        <Card>
          <CardContent className="p-0">
            {isLoading ? (
              <DataTableSkeleton columns={7} />
            ) : (
              <>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Nhân viên</TableHead>
                      <TableHead>Mã NV</TableHead>
                      <TableHead>Chi nhánh</TableHead>
                      <TableHead>Phòng ban / Chức vụ</TableHead>
                      <TableHead>Vai trò</TableHead>
                      <TableHead>Trạng thái</TableHead>
                      <TableHead className="text-right">Thao tác</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data?.data.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell>
                          <div className="font-medium">{user.name}</div>
                          <div className="text-xs text-muted-foreground">{user.email}</div>
                        </TableCell>
                        <TableCell className="font-mono text-sm">{user.employee_code}</TableCell>
                        <TableCell className="text-sm">
                          {user.branch?.name ?? (
                            <span className="text-muted-foreground italic">Chưa gán</span>
                          )}
                        </TableCell>
                        <TableCell className="text-sm">
                          <div>{user.department}</div>
                          <div className="text-xs text-muted-foreground">{user.position}</div>
                        </TableCell>
                        <TableCell>
                          <RoleBadge role={user.role} />
                        </TableCell>
                        <TableCell>
                          <ActiveBadge isActive={user.is_active} />
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              title="Reset mật khẩu"
                              onClick={() => setResetUser(user)}
                            >
                              <KeyRound className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => setEditingUser(user)}
                            >
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="text-destructive hover:text-destructive"
                              onClick={() => {
                                if (confirm(`Vô hiệu hoá tài khoản "${user.name}"?`)) {
                                  deleteUser.mutate(user.id);
                                }
                              }}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                    {data?.data.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                          Không tìm thấy nhân viên nào
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

      <UserFormDialog
        open={showCreate}
        branches={branches ?? []}
        onClose={() => setShowCreate(false)}
        onSubmit={(data) =>
          createUser.mutateAsync(data as CreateUserRequest).then(() => setShowCreate(false))
        }
        loading={createUser.isPending}
      />

      {editingUser && (
        <UserFormDialog
          open={!!editingUser}
          defaultValues={editingUser}
          branches={branches ?? []}
          onClose={() => setEditingUser(null)}
          onSubmit={(data) =>
            updateUser
              .mutateAsync({ id: editingUser.id, data: data as UpdateUserRequest })
              .then(() => setEditingUser(null))
          }
          loading={updateUser.isPending}
        />
      )}

      {resetUser && (
        <ResetPasswordDialog
          open={!!resetUser}
          userId={resetUser.id}
          userName={resetUser.name}
          onClose={() => setResetUser(null)}
        />
      )}
    </div>
  );
}
