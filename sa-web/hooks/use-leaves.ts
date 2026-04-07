"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { leaveService } from "@/services/leave.service";
import { LeaveFilter, ProcessLeaveRequest } from "@/types/leave";
import { toast } from "sonner";

export function useLeaves(filter: LeaveFilter = {}) {
  return useQuery({
    queryKey: ["leaves", filter],
    queryFn: () => leaveService.getList(filter),
    placeholderData: (prev) => prev,
  });
}

export function useLeave(id: number) {
  return useQuery({
    queryKey: ["leaves", id],
    queryFn: () => leaveService.getById(id),
    enabled: !!id,
  });
}

export function useBatchApproveLeaves() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: () => leaveService.batchApprove(),
    onSuccess: (data) => {
      toast.success(`Đã duyệt ${data.approved_count} yêu cầu nghỉ phép`);
      qc.invalidateQueries({ queryKey: ["leaves"] });
      qc.invalidateQueries({ queryKey: ["approvals"] });
    },
    onError: () => {
      toast.error("Duyệt hàng loạt thất bại");
    },
  });
}

export function useProcessLeave() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: number; req: ProcessLeaveRequest }) =>
      leaveService.process(id, req),
    onSuccess: (data) => {
      const action = data.status === "approved" ? "duyệt" : "từ chối";
      toast.success(`Đã ${action} yêu cầu nghỉ phép`);
      qc.invalidateQueries({ queryKey: ["leaves"] });
      qc.invalidateQueries({ queryKey: ["approvals"] });
    },
    onError: () => {
      toast.error("Xử lý yêu cầu thất bại");
    },
  });
}

export function usePendingApprovals(page = 1, limit = 20) {
  return useQuery({
    queryKey: ["approvals", "pending", page, limit],
    queryFn: () => leaveService.getPendingApprovals(page, limit),
    placeholderData: (prev) => prev,
  });
}
