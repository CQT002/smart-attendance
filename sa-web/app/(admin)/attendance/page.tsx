"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { useAttendanceLogs } from "@/hooks/use-attendance";
import { useCurrentUser } from "@/hooks/use-auth";
import { useActiveBranches } from "@/hooks/use-branches";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { Badge } from "@/components/ui/badge";
import { Search, AlertTriangle } from "lucide-react";
import { AttendanceFilter, AttendanceStatus } from "@/types/attendance";
import { formatDate, formatTime, formatHours } from "@/lib/utils";

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: "all", label: "Tất cả trạng thái" },
  { value: "present", label: "Đúng giờ" },
  { value: "late_group", label: "Đi trễ - Về sớm" },
  { value: "absent", label: "Vắng mặt" },
  { value: "leave_group", label: "Nghỉ phép" },
  { value: "incomplete:checkout", label: "Thiếu check-out" },
  { value: "incomplete:checkin", label: "Thiếu check-in" },
];

export default function AttendancePage() {
  const { data: currentUser } = useCurrentUser();
  const { data: branches } = useActiveBranches();
  const isManager = currentUser?.role === "manager";
  const managerBranchName = isManager
    ? branches?.find((b) => b.id === currentUser?.branch_id)?.name ?? "Chi nhánh của tôi"
    : "";

  const [filter, setFilter] = useState<AttendanceFilter>({ page: 1, limit: 10 });
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const { data, isLoading } = useAttendanceLogs(filter);

  const applyDateFilter = () => {
    setFilter((f) => ({
      ...f,
      date_from: dateFrom || undefined,
      date_to: dateTo || undefined,
      page: 1,
    }));
  };

  const resetFilter = () => {
    setDateFrom("");
    setDateTo("");
    setFilter({ page: 1, limit: 10, incomplete: undefined });
  };

  return (
    <div>
      <Header title="Dữ liệu Chấm công" />
      <div className="p-6 space-y-4">
        {/* Filters */}
        <div className="flex flex-wrap items-end gap-3">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Từ ngày</label>
            <Input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              className="w-40"
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">Đến ngày</label>
            <Input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
              className="w-40"
            />
          </div>
          {isManager ? (
            <Input
              value={managerBranchName}
              disabled
              className="w-52"
            />
          ) : (
            <Input
              placeholder="Tìm kiếm chi nhánh..."
              value={filter.search ?? ""}
              onChange={(e) =>
                setFilter((f) => ({
                  ...f,
                  search: e.target.value || undefined,
                  page: 1,
                }))
              }
              className="w-52"
            />
          )}
          <Select
            value={filter.incomplete ? `incomplete:${filter.incomplete}` : (filter.status ?? "all")}
            onValueChange={(v) => {
              if (v.startsWith("incomplete:")) {
                setFilter((f) => ({
                  ...f,
                  status: undefined,
                  incomplete: v.split(":")[1],
                  page: 1,
                }));
              } else {
                setFilter((f) => ({
                  ...f,
                  status: v === "all" ? undefined : (v as AttendanceStatus),
                  incomplete: undefined,
                  page: 1,
                }));
              }
            }}
          >
            <SelectTrigger className="w-44">
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
          <Button onClick={applyDateFilter}>
            <Search className="h-4 w-4 mr-1" />
            Tìm kiếm
          </Button>
          <Button variant="outline" onClick={resetFilter}>
            Đặt lại
          </Button>
        </div>

        {data?.meta && (
          <p className="text-sm text-muted-foreground">
            Tổng <strong>{data.meta.total}</strong> bản ghi
          </p>
        )}

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
                      <TableHead>Nhân viên</TableHead>
                      <TableHead>Chi nhánh</TableHead>
                      <TableHead>Ngày</TableHead>
                      <TableHead>Vào</TableHead>
                      <TableHead>Ra</TableHead>
                      <TableHead>Giờ làm</TableHead>
                      <TableHead>Trạng thái</TableHead>
                      <TableHead>Gian lận</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data?.data.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell>
                          <div className="font-medium">{log.user?.name ?? `#${log.user_id}`}</div>
                          <div className="text-xs text-muted-foreground">
                            {log.user?.employee_code}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm">
                          {log.branch?.name ?? `#${log.branch_id}`}
                        </TableCell>
                        <TableCell className="text-sm">{formatDate(log.date)}</TableCell>
                        <TableCell className="text-sm">
                          {log.check_in_time ? (
                            <div>
                              <div>{formatTime(log.check_in_time)}</div>
                              <div className="text-xs text-muted-foreground">
                                {log.check_in_method?.toUpperCase()}
                              </div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">—</span>
                          )}
                        </TableCell>
                        <TableCell className="text-sm">
                          {log.check_out_time ? (
                            <div>
                              <div>{formatTime(log.check_out_time)}</div>
                              <div className="text-xs text-muted-foreground">
                                {log.check_out_method?.toUpperCase()}
                              </div>
                            </div>
                          ) : (
                            <span className="text-muted-foreground">—</span>
                          )}
                        </TableCell>
                        <TableCell className="text-sm">
                          {log.work_hours > 0 ? formatHours(log.work_hours) : "—"}
                        </TableCell>
                        <TableCell>
                          <StatusBadge
                            status={log.status}
                            checkInTime={log.check_in_time}
                            checkOutTime={log.check_out_time}
                          />
                        </TableCell>
                        <TableCell>
                          {(log.is_fake_gps || log.is_vpn) ? (
                            <Badge variant="destructive" className="gap-1">
                              <AlertTriangle className="h-3 w-3" />
                              {log.is_fake_gps ? "GPS giả" : "VPN"}
                            </Badge>
                          ) : (
                            <span className="text-xs text-muted-foreground">—</span>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                    {data?.data.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                          Không có dữ liệu
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
    </div>
  );
}
