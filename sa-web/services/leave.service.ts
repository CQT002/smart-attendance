import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import {
  LeaveRequest,
  LeaveFilter,
  ProcessLeaveRequest,
  PendingApprovalItem,
} from "@/types/leave";

export const leaveService = {
  async getList(
    filter: LeaveFilter = {}
  ): Promise<PaginatedResponse<LeaveRequest>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<LeaveRequest[]>>(
      `/admin/leaves${qs}`
    );
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getById(id: number): Promise<LeaveRequest> {
    const res = await apiClient.get<ApiResponse<LeaveRequest>>(
      `/admin/leaves/${id}`
    );
    return res.data.data!;
  },

  async process(
    id: number,
    req: ProcessLeaveRequest
  ): Promise<LeaveRequest> {
    const res = await apiClient.put<ApiResponse<LeaveRequest>>(
      `/admin/leaves/${id}/process`,
      req
    );
    return res.data.data!;
  },

  async batchApprove(): Promise<{ approved_count: number }> {
    const res = await apiClient.post<ApiResponse<{ approved_count: number }>>(
      `/admin/leaves/batch-approve`
    );
    return res.data.data!;
  },

  async getPendingApprovals(
    page = 1,
    limit = 20
  ): Promise<PaginatedResponse<PendingApprovalItem>> {
    const qs = buildQueryString({ page, limit } as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<PendingApprovalItem[]>>(
      `/admin/approvals/pending${qs}`
    );
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },
};
