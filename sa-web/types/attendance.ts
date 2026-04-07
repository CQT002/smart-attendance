import { User } from "./user";
import { Branch } from "./branch";

export type AttendanceStatus =
  | "present"
  | "late"
  | "early_leave"
  | "absent"
  | "half_day"
  | "leave"
  | "half_day_leave";

export type CheckMethod = "wifi" | "gps";

export interface AttendanceLog {
  id: number;
  user_id: number;
  user?: User;
  branch_id: number;
  branch?: Branch;
  shift_id: number | null;
  date: string;
  check_in_time: string | null;
  check_in_lat: number | null;
  check_in_lng: number | null;
  check_in_method: CheckMethod | null;
  check_in_ssid: string;
  check_in_bssid: string;
  check_out_time: string | null;
  check_out_lat: number | null;
  check_out_lng: number | null;
  check_out_method: CheckMethod | null;
  check_out_ssid: string;
  check_out_bssid: string;
  device_id: string;
  device_model: string;
  ip_address: string;
  app_version: string;
  is_fake_gps: boolean;
  is_vpn: boolean;
  fraud_note: string;
  status: AttendanceStatus;
  work_hours: number;
  overtime: number;
  note: string;
  created_at: string;
  updated_at: string;
}

export interface AttendanceSummary {
  user_id: number;
  user?: User;
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
}

export interface AttendanceFilter {
  user_id?: number;
  branch_id?: number;
  department?: string;
  status?: AttendanceStatus;
  search?: string;
  date_from?: string;
  date_to?: string;
  page?: number;
  limit?: number;
}
