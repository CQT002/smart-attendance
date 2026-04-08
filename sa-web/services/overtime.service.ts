import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import {
  OvertimeRequest,
  OvertimeFilter,
  ProcessOvertimeRequest,
} from "@/types/overtime";

export const overtimeService = {
  async getList(
    filter: OvertimeFilter = {}
  ): Promise<PaginatedResponse<OvertimeRequest>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<OvertimeRequest[]>>(
      `/admin/overtime${qs}`
    );
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getById(id: number): Promise<OvertimeRequest> {
    const res = await apiClient.get<ApiResponse<OvertimeRequest>>(
      `/admin/overtime/${id}`
    );
    return res.data.data!;
  },

  async process(
    id: number,
    req: ProcessOvertimeRequest
  ): Promise<OvertimeRequest> {
    const res = await apiClient.put<ApiResponse<OvertimeRequest>>(
      `/admin/overtime/${id}/process`,
      req
    );
    return res.data.data!;
  },

  async batchApprove(): Promise<{ approved_count: number }> {
    const res = await apiClient.post<ApiResponse<{ approved_count: number }>>(
      `/admin/overtime/batch-approve`
    );
    return res.data.data!;
  },
};
