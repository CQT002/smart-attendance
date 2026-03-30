"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { branchService } from "@/services/branch.service";
import { BranchFilter, CreateBranchRequest, UpdateBranchRequest } from "@/types/branch";
import { toast } from "sonner";

export function useBranches(filter: BranchFilter = {}) {
  return useQuery({
    queryKey: ["branches", filter],
    queryFn: () => branchService.getList(filter),
    placeholderData: (prev) => prev,
  });
}

export function useActiveBranches() {
  return useQuery({
    queryKey: ["branches", "active"],
    queryFn: () => branchService.getActive(),
    staleTime: 10 * 60 * 1000,
  });
}

export function useBranch(id: number) {
  return useQuery({
    queryKey: ["branches", id],
    queryFn: () => branchService.getById(id),
    enabled: !!id,
  });
}

export function useCreateBranch() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateBranchRequest) => branchService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["branches"] });
      toast.success("Tạo chi nhánh thành công");
    },
    onError: () => toast.error("Tạo chi nhánh thất bại"),
  });
}

export function useUpdateBranch() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateBranchRequest }) =>
      branchService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["branches"] });
      toast.success("Cập nhật chi nhánh thành công");
    },
    onError: () => toast.error("Cập nhật thất bại"),
  });
}

export function useDeleteBranch() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => branchService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["branches"] });
      toast.success("Vô hiệu hoá chi nhánh thành công");
    },
    onError: () => toast.error("Thao tác thất bại"),
  });
}
