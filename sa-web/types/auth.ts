export type UserRole = "admin" | "manager" | "employee";

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: AuthUser;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

/**
 * Lightweight user shape returned by auth endpoints.
 * For the full User type (with relations), import from "./user".
 */
export interface AuthUser {
  id: number;
  employee_code: string;
  name: string;
  email: string;
  phone: string;
  department: string;
  position: string;
  avatar_url: string;
  hired_at: string | null;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
  branch_id: number | null;
  role: UserRole;
  is_active: boolean;
}
