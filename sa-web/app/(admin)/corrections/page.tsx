"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { useCorrections, useProcessCorrection, useBatchApproveCorrections } from "@/hooks/use-corrections";
import { useLeaves, useProcessLeave, useBatchApproveLeaves } from "@/hooks/use-leaves";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DataTableSkeleton } from "@/components/shared/data-table-skeleton";
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
import {
  CorrectionFilter,
  CorrectionStatus,
  AttendanceCorrection,
} from "@/types";
import { LeaveRequest, LeaveFilter, LeaveStatus } from "@/types/leave";
import { formatDate, formatDateTime, formatTime } from "@/lib/utils";

type UnifiedStatus = "pending" | "approved" | "rejected";

const STATUS_OPTIONS = [
  { value: "all", label: "Tất cả trạng thái" },
  { value: "pending", label: "Chờ duyệt" },
  { value: "approved", label: "Đã duyệt" },
  { value: "rejected", label: "Từ chối" },
];

const TYPE_OPTIONS = [
  { value: "all", label: "Tất cả loại" },
  { value: "correction", label: "Bù công" },
  { value: "leave", label: "Nghỉ phép" },
];

const APPROVAL_STATUS_CONFIG: Record<
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
  absent: "Vắng mặt",
  half_day: "Nửa ngày",
};

const LEAVE_TYPE_LABEL: Record<string, string> = {
  full_day: "Cả ngày",
  half_day_morning: "Buổi sáng",
  half_day_afternoon: "Buổi chiều",
};

// Unified row type for the table
interface UnifiedItem {
  id: number;
  type: "correction" | "leave";
  userName: string;
  employeeCode: string;
  date: string;
  detail: string;
  description: string;
  status: string;
  createdAt: string;
  raw: AttendanceCorrection | LeaveRequest;
}

