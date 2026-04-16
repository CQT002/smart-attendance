import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import {
  CreateHolidayRequest,
  Holiday,
  HolidayFilter,
  UpdateHolidayRequest,
} from "@/types/holiday";

export const holidayService = {
  async getList(filter: HolidayFilter = {}): Promise<PaginatedResponse<Holiday>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<Holiday[]>>(`/admin/holidays${qs}`);
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getById(id: number): Promise<Holiday> {
    const res = await apiClient.get<ApiResponse<Holiday>>(`/admin/holidays/${id}`);
    return res.data.data!;
  },

  async create(data: CreateHolidayRequest): Promise<Holiday> {
    const res = await apiClient.post<ApiResponse<Holiday>>("/admin/holidays", data);
    return res.data.data!;
  },

  async update(id: number, data: UpdateHolidayRequest): Promise<Holiday> {
    const res = await apiClient.put<ApiResponse<Holiday>>(`/admin/holidays/${id}`, data);
    return res.data.data!;
  },

  async delete(id: number): Promise<void> {
    await apiClient.delete(`/admin/holidays/${id}`);
  },
};
