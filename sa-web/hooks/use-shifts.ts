"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import {
  shiftService,
  CreateShiftRequest,
  UpdateShiftRequest,
} from "@/services/shift.service";
import { toast } from "sonner";

function getErrorMessage(error: unknown): string {
  if (error instanceof AxiosError) {
    return error.response?.data?.error?.message || error.message;
  }
  return "Lỗi không xác định";
}

export function useShifts(branchId: number) {
  return useQuery({
    queryKey: ["shifts", branchId],
    queryFn: () => shiftService.getByBranch(branchId),
    enabled: !!branchId,
  });
}

export function useCreateShift(branchId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateShiftRequest) =>
      shiftService.create(branchId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["shifts", branchId] });
      qc.invalidateQueries({ queryKey: ["branches"] });
      toast.success("Thêm ca làm việc thành công");
    },
    onError: (e) => toast.error("Thêm ca làm việc thất bại: " + getErrorMessage(e)),
  });
}

export function useUpdateShift(branchId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateShiftRequest }) =>
      shiftService.update(branchId, id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["shifts", branchId] });
      toast.success("Cập nhật ca làm việc thành công");
    },
    onError: (e) => toast.error("Cập nhật thất bại: " + getErrorMessage(e)),
  });
}

export function useDeleteShift(branchId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => shiftService.delete(branchId, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["shifts", branchId] });
      qc.invalidateQueries({ queryKey: ["branches"] });
      toast.success("Xóa ca làm việc thành công");
    },
    onError: (e) => toast.error("Xóa thất bại: " + getErrorMessage(e)),
  });
}
