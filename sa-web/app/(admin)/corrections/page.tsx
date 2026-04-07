"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { useCorrections, useProcessCorrection } from "@/hooks/use-corrections";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DataTableSkeleton } from "@/components/shared/data-table-skeleton";
import { StatusBadge } from "@/components/shared/status-badge";
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Check, X, Eye, Clock, User as UserIcon } from "lucide-react";
import { CorrectionFilter, CorrectionStatus, AttendanceCorrection } from "@/types";
import { formatDate, formatDateTime, formatTime } from "@/lib/utils";

const STATUS_OPTIONS = [
  { value: "all", label: "Tất cả trạng thái" },
  { value: "pending", label: "Chờ duyệt" },
  { value: "approved", label: "Đã duyệt" },
  { value: "rejected", label: "Từ chối" },
];

const CORRECTION_STATUS_CONFIG: Record<
  string,
  { label: string; variant: "warning" | "success" | "destructive" | "secondary" }
> = {
  pending: { label: "Chờ duyệt", variant: "warning" },
  approved: { label: "Đã duyệt", variant: "success" },
  rejected: { label: "Từ chối", variant: "destructive" },
};

const ORIGINAL_STATUS_LABEL: Record<string, string> = {
  late: "Đi trễ",
  early_leave: "Về sớm",
  late_early_leave: "Đi trễ - Về sớm",
};

