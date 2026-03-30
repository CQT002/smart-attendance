import Cookies from "js-cookie";
import { User } from "@/types/user";

export function getStoredUser(): User | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem("current_user");
    return raw ? (JSON.parse(raw) as User) : null;
  } catch {
    return null;
  }
}

export function setStoredUser(user: User): void {
  if (typeof window === "undefined") return;
  localStorage.setItem("current_user", JSON.stringify(user));
}

export function clearStoredUser(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem("current_user");
  Cookies.remove("access_token");
  Cookies.remove("refresh_token");
}

export function isAuthenticated(): boolean {
  return !!Cookies.get("access_token");
}
