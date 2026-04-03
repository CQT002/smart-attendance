import { User } from "./user";
import { AttendanceSummary } from "./attendance";

export type ReportPeriod = "daily" | "weekly" | "monthly" | "custom";

export interface DashboardStats {
  total_branches: number;
  total_employees: number;
  present_today: number;
  absent_today: number;
  late_today: number;
  attendance_rate: number;
  on_time_rate: number;
}

export interface BranchTodayStats {
  branch_id: number;
  branch_name: string;
  branch_code: string;
  date: string;
  total_employees: number;
  present_count: number;
  late_count: number;
  early_leave_count: number;
  half_day_count: number;
  absent_count: number;
  total_work_hours: number;
  total_overtime: number;
  fraud_count: number;
  attendance_rate: number;
  on_time_rate: number;
}

export interface EmployeeTodayDetail {
  user_id: number;
  user?: User;
  branch_id: number;
  branch_name: string;
  status: string;
  check_in_time: string | null;
  check_out_time: string | null;
  work_hours: number;
  is_fake_gps: boolean;
  is_vpn: boolean;
}

export interface UserAttendanceSummary extends AttendanceSummary {
  user_name: string;
  employee_code: string;
  department: string;
  date_from: string;
  date_to: string;
}

export interface BranchAttendanceReport {
  branch_id: number;
  branch_name: string;
  branch_code: string;
  total_days: number;
  present_count: number;
  late_count: number;
  early_leave_count: number;
  half_day_count: number;
  absent_count: number;
  total_work_hours: number;
  total_overtime: number;
  attendance_rate: number;
  on_time_rate: number;
  employees?: UserAttendanceSummary[];
}

export interface ReportFilter {
  branch_id?: number;
  user_id?: number;
  department?: string;
  period?: ReportPeriod;
  date_from?: string;
  date_to?: string;
  page?: number;
  limit?: number;
}

export interface TodayStatsFilter {
  branch_id?: number;
  search?: string;
  page?: number;
  limit?: number;
}
