"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { holidayService } from "@/services/holiday.service";
import {
  CreateHolidayRequest,
  HolidayFilter,
  UpdateHolidayRequest,
} from "@/types/holiday";
import { toast } from "sonner";

export function useHolidays(filter: HolidayFilter = {}) {
  return useQuery({
    queryKey: ["holidays", filter],
    queryFn: () => holidayService.getList(filter),
    placeholderData: (prev) => prev,
  });
}

export function useHoliday(id: number) {
  return useQuery({
    queryKey: ["holidays", id],
    queryFn: () => holidayService.getById(id),
    enabled: !!id,
  });
}

export function useCreateHoliday() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateHolidayRequest) => holidayService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["holidays"] });
      toast.success("Tạo ngày lễ thành công");
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Tạo ngày lễ thất bại";
      toast.error(msg);
    },
  });
}

export function useUpdateHoliday() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateHolidayRequest }) =>
      holidayService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["holidays"] });
      toast.success("Cập nhật ngày lễ thành công");
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Cập nhật thất bại";
      toast.error(msg);
    },
  });
}

export function useDeleteHoliday() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => holidayService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["holidays"] });
      toast.success("Đã xoá ngày lễ");
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Xoá thất bại";
      toast.error(msg);
    },
  });
}
