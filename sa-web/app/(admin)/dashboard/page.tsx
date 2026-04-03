"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useDashboardStats, useTodayBranchStats } from "@/hooks/use-reports";
import { useDebounce } from "@/hooks/use-debounce";
import { Input } from "@/components/ui/input";
import { formatPercent, formatHours } from "@/lib/utils";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from "recharts";
import {
  Users,
  Building2,
  CheckCircle2,
  XCircle,
  Clock,
  TrendingUp,
} from "lucide-react";
import { Pagination } from "@/components/shared/pagination";

// Màu cố định: Đúng giờ = xanh lá, Đi trễ - Về sớm = cam, Vắng mặt = đỏ
const STATUS_COLORS: Record<string, string> = {
  "Đúng giờ": "#22c55e",
  "Đi trễ - Về sớm": "#f59e0b",
  "Vắng mặt": "#ef4444",
};

function StatCard({
  title,
  value,
  sub,
  icon: Icon,
  color,
  loading,
}: {
  title: string;
  value: string | number;
  sub?: string;
  icon: React.ElementType;
  color: string;
  loading?: boolean;
}) {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{title}</p>
            {loading ? (
              <Skeleton className="mt-2 h-8 w-20" />
            ) : (
              <p className="mt-1 text-3xl font-bold">{value}</p>
            )}
            {sub && !loading && (
              <p className="mt-1 text-xs text-muted-foreground">{sub}</p>
            )}
          </div>
          <div className={`rounded-lg p-2 ${color}`}>
            <Icon className="h-6 w-6 text-white" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats();
  const [search, setSearch] = useState("");
  const searchDebounced = useDebounce(search, 500);
  const [page, setPage] = useState(1);

  const [chartSearch, setChartSearch] = useState("");
  const chartSearchDebounced = useDebounce(chartSearch, 500);

  const { data: chartStats, isLoading: chartLoading } = useTodayBranchStats({
    search: chartSearchDebounced,
    page: 1,
    limit: 100,
  });

  const { data: todayStats, isLoading: todayLoading } = useTodayBranchStats({
    search: searchDebounced,
    page,
    limit: 10,
  });

  // Gom: late + early_leave + late_early_leave + half_day → "Đi trễ - Về sớm"
  // present_today = tổng có mặt, late_today = late (chưa gồm early_leave)
  // Đúng giờ = present_today - late_today (present_today đã bao gồm late)
  const lateEarlyTotal = (stats?.late_today ?? 0);
  const pieData = stats
    ? [
        { name: "Đúng giờ", value: Math.max(0, stats.present_today - lateEarlyTotal) },
        { name: "Đi trễ - Về sớm", value: lateEarlyTotal },
        { name: "Vắng mặt", value: stats.absent_today },
      ].filter((d) => d.value > 0)
    : [];

  // Stacked bar: gom late + early_leave + half_day → "Đi trễ - Về sớm"
  const barData =
    chartStats?.data.map((b) => {
      const total = b.total_employees || 1;
      const lateEarly = b.late_count + b.early_leave_count + b.half_day_count;
      return {
        name: b.branch_code,
        "Đúng giờ": Math.round(b.present_count / total * 100),
        "Đi trễ - Về sớm": Math.round(lateEarly / total * 100),
        "Vắng mặt": Math.round(b.absent_count / total * 100),
      };
    }) ?? [];

  return (
    <div>
      <Header title="Dashboard" />
      <div className="p-6 space-y-6">
        {/* KPI Cards */}
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3 xl:grid-cols-6">
          <StatCard
            title="Tổng chi nhánh"
            value={stats?.total_branches ?? 0}
            icon={Building2}
            color="bg-blue-500"
            loading={statsLoading}
          />
          <StatCard
            title="Tổng nhân viên"
            value={stats?.total_employees ?? 0}
            icon={Users}
            color="bg-indigo-500"
            loading={statsLoading}
          />
          <StatCard
            title="Có mặt hôm nay"
            value={stats?.present_today ?? 0}
            icon={CheckCircle2}
            color="bg-green-500"
            loading={statsLoading}
          />
          <StatCard
            title="Vắng mặt"
            value={stats?.absent_today ?? 0}
            icon={XCircle}
            color="bg-red-500"
            loading={statsLoading}
          />
          <StatCard
            title="Đi trễ - Về sớm"
            value={stats?.late_today ?? 0}
            icon={Clock}
            color="bg-yellow-500"
            loading={statsLoading}
          />
          <StatCard
            title="Tỷ lệ chuyên cần"
            value={stats ? formatPercent(stats.attendance_rate) : "—"}
            sub={`Đúng giờ: ${stats ? formatPercent(stats.on_time_rate) : "—"}`}
            icon={TrendingUp}
            color="bg-purple-500"
            loading={statsLoading}
          />
        </div>

        {/* Charts */}
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Pie chart */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Phân bố trạng thái hôm nay</CardTitle>
            </CardHeader>
            <CardContent>
              {statsLoading ? (
                <Skeleton className="h-64 w-full" />
              ) : (
                <ResponsiveContainer width="100%" height={240}>
                  <PieChart>
                    <Pie
                      data={pieData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={90}
                      paddingAngle={3}
                      dataKey="value"
                    >
                      {pieData.map((entry, index) => (
                        <Cell key={index} fill={STATUS_COLORS[entry.name] ?? "#6366f1"} />
                      ))}
                    </Pie>
                    <Tooltip />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>

          {/* Bar chart */}
          <Card className="lg:col-span-2">
            <CardHeader className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <CardTitle className="text-base">Thống kê theo chi nhánh hôm nay</CardTitle>
              <Input
                placeholder="Tìm theo tên/code..."
                value={chartSearch}
                onChange={(e) => setChartSearch(e.target.value)}
                className="w-full sm:w-52"
              />
            </CardHeader>
            <CardContent>
              {chartLoading ? (
                <Skeleton className="h-64 w-full" />
              ) : (
                <>
                  <div className="overflow-x-auto pb-2">
                    <div style={{ minWidth: barData.length > 5 ? barData.length * 80 + "px" : undefined }}>
                      <ResponsiveContainer width="100%" height={240}>
                        <BarChart data={barData} maxBarSize={60}>
                          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                          <XAxis dataKey="name" className="text-xs" tick={{ fontSize: 11 }} />
                          <YAxis className="text-xs" domain={[0, 100]} unit="%" allowDecimals={false} />
                          <Tooltip formatter={(v: number) => `${v}%`} />
                          <Bar dataKey="Đúng giờ" stackId="a" fill="#22c55e" />
                          <Bar dataKey="Đi trễ - Về sớm" stackId="a" fill="#f59e0b" />
                          <Bar dataKey="Vắng mặt" stackId="a" fill="#ef4444" radius={[4, 4, 0, 0]} />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  </div>
                  <div className="flex items-center justify-center gap-6 pt-2 text-sm">
                    <div className="flex items-center gap-1.5"><div className="h-3 w-3 rounded-sm bg-[#22c55e]" />Đúng giờ</div>
                    <div className="flex items-center gap-1.5"><div className="h-3 w-3 rounded-sm bg-[#f59e0b]" />Đi trễ - Về sớm</div>
                    <div className="flex items-center gap-1.5"><div className="h-3 w-3 rounded-sm bg-[#ef4444]" />Vắng mặt</div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Today branch table */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between border-b p-6">
            <CardTitle className="text-base">Bảng thống kê chi nhánh</CardTitle>
            <Input
              placeholder="Tìm kiếm theo tên chi nhánh..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              className="max-w-xs"
            />
          </CardHeader>
          <CardContent className="p-0">
            {todayLoading ? (
              <div className="p-6 space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-12 w-full" />
                ))}
              </div>
            ) : (
              <>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="border-b">
                      <tr>
                        {["Chi nhánh", "Tổng NV", "Có mặt", "Đi trễ - Về sớm", "Vắng", "Tỷ lệ"].map(
                          (h) => (
                            <th key={h} className="h-12 px-4 text-left font-medium text-muted-foreground">
                              {h}
                            </th>
                          )
                        )}
                      </tr>
                    </thead>
                    <tbody>
                      {todayStats?.data.map((b) => (
                        <tr key={b.branch_id} className="border-b hover:bg-muted/50">
                          <td className="px-4 py-3 font-medium">
                            <div>{b.branch_name}</div>
                            <div className="text-xs text-muted-foreground">{b.branch_code}</div>
                          </td>
                          <td className="px-4 py-3">{b.total_employees}</td>
                          <td className="px-4 py-3 text-green-600">{b.present_count + b.late_count + b.early_leave_count + b.half_day_count}</td>
                          <td className="px-4 py-3 text-yellow-600">{b.late_count + b.early_leave_count + b.half_day_count}</td>
                          <td className="px-4 py-3 text-red-600">{b.absent_count}</td>
                          <td className="px-4 py-3">
                            <div className="flex items-center gap-2">
                              <div className="h-2 flex-1 rounded-full bg-muted overflow-hidden">
                                <div
                                  className="h-full rounded-full bg-green-500"
                                  style={{ width: `${b.attendance_rate}%` }}
                                />
                              </div>
                              <span className="text-xs w-12 text-right">
                                {formatPercent(b.attendance_rate)}
                              </span>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                {todayStats?.meta && todayStats.meta.total_pages > 1 && (
                  <Pagination meta={todayStats.meta} onPageChange={setPage} />
                )}
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