export default function CorrectionsPage() {
  const [filter, setFilter] = useState<CorrectionFilter>({
    page: 1,
    limit: 10,
    status: "pending",
  });
  const [detailItem, setDetailItem] = useState<AttendanceCorrection | null>(null);
  const [processItem, setProcessItem] = useState<{
    correction: AttendanceCorrection;
    action: "approved" | "rejected";
  } | null>(null);
  const [managerNote, setManagerNote] = useState("");

  const { data, isLoading } = useCorrections(filter);
  const processMutation = useProcessCorrection();

  const handleProcess = () => {
    if (!processItem) return;
    processMutation.mutate(
      {
        id: processItem.correction.id,
        req: { status: processItem.action, manager_note: managerNote },
      },
      {
        onSuccess: () => {
          setProcessItem(null);
          setManagerNote("");
        },
      }
    );
  };

  return (
    <div>
      <Header title="Duyệt bù công" />
      <div className="p-6 space-y-4">
        {/* Filters */}
        <div className="flex items-center gap-3">
          <Select
            value={filter.status ?? "all"}
            onValueChange={(v) =>
              setFilter((f) => ({
                ...f,
                status: v === "all" ? undefined : (v as CorrectionStatus),
                page: 1,
              }))
            }
          >
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Trạng thái" />
            </SelectTrigger>
            <SelectContent>
              {STATUS_OPTIONS.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {data?.meta && (
            <p className="text-sm text-muted-foreground ml-auto">
              Tổng <strong>{data.meta.total}</strong> yêu cầu
            </p>
          )}
        </div>

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
                      <TableHead>Ngày chấm công</TableHead>
                      <TableHead>Trạng thái gốc</TableHead>
                      <TableHead>Lý do</TableHead>
                      <TableHead>Ngày gửi</TableHead>
                      <TableHead>Trạng thái</TableHead>
                      <TableHead className="text-right">Thao tác</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data?.data.map((item) => {
                      const statusConfig =
                        CORRECTION_STATUS_CONFIG[item.status] ?? {
                          label: item.status,
                          variant: "secondary" as const,
                        };
                      return (
                        <TableRow key={item.id}>
                          <TableCell>
                            <div className="font-medium">
                              {item.user?.name ?? `#${item.user_id}`}
                            </div>
                            <div className="text-xs text-muted-foreground">
                              {item.user?.employee_code}
                            </div>
                          </TableCell>
                          <TableCell className="text-sm">
                            {item.attendance_log
                              ? formatDate(item.attendance_log.date)
                              : "—"}
                          </TableCell>
                          <TableCell>
                            <StatusBadge status={item.original_status} />
                          </TableCell>
                          <TableCell className="text-sm max-w-[200px] truncate">
                            {item.description}
                          </TableCell>
                          <TableCell className="text-sm">
                            {formatDateTime(item.created_at)}
                          </TableCell>
                          <TableCell>
                            <Badge variant={statusConfig.variant}>
                              {statusConfig.label}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="flex justify-end gap-1">
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setDetailItem(item)}
                              >
                                <Eye className="h-4 w-4" />
                              </Button>
                              {item.status === "pending" && (
                                <>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="text-green-600 hover:text-green-700"
                                    onClick={() =>
                                      setProcessItem({
                                        correction: item,
                                        action: "approved",
                                      })
                                    }
                                  >
                                    <Check className="h-4 w-4" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="text-red-600 hover:text-red-700"
                                    onClick={() =>
                                      setProcessItem({
                                        correction: item,
                                        action: "rejected",
                                      })
                                    }
                                  >
                                    <X className="h-4 w-4" />
                                  </Button>
                                </>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                    {data?.data.length === 0 && (
                      <TableRow>
                        <TableCell
                          colSpan={7}
                          className="text-center py-8 text-muted-foreground"
                        >
                          Không có yêu cầu nào
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

      {/* Detail Dialog */}
      <Dialog open={!!detailItem} onOpenChange={() => setDetailItem(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Chi tiết yêu cầu bù công</DialogTitle>
          </DialogHeader>
          {detailItem && (
            <div className="space-y-4">
              {/* Employee info */}
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100">
                  <UserIcon className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <div className="font-medium">{detailItem.user?.name}</div>
                  <div className="text-sm text-muted-foreground">
                    {detailItem.user?.employee_code} &middot;{" "}
                    {detailItem.user?.department}
                  </div>
                </div>
              </div>

              {/* Attendance info */}
              <div className="rounded-md border p-3 space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Ngày chấm công</span>
                  <span className="font-medium">
                    {detailItem.attendance_log
                      ? formatDate(detailItem.attendance_log.date)
                      : "—"}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Trạng thái gốc</span>
                  <span className="font-medium">
                    {ORIGINAL_STATUS_LABEL[detailItem.original_status] ??
                      detailItem.original_status}
                  </span>
                </div>
                {detailItem.attendance_log && (
                  <>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Check-in</span>
                      <span className="font-medium">
                        {formatTime(detailItem.attendance_log.check_in_time)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Check-out</span>
                      <span className="font-medium">
                        {formatTime(detailItem.attendance_log.check_out_time)}
                      </span>
                    </div>
                  </>
                )}
              </div>

              {/* Description */}
              <div>
                <div className="text-sm font-medium mb-1">Lý do</div>
                <div className="rounded-md bg-muted p-3 text-sm">
                  {detailItem.description}
                </div>
              </div>

              {/* Audit log */}
              {detailItem.processed_at && (
                <div className="rounded-md border p-3 space-y-2 text-sm">
                  <div className="flex items-center gap-2 text-muted-foreground mb-1">
                    <Clock className="h-4 w-4" />
                    <span className="font-medium">Audit Log</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Người duyệt</span>
                    <span className="font-medium">
                      {detailItem.processed_by?.name ?? "Hệ thống"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Thời gian</span>
                    <span className="font-medium">
                      {formatDateTime(detailItem.processed_at)}
                    </span>
                  </div>
                  {detailItem.manager_note && (
                    <div>
                      <span className="text-muted-foreground">Ghi chú: </span>
                      <span className="italic">{detailItem.manager_note}</span>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Process Dialog */}
      <Dialog
        open={!!processItem}
        onOpenChange={() => {
          setProcessItem(null);
          setManagerNote("");
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {processItem?.action === "approved"
                ? "Duyệt yêu cầu bù công"
                : "Từ chối yêu cầu bù công"}
            </DialogTitle>
          </DialogHeader>
          {processItem && (
            <div className="space-y-4">
              <div className="text-sm">
                <span className="text-muted-foreground">Nhân viên: </span>
                <span className="font-medium">
                  {processItem.correction.user?.name}
                </span>
              </div>
              <div className="text-sm">
                <span className="text-muted-foreground">Lý do: </span>
                <span>{processItem.correction.description}</span>
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">
                  Ghi chú (không bắt buộc)
                </label>
                <Input
                  placeholder={
                    processItem.action === "approved"
                      ? "Ví dụ: Đã xác nhận với phòng HC..."
                      : "Ví dụ: Lý do không hợp lệ..."
                  }
                  value={managerNote}
                  onChange={(e) => setManagerNote(e.target.value)}
                />
              </div>
            </div>
          )}
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setProcessItem(null);
                setManagerNote("");
              }}
            >
              Huỷ
            </Button>
            <Button
              variant={
                processItem?.action === "approved" ? "default" : "destructive"
              }
              onClick={handleProcess}
              disabled={processMutation.isPending}
            >
              {processMutation.isPending
                ? "Đang xử lý..."
                : processItem?.action === "approved"
                  ? "Duyệt"
                  : "Từ chối"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
