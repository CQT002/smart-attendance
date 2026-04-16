export type HolidayType = "national" | "company";

export interface Holiday {
  id: number;
  name: string;
  date: string; // YYYY-MM-DD
  year: number;
  coefficient: number;
  type: HolidayType;
  is_compensated: boolean;
  compensate_for?: string | null;
  description?: string;
  created_by_id?: number | null;
  created_at: string;
  updated_at: string;
}

export interface CreateHolidayRequest {
  name: string;
  date: string; // YYYY-MM-DD
  coefficient?: number; // empty → default theo type
  type?: HolidayType;
  is_compensated?: boolean;
  compensate_for?: string;
  description?: string;
}

export interface UpdateHolidayRequest {
  name: string;
  date: string;
  coefficient?: number;
  type?: HolidayType;
  is_compensated?: boolean;
  compensate_for?: string;
  description?: string;
}

export interface HolidayFilter {
  year?: number;
  type?: HolidayType;
  page?: number;
  limit?: number;
}
