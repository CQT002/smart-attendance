import apiClient, { buildQueryString } from "@/lib/api-client";
import { ApiResponse, PaginatedResponse } from "@/types/api";
import { User, CreateUserRequest, UpdateUserRequest, UserFilter } from "@/types/user";

export const userService = {
  async getList(filter: UserFilter = {}): Promise<PaginatedResponse<User>> {
    const qs = buildQueryString(filter as Record<string, unknown>);
    const res = await apiClient.get<ApiResponse<User[]>>(`/admin/users${qs}`);
    return { data: res.data.data ?? [], meta: res.data.meta! };
  },

  async getById(id: number): Promise<User> {
    const res = await apiClient.get<ApiResponse<User>>(`/admin/users/${id}`);
    return res.data.data!;
  },

  async create(data: CreateUserRequest): Promise<User> {
    const res = await apiClient.post<ApiResponse<User>>("/admin/users", data);
    return res.data.data!;
  },

  async update(id: number, data: UpdateUserRequest): Promise<User> {
    const res = await apiClient.put<ApiResponse<User>>(`/admin/users/${id}`, data);
    return res.data.data!;
  },

  async delete(id: number): Promise<void> {
    await apiClient.delete(`/admin/users/${id}`);
  },

  async resetPassword(id: number, newPassword: string): Promise<void> {
    await apiClient.post(`/admin/users/${id}/reset-password`, {
      new_password: newPassword,
    });
  },
};
