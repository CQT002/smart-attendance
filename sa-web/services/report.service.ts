import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import {
  DashboardStats,
  BranchTodayStats,
  EmployeeTodayDetail,
  UserAttendanceSummary,
  BranchAttendanceReport,
  ReportFilter,
  TodayStatsFilter,
} from "@/types/report";

export const reportService = {
  async getDashboardStats(): Promise<DashboardStats> {
    const res = await apiClient.get<ApiResponse<DashboardStats>>("/admin/reports/dashboard");
    return res.data.data!;
  },

  async getTodayBranchStats(filter: TodayStatsFilter = {}): Promise<PaginatedResponse<BranchTodayStats>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<BranchTodayStats[]>>(`/admin/reports/today${qs}`);
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getTodayEmployees(filter: {
    branch_id?: number;
    status?: string;
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<EmployeeTodayDetail>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<EmployeeTodayDetail[]>>(
      `/admin/reports/today/employees${qs}`
    );
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getAttendanceReport(filter: ReportFilter = {}): Promise<PaginatedResponse<UserAttendanceSummary>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<UserAttendanceSummary[]>>(
      `/admin/reports/attendance${qs}`
    );
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getBranchReport(filter: ReportFilter = {}): Promise<BranchAttendanceReport[]> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<BranchAttendanceReport[]>>(
      `/admin/reports/branches${qs}`
    );
    return res.data.data ?? [];
  },

  async getUserReport(userId: number, from: string, to: string): Promise<UserAttendanceSummary> {
    const res = await apiClient.get<ApiResponse<UserAttendanceSummary>>(
      `/admin/reports/users/${userId}?date_from=${from}&date_to=${to}`
    );
    return res.data.data!;
  },
};
