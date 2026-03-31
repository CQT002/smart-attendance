"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  wifiConfigService,
  CreateWiFiConfigRequest,
  UpdateWiFiConfigRequest,
} from "@/services/wifi-config.service";
import { toast } from "sonner";

export function useWifiConfigs(branchId: number) {
  return useQuery({
    queryKey: ["wifi-configs", branchId],
    queryFn: () => wifiConfigService.getByBranch(branchId),
    enabled: !!branchId,
  });
}

export function useCreateWifiConfig(branchId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateWiFiConfigRequest) =>
      wifiConfigService.create(branchId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["wifi-configs", branchId] });
      qc.invalidateQueries({ queryKey: ["branches"] });
      toast.success("Thêm WiFi config thành công");
    },
    onError: () => toast.error("Thêm WiFi config thất bại"),
  });
}

export function useUpdateWifiConfig(branchId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateWiFiConfigRequest }) =>
      wifiConfigService.update(branchId, id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["wifi-configs", branchId] });
      toast.success("Cập nhật WiFi config thành công");
    },
    onError: () => toast.error("Cập nhật thất bại"),
  });
}

export function useDeleteWifiConfig(branchId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => wifiConfigService.delete(branchId, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["wifi-configs", branchId] });
      qc.invalidateQueries({ queryKey: ["branches"] });
      toast.success("Xóa WiFi config thành công");
    },
    onError: () => toast.error("Xóa thất bại"),
  });
}
