"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { overtimeService } from "@/services/overtime.service";
import { OvertimeFilter, ProcessOvertimeRequest } from "@/types/overtime";
import { toast } from "sonner";

export function useOvertime(filter: OvertimeFilter = {}) {
  return useQuery({
    queryKey: ["overtime", filter],
    queryFn: () => overtimeService.getList(filter),
    placeholderData: (prev) => prev,
  });
}

export function useOvertimeDetail(id: number) {
  return useQuery({
    queryKey: ["overtime", id],
    queryFn: () => overtimeService.getById(id),
    enabled: !!id,
  });
}

export function useBatchApproveOvertime() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: () => overtimeService.batchApprove(),
    onSuccess: (data) => {
      toast.success(`Đã duyệt ${data.approved_count} yêu cầu tăng ca`);
      qc.invalidateQueries({ queryKey: ["overtime"] });
      qc.invalidateQueries({ queryKey: ["approvals"] });
    },
    onError: () => {
      toast.error("Duyệt hàng loạt thất bại");
    },
  });
}

export function useProcessOvertime() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: number; req: ProcessOvertimeRequest }) =>
      overtimeService.process(id, req),
    onSuccess: (data) => {
      const action = data.status === "approved" ? "duyệt" : "từ chối";
      toast.success(`Đã ${action} yêu cầu tăng ca`);
      qc.invalidateQueries({ queryKey: ["overtime"] });
      qc.invalidateQueries({ queryKey: ["approvals"] });
    },
    onError: () => {
      toast.error("Xử lý yêu cầu thất bại");
    },
  });
}
