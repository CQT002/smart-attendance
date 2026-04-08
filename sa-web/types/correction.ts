import { User } from "./user";
import { AttendanceLog, AttendanceStatus } from "./attendance";
import { OvertimeRequest } from "./overtime";

export type CorrectionStatus = "pending" | "approved" | "rejected";
export type CorrectionType = "attendance" | "overtime";

export interface AttendanceCorrection {
  id: number;
  correction_type: CorrectionType;
  user_id: number;
  user?: User;
  branch_id: number;
  attendance_log_id: number | null;
  attendance_log?: AttendanceLog;
  overtime_request_id: number | null;
  overtime_request?: OvertimeRequest;
  original_status: AttendanceStatus | "missing_checkin" | "missing_checkout";
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
  correction_type?: CorrectionType;
  page?: number;
  limit?: number;
}

export interface ProcessCorrectionRequest {
  status: "approved" | "rejected";
  manager_note?: string;
}
