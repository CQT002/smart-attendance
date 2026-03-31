import { UserRole } from "./auth";
import { Branch } from "./branch";

export interface User {
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
  branch?: Branch;
  role: UserRole;
  is_active: boolean;
}

export interface CreateUserRequest {
  branch_id?: number;
  employee_code: string;
  name: string;
  email: string;
  phone: string;
  role: UserRole;
  department: string;
  position: string;
}

export interface UpdateUserRequest {
  name: string;
  phone: string;
  department: string;
  position: string;
  avatar_url: string;
}

export interface UserFilter {
  branch_id?: number;
  role?: UserRole;
  department?: string;
  search?: string;
  is_active?: boolean;
  page?: number;
  limit?: number;
}
