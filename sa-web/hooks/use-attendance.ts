"use client";

import { useQuery } from "@tanstack/react-query";
import { attendanceService } from "@/services/attendance.service";
import { AttendanceFilter } from "@/types/attendance";

export function useAttendanceLogs(filter: AttendanceFilter = {}) {
  return useQuery({
    queryKey: ["attendance", filter],
    queryFn: () => attendanceService.getList(filter),
    placeholderData: (prev) => prev,
  });
}

export function useAttendanceLog(id: number) {
  return useQuery({
    queryKey: ["attendance", id],
    queryFn: () => attendanceService.getById(id),
    enabled: !!id,
  });
}

export function useAttendanceSummary(userId: number, from: string, to: string) {
  return useQuery({
    queryKey: ["attendance-summary", userId, from, to],
    queryFn: () => attendanceService.getSummary(userId, from, to),
    enabled: !!userId && !!from && !!to,
  });
}
