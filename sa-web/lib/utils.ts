import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { format, parseISO } from "date-fns";
import { toZonedTime } from "date-fns-tz";
import { vi } from "date-fns/locale";

const TZ_HCM = "Asia/Ho_Chi_Minh";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(dateStr: string | null | undefined, fmt = "dd/MM/yyyy"): string {
  if (!dateStr) return "—";
  try {
    const zonedDate = toZonedTime(parseISO(dateStr), TZ_HCM);
    return format(zonedDate, fmt, { locale: vi });
  } catch {
    return "—";
  }
}

export function formatDateTime(dateStr: string | null | undefined): string {
  return formatDate(dateStr, "HH:mm dd/MM/yyyy");
}

export function formatTime(dateStr: string | null | undefined): string {
  return formatDate(dateStr, "HH:mm");
}

export function formatPercent(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

export function formatHours(hours: number): string {
  const h = Math.floor(hours);
  const m = Math.round((hours - h) * 60);
  if (m === 0) return `${h}h`;
  return `${h}h${m}m`;
}
