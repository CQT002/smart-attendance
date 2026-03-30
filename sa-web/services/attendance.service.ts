import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import { AttendanceLog, AttendanceSummary, AttendanceFilter } from "@/types/attendance";

export const attendanceService = {
  async getList(filter: AttendanceFilter = {}): Promise<PaginatedResponse<AttendanceLog>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<AttendanceLog[]>>(`/admin/attendance${qs}`);
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getById(id: number): Promise<AttendanceLog> {
    const res = await apiClient.get<ApiResponse<AttendanceLog>>(`/admin/attendance/${id}`);
    return res.data.data!;
  },

  async getSummary(userId: number, from: string, to: string): Promise<AttendanceSummary> {
    const res = await apiClient.get<ApiResponse<AttendanceSummary>>(
      `/admin/attendance/summary/${userId}?date_from=${from}&date_to=${to}`
    );
    return res.data.data!;
  },
};
