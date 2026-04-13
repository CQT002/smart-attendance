import apiClient from "@/lib/api-client";
import { ApiResponse } from "@/types/api";

export interface Shift {
  id: number;
  branch_id: number;
  name: string;
  start_time: string;
  end_time: string;
  late_after: number;
  early_before: number;
  work_hours: number;
  morning_end: string;
  afternoon_start: string;
  regular_end_day: number;
  regular_end_time: string;
  ot_min_checkin_hour: number;
  ot_start_hour: number;
  ot_end_hour: number;
  is_default: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateShiftRequest {
  name: string;
  start_time: string;
  end_time: string;
  late_after?: number;
  early_before?: number;
  work_hours?: number;
  morning_end?: string;
  afternoon_start?: string;
  regular_end_day?: number;
  regular_end_time?: string;
  ot_min_checkin_hour?: number;
  ot_start_hour?: number;
  ot_end_hour?: number;
  is_default?: boolean;
}

export interface UpdateShiftRequest {
  name?: string;
  start_time?: string;
  end_time?: string;
  late_after?: number;
  early_before?: number;
  work_hours?: number;
  morning_end?: string;
  afternoon_start?: string;
  regular_end_day?: number;
  regular_end_time?: string;
  ot_min_checkin_hour?: number;
  ot_start_hour?: number;
  ot_end_hour?: number;
  is_default?: boolean;
  is_active?: boolean;
}

/** Tên thứ trong tuần (index = Go time.Weekday: 0=CN, 1=T2, ..., 6=T7) */
export const DAY_NAMES = [
  "Chủ nhật",
  "Thứ 2",
  "Thứ 3",
  "Thứ 4",
  "Thứ 5",
  "Thứ 6",
  "Thứ 7",
];

export const shiftService = {
  async getByBranch(branchId: number): Promise<Shift[]> {
    const res = await apiClient.get<ApiResponse<Shift[]>>(
      `/admin/branches/${branchId}/shifts`
    );
    return res.data.data ?? [];
  },

  async create(branchId: number, data: CreateShiftRequest): Promise<Shift> {
    const res = await apiClient.post<ApiResponse<Shift>>(
      `/admin/branches/${branchId}/shifts`,
      data
    );
    return res.data.data!;
  },

  async update(
    branchId: number,
    id: number,
    data: UpdateShiftRequest
  ): Promise<Shift> {
    const res = await apiClient.put<ApiResponse<Shift>>(
      `/admin/branches/${branchId}/shifts/${id}`,
      data
    );
    return res.data.data!;
  },

  async delete(branchId: number, id: number): Promise<void> {
    await apiClient.delete(`/admin/branches/${branchId}/shifts/${id}`);
  },
};
