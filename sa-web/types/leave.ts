import { User } from "./user";

export type LeaveStatus = "pending" | "approved" | "rejected";
export type LeaveType = "full_day" | "half_day_morning" | "half_day_afternoon";

export interface LeaveRequest {
  id: number;
  user_id: number;
  user?: User;
  branch_id: number;
  leave_date: string;
  leave_type: LeaveType;
  time_from: string;
  time_to: string;
  original_status: string;
  description: string;
  status: LeaveStatus;
  processed_by_id: number | null;
  processed_by?: User;
  processed_at: string | null;
  manager_note: string;
  created_at: string;
  updated_at: string;
}

export interface LeaveFilter {
  status?: LeaveStatus;
  page?: number;
  limit?: number;
}

export interface ProcessLeaveRequest {
  status: "approved" | "rejected";
  manager_note?: string;
}

export interface PendingApprovalItem {
  id: number;
  type: "correction" | "leave" | "overtime";
  user_id: number;
  user_name: string;
  employee_code: string;
  department: string;
  branch_id: number;
  date: string;
  description: string;
  detail: string;
  status: string;
  created_at: string;
  // Audit fields
  processed_by_name?: string;
  processed_at?: string;
  manager_note?: string;
  // Correction-specific
  check_in_time?: string;
  check_out_time?: string;
  // Leave-specific
  leave_type?: string;
  time_from?: string;
  time_to?: string;
  // Overtime-specific
  actual_checkin?: string;
  actual_checkout?: string;
  calculated_start?: string;
  calculated_end?: string;
  total_hours?: number;
}
