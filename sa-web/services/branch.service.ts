import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import { Branch, CreateBranchRequest, UpdateBranchRequest, BranchFilter } from "@/types/branch";

export const branchService = {
  async getList(filter: BranchFilter = {}): Promise<PaginatedResponse<Branch>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<Branch[]>>(`/admin/branches${qs}`);
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getActive(): Promise<Branch[]> {
    const res = await apiClient.get<ApiResponse<Branch[]>>("/admin/branches/active");
    return res.data.data ?? [];
  },

  async getById(id: number): Promise<Branch> {
    const res = await apiClient.get<ApiResponse<Branch>>(`/admin/branches/${id}`);
    return res.data.data!;
  },

  async create(data: CreateBranchRequest): Promise<Branch> {
    const res = await apiClient.post<ApiResponse<Branch>>("/admin/branches", data);
    return res.data.data!;
  },

  async update(id: number, data: UpdateBranchRequest): Promise<Branch> {
    const res = await apiClient.put<ApiResponse<Branch>>(`/admin/branches/${id}`, data);
    return res.data.data!;
  },

  async delete(id: number): Promise<void> {
    await apiClient.delete(`/admin/branches/${id}`);
  },
};
