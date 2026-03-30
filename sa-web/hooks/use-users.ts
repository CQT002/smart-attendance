"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { userService } from "@/services/user.service";
import { UserFilter, CreateUserRequest, UpdateUserRequest } from "@/types/user";
import { toast } from "sonner";

export function useUsers(filter: UserFilter = {}) {
  return useQuery({
    queryKey: ["users", filter],
    queryFn: () => userService.getList(filter),
    placeholderData: (prev) => prev,
  });
}

export function useUser(id: number) {
  return useQuery({
    queryKey: ["users", id],
    queryFn: () => userService.getById(id),
    enabled: !!id,
  });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateUserRequest) => userService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["users"] });
      toast.success("Tạo nhân viên thành công");
    },
    onError: () => toast.error("Tạo nhân viên thất bại"),
  });
}

export function useUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateUserRequest }) =>
      userService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["users"] });
      toast.success("Cập nhật nhân viên thành công");
    },
    onError: () => toast.error("Cập nhật thất bại"),
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => userService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["users"] });
      toast.success("Vô hiệu hoá tài khoản thành công");
    },
    onError: () => toast.error("Thao tác thất bại"),
  });
}

export function useResetPassword() {
  return useMutation({
    mutationFn: ({ id, password }: { id: number; password: string }) =>
      userService.resetPassword(id, password),
    onSuccess: () => toast.success("Reset mật khẩu thành công"),
    onError: () => toast.error("Reset mật khẩu thất bại"),
  });
}