export default function CorrectionsPage() {
  const [statusFilter, setStatusFilter] = useState<UnifiedStatus | undefined>(
    "pending"
  );
  const [typeFilter, setTypeFilter] = useState<"all" | "correction" | "leave">(
    "all"
  );
  const [page, setPage] = useState(1);
  const limit = 10;

  // Detail dialogs
  const [detailCorrection, setDetailCorrection] =
    useState<AttendanceCorrection | null>(null);
  const [detailLeave, setDetailLeave] = useState<LeaveRequest | null>(null);

  // Process dialog
  const [processItem, setProcessItem] = useState<{
    id: number;
    type: "correction" | "leave";
    action: "approved" | "rejected";
    userName: string;
    description: string;
  } | null>(null);
  const [managerNote, setManagerNote] = useState("");

  // Fetch both data sources
  const correctionFilter: CorrectionFilter = {
    status: statusFilter as CorrectionStatus | undefined,
    page,
    limit,
  };
  const leaveFilter: LeaveFilter = {
    status: statusFilter as LeaveStatus | undefined,
    page,
    limit,
  };

  const {
    data: corrData,
    isLoading: corrLoading,
  } = useCorrections(
    typeFilter === "leave" ? { ...correctionFilter, page: 1, limit: 0 } : correctionFilter
  );
  const {
    data: leaveData,
    isLoading: leaveLoading,
  } = useLeaves(
    typeFilter === "correction" ? { ...leaveFilter, page: 1, limit: 0 } : leaveFilter
  );

  const processCorrectionMutation = useProcessCorrection();
  const processLeaveMutation = useProcessLeave();
  const batchApproveCorrMutation = useBatchApproveCorrections();
  const batchApproveLeavesMutation = useBatchApproveLeaves();
  const [showBatchConfirm, setShowBatchConfirm] = useState(false);

  const isLoading = corrLoading || leaveLoading;

  // Build unified list
  const unifiedItems: UnifiedItem[] = [];

  if (typeFilter !== "leave" && corrData?.data) {
    for (const c of corrData.data) {
      unifiedItems.push({
        id: c.id,
        type: "correction",
        userName: c.user?.name ?? `#${c.user_id}`,
        employeeCode: c.user?.employee_code ?? "",
        date: c.attendance_log ? formatDate(c.attendance_log.date) : "—",
        detail:
          ORIGINAL_STATUS_LABEL[c.original_status] ?? c.original_status,
        description: c.description,
        status: c.status,
        createdAt: c.created_at,
        raw: c,
      });
    }
  }

  if (typeFilter !== "correction" && leaveData?.data) {
    for (const l of leaveData.data) {
      unifiedItems.push({
        id: l.id,
        type: "leave",
        userName: l.user?.name ?? `#${l.user_id}`,
        employeeCode: l.user?.employee_code ?? "",
        date: formatDate(l.leave_date),
        detail: LEAVE_TYPE_LABEL[l.leave_type] ?? l.leave_type,
        description: l.description,
        status: l.status,
        createdAt: l.created_at,
        raw: l,
      });
    }
  }

  // Sort by created_at DESC
  unifiedItems.sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );

  const totalItems =
    (typeFilter !== "leave" ? corrData?.meta?.total ?? 0 : 0) +
    (typeFilter !== "correction" ? leaveData?.meta?.total ?? 0 : 0);

  const totalPages = Math.ceil(totalItems / limit) || 1;

  const handleProcess = () => {
    if (!processItem) return;
    const req = {
      status: processItem.action as "approved" | "rejected",
      manager_note: managerNote,
    };

    const onSuccess = () => {
      setProcessItem(null);
      setManagerNote("");
    };

    if (processItem.type === "correction") {
      processCorrectionMutation.mutate(
        { id: processItem.id, req },
        { onSuccess }
      );
    } else {
      processLeaveMutation.mutate(
        { id: processItem.id, req },
        { onSuccess }
      );
    }
  };

  const isPending =
    processCorrectionMutation.isPending || processLeaveMutation.isPending;

  return (
    <div>
      <Header title="Duyệt chấm công" />
      <div className="p-6 space-y-4">
        {/* Filters */}
        <div className="flex items-center gap-3">
          <Select
            value={statusFilter ?? "all"}
            onValueChange={(v) => {
              setStatusFilter(
                v === "all" ? undefined : (v as UnifiedStatus)
              );
              setPage(1);
            }}
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

          <Select
            value={typeFilter}
            onValueChange={(v) => {
              setTypeFilter(v as "all" | "correction" | "leave");
              setPage(1);
            }}
          >
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Loại yêu cầu" />
            </SelectTrigger>
            <SelectContent>
              {TYPE_OPTIONS.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <p className="text-sm text-muted-foreground ml-auto">
            Tổng <strong>{totalItems}</strong> yêu cầu
          </p>

          {statusFilter === "pending" && totalItems > 0 && (
            <Button
              variant="outline"
              size="sm"
              className="text-green-600 border-green-600 hover:bg-green-50"
              onClick={() => setShowBatchConfirm(true)}
            >
              <Check className="h-4 w-4 mr-1" />
              Duyệt tất cả
            </Button>
          )}
        </div>

        {/* Table */}
        <Card>
          <CardContent className="p-0">
            {isLoading ? (
              <DataTableSkeleton columns={8} />
            ) : (
              <>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Loại</TableHead>
                      <TableHead>Nhân viên</TableHead>
                      <TableHead>Ngày</TableHead>
                      <TableHead>Chi tiết</TableHead>
                      <TableHead>Lý do</TableHead>
                      <TableHead>Ngày gửi</TableHead>
                      <TableHead>Trạng thái</TableHead>
                      <TableHead className="text-right">Thao tác</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {unifiedItems.map((item) => {
                      const statusConfig =
                        APPROVAL_STATUS_CONFIG[item.status] ?? {
                          label: item.status,
                          variant: "secondary" as const,
                        };
                      return (
                        <TableRow key={`${item.type}-${item.id}`}>
                          <TableCell>
                            <Badge
                              variant={
                                item.type === "correction"
                                  ? "default"
                                  : "secondary"
                              }
                              className={
                                item.type === "leave"
                                  ? "bg-blue-100 text-blue-700 hover:bg-blue-100"
                                  : ""
                              }
                            >
                              {item.type === "correction"
                                ? "Bù công"
                                : "Nghỉ phép"}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <div className="font-medium">{item.userName}</div>
                            <div className="text-xs text-muted-foreground">
                              {item.employeeCode}
                            </div>
                          </TableCell>
                          <TableCell className="text-sm">{item.date}</TableCell>
                          <TableCell className="text-sm">{item.detail}</TableCell>
                          <TableCell className="text-sm max-w-[200px] truncate">
                            {item.description}
                          </TableCell>
                          <TableCell className="text-sm">
                            {formatDateTime(item.createdAt)}
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
                                onClick={() => {
                                  if (item.type === "correction") {
                                    setDetailCorrection(
                                      item.raw as AttendanceCorrection
                                    );
                                  } else {
                                    setDetailLeave(item.raw as LeaveRequest);
                                  }
                                }}
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
                                        id: item.id,
                                        type: item.type,
                                        action: "approved",
                                        userName: item.userName,
                                        description: item.description,
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
                                        id: item.id,
                                        type: item.type,
                                        action: "rejected",
                                        userName: item.userName,
                                        description: item.description,
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
                    {unifiedItems.length === 0 && (
                      <TableRow>
                        <TableCell
                          colSpan={8}
                          className="text-center py-8 text-muted-foreground"
                        >
                          Không có yêu cầu nào
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
                {totalPages > 1 && (
                  <div className="p-4 border-t">
                    <div className="flex items-center justify-between">
                      <p className="text-sm text-muted-foreground">
                        Trang {page} / {totalPages}
                      </p>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={page <= 1}
                          onClick={() => setPage((p) => p - 1)}
                        >
                          Trước
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={page >= totalPages}
                          onClick={() => setPage((p) => p + 1)}
                        >
                          Sau
                        </Button>
                      </div>
                    </div>
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Correction Detail Dialog */}
      <Dialog
        open={!!detailCorrection}
        onOpenChange={() => setDetailCorrection(null)}
      >
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Chi tiết yêu cầu bù công</DialogTitle>
          </DialogHeader>
          {detailCorrection && (
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100">
                  <UserIcon className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <div className="font-medium">
                    {detailCorrection.user?.name}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {detailCorrection.user?.employee_code} &middot;{" "}
                    {detailCorrection.user?.department}
                  </div>
                </div>
              </div>
              <div className="rounded-md border p-3 space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Ngày chấm công</span>
                  <span className="font-medium">
                    {detailCorrection.attendance_log
                      ? formatDate(detailCorrection.attendance_log.date)
                      : "—"}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Trạng thái gốc</span>
                  <span className="font-medium">
                    {ORIGINAL_STATUS_LABEL[detailCorrection.original_status] ??
                      detailCorrection.original_status}
                  </span>
                </div>
                {detailCorrection.attendance_log && (
                  <>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Check-in</span>
                      <span className="font-medium">
                        {formatTime(
                          detailCorrection.attendance_log.check_in_time
                        )}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Check-out</span>
                      <span className="font-medium">
                        {formatTime(
                          detailCorrection.attendance_log.check_out_time
                        )}
                      </span>
                    </div>
                  </>
                )}
              </div>
              <div>
                <div className="text-sm font-medium mb-1">Lý do</div>
                <div className="rounded-md bg-muted p-3 text-sm">
                  {detailCorrection.description}
                </div>
              </div>
              {detailCorrection.processed_at && (
                <div className="rounded-md border p-3 space-y-2 text-sm">
                  <div className="flex items-center gap-2 text-muted-foreground mb-1">
                    <Clock className="h-4 w-4" />
                    <span className="font-medium">Audit Log</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Người duyệt</span>
                    <span className="font-medium">
                      {detailCorrection.processed_by?.name ?? "Hệ thống"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Thời gian</span>
                    <span className="font-medium">
                      {formatDateTime(detailCorrection.processed_at)}
                    </span>
                  </div>
                  {detailCorrection.manager_note && (
                    <div>
                      <span className="text-muted-foreground">Ghi chú: </span>
                      <span className="italic">
                        {detailCorrection.manager_note}
                      </span>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Leave Detail Dialog */}
      <Dialog open={!!detailLeave} onOpenChange={() => setDetailLeave(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Chi tiết yêu cầu nghỉ phép</DialogTitle>
          </DialogHeader>
          {detailLeave && (
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100">
                  <UserIcon className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <div className="font-medium">{detailLeave.user?.name}</div>
                  <div className="text-sm text-muted-foreground">
                    {detailLeave.user?.employee_code} &middot;{" "}
                    {detailLeave.user?.department}
                  </div>
                </div>
              </div>
              <div className="rounded-md border p-3 space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Ngày nghỉ</span>
                  <span className="font-medium">
                    {formatDate(detailLeave.leave_date)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Loại nghỉ phép</span>
                  <span className="font-medium">
                    {LEAVE_TYPE_LABEL[detailLeave.leave_type] ??
                      detailLeave.leave_type}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Khung giờ</span>
                  <span className="font-medium">
                    {detailLeave.time_from} - {detailLeave.time_to}
                  </span>
                </div>
                {detailLeave.original_status && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">
                      Trạng thái gốc
                    </span>
                    <span className="font-medium">
                      {ORIGINAL_STATUS_LABEL[detailLeave.original_status] ??
                        detailLeave.original_status}
                    </span>
                  </div>
                )}
              </div>
              <div>
                <div className="text-sm font-medium mb-1">Lý do</div>
                <div className="rounded-md bg-muted p-3 text-sm">
                  {detailLeave.description}
                </div>
              </div>
              {detailLeave.processed_at && (
                <div className="rounded-md border p-3 space-y-2 text-sm">
                  <div className="flex items-center gap-2 text-muted-foreground mb-1">
                    <Clock className="h-4 w-4" />
                    <span className="font-medium">Audit Log</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Người duyệt</span>
                    <span className="font-medium">
                      {detailLeave.processed_by?.name ?? "Hệ thống"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Thời gian</span>
                    <span className="font-medium">
                      {formatDateTime(detailLeave.processed_at)}
                    </span>
                  </div>
                  {detailLeave.manager_note && (
                    <div>
                      <span className="text-muted-foreground">Ghi chú: </span>
                      <span className="italic">
                        {detailLeave.manager_note}
                      </span>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Process Dialog (unified) */}
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
              {processItem?.action === "approved" ? "Duyệt" : "Từ chối"} yêu
              cầu{" "}
              {processItem?.type === "correction" ? "bù công" : "nghỉ phép"}
            </DialogTitle>
          </DialogHeader>
          {processItem && (
            <div className="space-y-4">
              <div className="text-sm">
                <span className="text-muted-foreground">Nhân viên: </span>
                <span className="font-medium">{processItem.userName}</span>
              </div>
              <div className="text-sm">
                <span className="text-muted-foreground">Lý do: </span>
                <span>{processItem.description}</span>
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">
                  Ghi chú (không bắt buộc)
                </label>
                <Input
                  placeholder={
                    processItem.action === "approved"
                      ? "Ví dụ: Đã xác nhận..."
                      : "Ví dụ: Không đủ điều kiện..."
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
              disabled={isPending}
            >
              {isPending
                ? "Đang xử lý..."
                : processItem?.action === "approved"
                  ? "Duyệt"
                  : "Từ chối"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Batch Approve Confirm Dialog */}
      <Dialog open={showBatchConfirm} onOpenChange={setShowBatchConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Xác nhận duyệt tất cả</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 text-sm">
            <p>
              Bạn có chắc muốn duyệt tất cả{" "}
              <strong>{totalItems}</strong> yêu cầu đang chờ?
            </p>
            <p className="text-muted-foreground">
              Hành động này không thể hoàn tác.
            </p>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowBatchConfirm(false)}
            >
              Huỷ
            </Button>
            <Button
              className="bg-green-600 hover:bg-green-700"
              disabled={
                batchApproveCorrMutation.isPending ||
                batchApproveLeavesMutation.isPending
              }
              onClick={() => {
                batchApproveCorrMutation.mutate(undefined, {
                  onSettled: () => {
                    batchApproveLeavesMutation.mutate(undefined, {
                      onSettled: () => setShowBatchConfirm(false),
                    });
                  },
                });
              }}
            >
              {batchApproveCorrMutation.isPending ||
              batchApproveLeavesMutation.isPending
                ? "Đang xử lý..."
                : "Duyệt tất cả"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
