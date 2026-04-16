"use client";

import { useMemo, useState } from "react";
import { Header } from "@/components/layout/header";
import {
  useHolidays,
  useCreateHoliday,
  useUpdateHoliday,
  useDeleteHoliday,
} from "@/hooks/use-holidays";
import { useCurrentUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DataTableSkeleton } from "@/components/shared/data-table-skeleton";
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
import { Plus, Pencil, Trash2, Badge as BadgeIcon } from "lucide-react";
import {
  CreateHolidayRequest,
  Holiday,
  HolidayFilter,
  HolidayType,
  UpdateHolidayRequest,
} from "@/types/holiday";
import { HolidayFormDialog } from "@/components/holidays/holiday-form-dialog";
import { formatDate } from "@/lib/utils";

export default function HolidaysPage() {
  const { data: currentUser } = useCurrentUser();
  const isAdmin = currentUser?.role === "admin";

  const currentYear = new Date().getFullYear();
  const years = useMemo(
    () => [currentYear - 1, currentYear, currentYear + 1],
    [currentYear]
  );

  const [filter, setFilter] = useState<HolidayFilter>({
    year: currentYear,
    page: 1,
    limit: 100,
  });
  const [editing, setEditing] = useState<Holiday | null>(null);
  const [showCreate, setShowCreate] = useState(false);

  const { data, isLoading } = useHolidays(filter);
  const createHoliday = useCreateHoliday();
  const updateHoliday = useUpdateHoliday();
  const deleteHoliday = useDeleteHoliday();

  return (
    <div>
      <Header title="Quản lý ngày lễ" />
      <div className="p-6 space-y-4">
        {/* Toolbar */}
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">Năm</span>
            <Select
              value={String(filter.year)}
              onValueChange={(v) =>
                setFilter((f) => ({ ...f, year: Number(v), page: 1 }))
              }
            >
              <SelectTrigger className="w-[120px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {years.map((y) => (
                  <SelectItem key={y} value={String(y)}>
                    {y}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">Loại</span>
            <Select
              value={filter.type ?? "all"}
              onValueChange={(v) =>
                setFilter((f) => ({
                  ...f,
                  type: v === "all" ? undefined : (v as HolidayType),
                  page: 1,
                }))
              }
            >
              <SelectTrigger className="w-[160px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Tất cả</SelectItem>
                <SelectItem value="national">Lễ quốc gia</SelectItem>
                <SelectItem value="company">Lễ công ty</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex-1" />

          {isAdmin && (
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="h-4 w-4" />
              Thêm ngày lễ
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
                      <TableHead>Ngày</TableHead>
                      <TableHead>Tên ngày lễ</TableHead>
                      <TableHead>Loại</TableHead>
                      <TableHead>Hệ số lương</TableHead>
                      <TableHead>Nghỉ bù</TableHead>
                      <TableHead>Ngày tạo</TableHead>
                      {isAdmin && <TableHead className="text-right">Thao tác</TableHead>}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data?.data.map((h) => (
                      <TableRow key={h.id}>
                        <TableCell className="font-medium">
                          {formatDate(h.date)}
                        </TableCell>
                        <TableCell>
                          <div className="font-medium">{h.name}</div>
                          {h.description && (
                            <div className="text-xs text-muted-foreground">
                              {h.description}
                            </div>
                          )}
                        </TableCell>
                        <TableCell>
                          <span
                            className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
                              h.type === "national"
                                ? "bg-pink-100 text-pink-700"
                                : "bg-purple-100 text-purple-700"
                            }`}
                          >
                            <BadgeIcon className="h-3 w-3" />
                            {h.type === "national" ? "Quốc gia" : "Công ty"}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className="font-mono text-sm font-semibold text-green-700">
                            x{h.coefficient.toFixed(1)}
                          </span>
                          <span className="ml-1 text-xs text-muted-foreground">
                            ({Math.round(h.coefficient * 100)}%)
                          </span>
                        </TableCell>
                        <TableCell>
                          {h.is_compensated ? (
                            <div className="text-xs">
                              <span className="text-orange-600 font-medium">Có</span>
                              {h.compensate_for && (
                                <div className="text-muted-foreground">
                                  Bù cho {formatDate(h.compensate_for)}
                                </div>
                              )}
                            </div>
                          ) : (
                            <span className="text-xs text-muted-foreground">—</span>
                          )}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(h.created_at)}
                        </TableCell>
                        {isAdmin && (
                          <TableCell className="text-right">
                            <div className="flex items-center justify-end gap-1">
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => setEditing(h)}
                                title="Chỉnh sửa"
                              >
                                <Pencil className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="text-destructive hover:text-destructive"
                                onClick={() => {
                                  if (confirm(`Xoá ngày lễ "${h.name}"?`)) {
                                    deleteHoliday.mutate(h.id);
                                  }
                                }}
                                title="Xoá"
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </div>
                          </TableCell>
                        )}
                      </TableRow>
                    ))}
                    {data?.data.length === 0 && (
                      <TableRow>
                        <TableCell
                          colSpan={isAdmin ? 7 : 6}
                          className="text-center py-8 text-muted-foreground"
                        >
                          Không có ngày lễ nào trong năm {filter.year}
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
      <HolidayFormDialog
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onSubmit={(data) =>
          createHoliday
            .mutateAsync(data as CreateHolidayRequest)
            .then(() => setShowCreate(false))
        }
        loading={createHoliday.isPending}
      />

      {/* Edit dialog */}
      {editing && (
        <HolidayFormDialog
          open={!!editing}
          defaultValues={editing}
          onClose={() => setEditing(null)}
          onSubmit={(data) =>
            updateHoliday
              .mutateAsync({ id: editing.id, data: data as UpdateHolidayRequest })
              .then(() => setEditing(null))
          }
          loading={updateHoliday.isPending}
        />
      )}
    </div>
  );
}
