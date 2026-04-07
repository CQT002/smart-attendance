"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { correctionService } from "@/services/correction.service";
import { CorrectionFilter, ProcessCorrectionRequest } from "@/types/correction";
import { toast } from "sonner";

export function useCorrections(filter: CorrectionFilter = {}) {
  return useQuery({
    queryKey: ["corrections", filter],
    queryFn: () => correctionService.getList(filter),
    placeholderData: (prev) => prev,
  });
}

export function useCorrection(id: number) {
  return useQuery({
    queryKey: ["corrections", id],
    queryFn: () => correctionService.getById(id),
    enabled: !!id,
  });
}

export function useBatchApproveCorrections() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: () => correctionService.batchApprove(),
    onSuccess: (data) => {
      toast.success(`Đã duyệt ${data.approved_count} yêu cầu bù công`);
      qc.invalidateQueries({ queryKey: ["corrections"] });
    },
    onError: () => {
      toast.error("Duyệt hàng loạt thất bại");
    },
  });
}

export function useProcessCorrection() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: number; req: ProcessCorrectionRequest }) =>
      correctionService.process(id, req),
    onSuccess: (data) => {
      const action = data.status === "approved" ? "duyệt" : "từ chối";
      toast.success(`Đã ${action} yêu cầu bù công`);
      qc.invalidateQueries({ queryKey: ["corrections"] });
    },
    onError: () => {
      toast.error("Xử lý yêu cầu thất bại");
    },
  });
}
