import apiClient from "@/lib/api-client";
import { setStoredUser, clearStoredUser } from "@/lib/auth";
import { ApiResponse } from "@/types/api";
import { LoginRequest, LoginResponse, ChangePasswordRequest } from "@/types/auth";
import { User } from "@/types/user";
import { AuthUser } from "@/types/auth";
import Cookies from "js-cookie";

export const authService = {
  async login(data: LoginRequest): Promise<LoginResponse> {
    const res = await apiClient.post<ApiResponse<LoginResponse>>(
      "/admin/auth/login",
      data
    );
    const result = res.data.data!;
    Cookies.set("access_token", result.access_token, { expires: 1 });
    Cookies.set("refresh_token", result.refresh_token, { expires: 7 });
    setStoredUser(result.user);
    return result;
  },

  async getMe(): Promise<User> {
    const res = await apiClient.get<ApiResponse<User>>("/admin/auth/me");
    const user = res.data.data!;
    setStoredUser(user);
    return user;
  },

  async changePassword(data: ChangePasswordRequest): Promise<void> {
    await apiClient.put("/admin/auth/change-password", data);
  },

  logout(): void {
    clearStoredUser();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
  },
};
