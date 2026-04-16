"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { useAttendanceReport, useBranchReport } from "@/hooks/use-reports";
import { useActiveBranches } from "@/hooks/use-branches";
import { useCurrentUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Search, TrendingUp } from "lucide-react";
import { ReportFilter, ReportPeriod } from "@/types/report";
import { formatPercent, formatHours } from "@/lib/utils";
import { BranchAutocomplete } from "@/components/shared/branch-autocomplete";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";

const PERIOD_OPTIONS: { value: ReportPeriod; label: string }[] = [
  { value: "daily", label: "Hàng ngày" },
  { value: "weekly", label: "Hàng tuần" },
  { value: "monthly", label: "Hàng tháng" },
  { value: "custom", label: "Tuỳ chỉnh" },
];

function RateCell({ value }: { value: number }) {
  const color =
    value >= 90 ? "text-green-600" : value >= 70 ? "text-yellow-600" : "text-red-600";
  return <span className={`font-medium ${color}`}>{formatPercent(value)}</span>;
}

export default function ReportsPage() {
  const { data: currentUser } = useCurrentUser();
  const isAdmin = currentUser?.role === "admin";

  const [activeTab, setActiveTab] = useState<"employee" | "branch">("employee");
  const [filter, setFilter] = useState<ReportFilter>({
    period: "monthly",
    page: 1,
    limit: 10,
  });
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  // Manager: auto-filter theo chi nhánh của mình
  const effectiveFilter = !isAdmin && currentUser?.branch_id
    ? { ...filter, branch_id: currentUser.branch_id }
    : filter;

  const { data: branches } = useActiveBranches();
  const { data: empReport, isLoading: empLoading } = useAttendanceReport(
    effectiveFilter,
    activeTab === "employee"
  );
  const { data: branchReport, isLoading: branchLoading } = useBranchReport(
    effectiveFilter,
    activeTab === "branch"
  );

  const applyFilter = () => {
    setFilter((f) => ({
      ...f,
      date_from: dateFrom || undefined,
      date_to: dateTo || undefined,
      page: 1,
    }));
  };

  const barData =
    branchReport?.slice(0, 10).map((b) => ({
      name: b.branch_code,
      "Tỷ lệ chuyên cần": b.attendance_rate,
      "Tỷ lệ đúng giờ": b.on_time_rate,
    })) ?? [];

  return (
    <div>
      <Header title="Báo cáo Chấm công" />
      <div className="p-6 space-y-4">
        {/* Filters */}
        <div className="flex flex-wrap items-end gap-3">
          <Select
            value={filter.period}
            onValueChange={(v) =>
              setFilter((f) => ({ ...f, period: v as ReportPeriod, page: 1 }))
            }
          >
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PERIOD_OPTIONS.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {filter.period === "custom" && (
            <>
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
            </>
          )}

          {isAdmin ? (
            <BranchAutocomplete
              branches={branches}
              value={filter.branch_id}
              onChange={(id) => setFilter((f) => ({ ...f, branch_id: id, page: 1 }))}
              placeholder="Tìm kiếm chi nhánh"
              className="w-64"
            />
          ) : (
            <Input
              value={branches?.find((b) => b.id === currentUser?.branch_id)?.name ?? "Chi nhánh của tôi"}
              disabled
              className="w-44"
            />
          )}

          <Button onClick={applyFilter}>
            <Search className="h-4 w-4 mr-1" />
            Xem báo cáo
          </Button>
        </div>

        {/* Tab selector */}
        <div className="flex gap-2">
          <Button
            variant={activeTab === "employee" ? "default" : "outline"}
            size="sm"
            onClick={() => setActiveTab("employee")}
          >
            Theo nhân viên
          </Button>
          <Button
            variant={activeTab === "branch" ? "default" : "outline"}
            size="sm"
            onClick={() => setActiveTab("branch")}
          >
            Theo chi nhánh
          </Button>
        </div>

        {/* Branch chart */}
        {activeTab === "branch" && barData.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <TrendingUp className="h-4 w-4" />
                Biểu đồ tỷ lệ chuyên cần theo chi nhánh
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={260}>
                <BarChart data={barData}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis dataKey="name" className="text-xs" />
                  <YAxis domain={[0, 100]} unit="%" className="text-xs" />
                  <Tooltip formatter={(v: number) => formatPercent(v)} />
                  <Legend />
                  <Bar dataKey="Tỷ lệ chuyên cần" fill="#22c55e" radius={[4, 4, 0, 0]} />
                  <Bar dataKey="Tỷ lệ đúng giờ" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        )}

        {/* Employee report table */}
        {activeTab === "employee" && (
          <Card>
            <CardContent className="p-0">
              {empLoading ? (
                <DataTableSkeleton columns={8} />
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Nhân viên</TableHead>
                        <TableHead>{isAdmin ? "Chi nhánh" : "Phòng ban"}</TableHead>
                        <TableHead>Ngày công</TableHead>
                        <TableHead>Đúng giờ</TableHead>
                        <TableHead>Đi trễ - Về sớm</TableHead>
                        <TableHead>Nghỉ phép</TableHead>
                        <TableHead>Thiếu</TableHead>
                        <TableHead>Giờ làm</TableHead>
                        <TableHead>Chuyên cần</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {empReport?.data.map((r) => (
                        <TableRow key={r.user_id}>
                          <TableCell>
                            <div className="font-medium">{r.user_name || r.user?.name || `#${r.user_id}`}</div>
                            <div className="text-xs text-muted-foreground">
                              {r.employee_code || r.user?.employee_code}
                            </div>
                          </TableCell>
                          <TableCell className="text-sm">
                            {isAdmin
                              ? (() => {
                                  const b = branches?.find((x) => x.id === filter.branch_id);
                                  return b ? `${b.name} (${b.code})` : r.user?.branch?.name || "—";
                                })()
                              : r.department || r.user?.department || "—"}
                          </TableCell>
                          <TableCell className="text-sm">
                            {r.present_count + r.late_count + r.early_leave_count + r.half_day_count + (r.leave_count ?? 0)}/{r.total_days}
                          </TableCell>
                          <TableCell className="text-sm text-green-600">{r.present_count}</TableCell>
                          <TableCell className="text-sm text-yellow-600">{r.late_count + r.early_leave_count}</TableCell>
                          <TableCell className="text-sm text-blue-600">{r.leave_count ?? 0}</TableCell>
                          <TableCell className="text-sm text-orange-600">{r.incomplete_count ?? 0}</TableCell>
                          <TableCell className="text-sm">{formatHours(r.total_work_hours)}</TableCell>
                          <TableCell>
                            <RateCell value={r.attendance_rate} />
                          </TableCell>
                        </TableRow>
                      ))}
                      {empReport?.data.length === 0 && (
                        <TableRow>
                          <TableCell colSpan={10} className="text-center py-8 text-muted-foreground">
                            Không có dữ liệu
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                  {empReport?.meta && empReport.meta.total_pages > 1 && (
                    <Pagination
                      meta={empReport.meta}
                      onPageChange={(p) => setFilter((f) => ({ ...f, page: p }))}
                    />
                  )}
                </>
              )}
            </CardContent>
          </Card>
        )}

        {/* Branch report table */}
        {activeTab === "branch" && (
          <Card>
            <CardContent className="p-0">
              {branchLoading ? (
                <DataTableSkeleton columns={7} />
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Chi nhánh</TableHead>
                      <TableHead>Ngày công</TableHead>
                      <TableHead>Có mặt</TableHead>
                      <TableHead>Đi trễ - Về sớm</TableHead>
                      <TableHead>Vắng</TableHead>
                      <TableHead>Giờ làm</TableHead>
                      <TableHead>Chuyên cần</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {branchReport?.map((b) => (
                      <TableRow key={b.branch_id}>
                        <TableCell>
                          <div className="font-medium">{b.branch_name}</div>
                          <div className="text-xs text-muted-foreground">{b.branch_code}</div>
                        </TableCell>
                        <TableCell className="text-sm">{b.total_days}</TableCell>
                        <TableCell className="text-sm text-green-600">{b.present_count}</TableCell>
                        <TableCell className="text-sm text-yellow-600">{b.late_count}</TableCell>
                        <TableCell className="text-sm text-red-600">{b.absent_count}</TableCell>
                        <TableCell className="text-sm">{formatHours(b.total_work_hours)}</TableCell>
                        <TableCell>
                          <RateCell value={b.attendance_rate} />
                        </TableCell>
                      </TableRow>
                    ))}
                    {branchReport?.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                          Không có dữ liệu
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
