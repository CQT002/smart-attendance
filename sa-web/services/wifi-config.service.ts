import apiClient from "@/lib/api-client";
import { ApiResponse } from "@/types/api";

export interface WiFiConfig {
  id: number;
  branch_id: number;
  ssid: string;
  bssid: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateWiFiConfigRequest {
  ssid: string;
  bssid?: string;
  description?: string;
}

export interface UpdateWiFiConfigRequest {
  ssid?: string;
  bssid?: string;
  description?: string;
  is_active?: boolean;
}

export const wifiConfigService = {
  async getByBranch(branchId: number): Promise<WiFiConfig[]> {
    const res = await apiClient.get<ApiResponse<WiFiConfig[]>>(
      `/admin/branches/${branchId}/wifi-configs`
    );
    return res.data.data ?? [];
  },

  async create(branchId: number, data: CreateWiFiConfigRequest): Promise<WiFiConfig> {
    const res = await apiClient.post<ApiResponse<WiFiConfig>>(
      `/admin/branches/${branchId}/wifi-configs`,
      data
    );
    return res.data.data!;
  },

  async update(
    branchId: number,
    id: number,
    data: UpdateWiFiConfigRequest
  ): Promise<WiFiConfig> {
    const res = await apiClient.put<ApiResponse<WiFiConfig>>(
      `/admin/branches/${branchId}/wifi-configs/${id}`,
      data
    );
    return res.data.data!;
  },

  async delete(branchId: number, id: number): Promise<void> {
    await apiClient.delete(`/admin/branches/${branchId}/wifi-configs/${id}`);
  },
};
