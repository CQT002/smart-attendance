"use client";

import { useState } from "react";
import { Header } from "@/components/layout/header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useDashboardStats, useTodayBranchStats } from "@/hooks/use-reports";
import { useActiveBranches } from "@/hooks/use-branches";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
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

const COLORS = ["#22c55e", "#f59e0b", "#ef4444", "#6366f1", "#8b5cf6"];

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
  const { data: branches } = useActiveBranches();
  const [branchId, setBranchId] = useState<number | undefined>();
  const [page, setPage] = useState(1);

  const { data: todayStats, isLoading: todayLoading } = useTodayBranchStats({
    branch_id: branchId,
    page,
    limit: 10,
  });

  const pieData = stats
    ? [
        { name: "Đúng giờ", value: stats.present_today - stats.late_today },
        { name: "Đi trễ", value: stats.late_today },
        { name: "Vắng mặt", value: stats.absent_today },
      ]
    : [];

  const barData =
    todayStats?.data.slice(0, 8).map((b) => ({
      name: b.branch_code,
      "Đúng giờ": b.present_count - b.late_count,
      "Đi trễ": b.late_count,
      "Vắng mặt": b.absent_count,
    })) ?? [];

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
            title="Đi trễ"
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
                      {pieData.map((_, index) => (
                        <Cell key={index} fill={COLORS[index]} />
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
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-base">Thống kê theo chi nhánh hôm nay</CardTitle>
              <Select
                value={branchId?.toString() ?? "all"}
                onValueChange={(v) => {
                  setBranchId(v === "all" ? undefined : Number(v));
                  setPage(1);
                }}
              >
                <SelectTrigger className="w-44">
                  <SelectValue placeholder="Tất cả chi nhánh" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Tất cả chi nhánh</SelectItem>
                  {branches?.map((b) => (
                    <SelectItem key={b.id} value={b.id.toString()}>
                      {b.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </CardHeader>
            <CardContent>
              {todayLoading ? (
                <Skeleton className="h-64 w-full" />
              ) : (
                <ResponsiveContainer width="100%" height={240}>
                  <BarChart data={barData}>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis dataKey="name" className="text-xs" />
                    <YAxis className="text-xs" />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="Đúng giờ" fill="#22c55e" radius={[4, 4, 0, 0]} />
                    <Bar dataKey="Đi trễ" fill="#f59e0b" radius={[4, 4, 0, 0]} />
                    <Bar dataKey="Vắng mặt" fill="#ef4444" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Today branch table */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Bảng thống kê chi nhánh</CardTitle>
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
                        {["Chi nhánh", "Tổng NV", "Có mặt", "Đi trễ", "Vắng", "Tỷ lệ"].map(
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
                          <td className="px-4 py-3 text-green-600">{b.present_count}</td>
                          <td className="px-4 py-3 text-yellow-600">{b.late_count}</td>
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
