import { User } from "./user";

export type OvertimeStatus = "pending" | "approved" | "rejected";

export interface OvertimeRequest {
  id: number;
  user_id: number;
  user?: User;
  branch_id: number;
  date: string;
  actual_checkin: string | null;
  actual_checkout: string | null;
  calculated_start: string | null;
  calculated_end: string | null;
  total_hours: number;
  status: OvertimeStatus;
  manager_id: number | null;
  processed_by?: User;
  processed_at: string | null;
  manager_note: string;
  created_at: string;
  updated_at: string;
}

export interface OvertimeFilter {
  status?: OvertimeStatus;
  page?: number;
  limit?: number;
}

export interface ProcessOvertimeRequest {
  status: "approved" | "rejected";
  manager_note?: string;
}
