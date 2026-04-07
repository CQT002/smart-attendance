import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import {
  AttendanceCorrection,
  CorrectionFilter,
  ProcessCorrectionRequest,
} from "@/types/correction";

export const correctionService = {
  async getList(
    filter: CorrectionFilter = {}
  ): Promise<PaginatedResponse<AttendanceCorrection>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<AttendanceCorrection[]>>(
      `/admin/corrections${qs}`
    );
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getById(id: number): Promise<AttendanceCorrection> {
    const res = await apiClient.get<ApiResponse<AttendanceCorrection>>(
      `/admin/corrections/${id}`
    );
    return res.data.data!;
  },

  async process(
    id: number,
    req: ProcessCorrectionRequest
  ): Promise<AttendanceCorrection> {
    const res = await apiClient.put<ApiResponse<AttendanceCorrection>>(
      `/admin/corrections/${id}/process`,
      req
    );
    return res.data.data!;
  },

  async batchApprove(): Promise<{ approved_count: number }> {
    const res = await apiClient.post<ApiResponse<{ approved_count: number }>>(
      `/admin/corrections/batch-approve`
    );
    return res.data.data!;
  },
};
