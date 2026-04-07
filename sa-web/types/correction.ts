import { User } from "./user";
import { AttendanceLog, AttendanceStatus } from "./attendance";

export type CorrectionStatus = "pending" | "approved" | "rejected";

export interface AttendanceCorrection {
  id: number;
  user_id: number;
  user?: User;
  branch_id: number;
  attendance_log_id: number;
  attendance_log?: AttendanceLog;
  original_status: AttendanceStatus;
  credit_count: number;
  description: string;
  status: CorrectionStatus;
  processed_by_id: number | null;
  processed_by?: User;
  processed_at: string | null;
  manager_note: string;
  created_at: string;
  updated_at: string;
}

export interface CorrectionFilter {
  status?: CorrectionStatus;
  page?: number;
  limit?: number;
}

export interface ProcessCorrectionRequest {
  status: "approved" | "rejected";
  manager_note?: string;
}
