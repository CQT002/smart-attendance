"use client";

import { useQuery } from "@tanstack/react-query";
import { reportService } from "@/services/report.service";
import { ReportFilter, TodayStatsFilter } from "@/types/report";

export function useDashboardStats() {
  return useQuery({
    queryKey: ["dashboard-stats"],
    queryFn: () => reportService.getDashboardStats(),
    refetchInterval: 5 * 60 * 1000,
  });
}

export function useTodayBranchStats(filter: TodayStatsFilter = {}) {
  return useQuery({
    queryKey: ["today-branch-stats", filter],
    queryFn: () => reportService.getTodayBranchStats(filter),
    placeholderData: (prev) => prev,
    refetchInterval: 5 * 60 * 1000,
  });
}

export function useTodayEmployees(filter: {
  branch_id?: number;
  status?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery({
    queryKey: ["today-employees", filter],
    queryFn: () => reportService.getTodayEmployees(filter),
    placeholderData: (prev) => prev,
  });
}

export function useAttendanceReport(filter: ReportFilter = {}, enabled = true) {
  return useQuery({
    queryKey: ["attendance-report", filter],
    queryFn: () => reportService.getAttendanceReport(filter),
    placeholderData: (prev) => prev,
    enabled,
  });
}

export function useBranchReport(filter: ReportFilter = {}, enabled = true) {
  return useQuery({
    queryKey: ["branch-report", filter],
    queryFn: () => reportService.getBranchReport(filter),
    placeholderData: (prev) => prev,
    enabled,
  });
}
