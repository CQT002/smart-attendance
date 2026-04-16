import { Badge } from "@/components/ui/badge";

const STATUS_CONFIG: Record<
  string,
  { label: string; variant: "success" | "warning" | "destructive" | "secondary" | "info" }
> = {
  present: { label: "Đúng giờ", variant: "success" },
  late: { label: "Đi trễ - Về sớm", variant: "warning" },
  early_leave: { label: "Đi trễ - Về sớm", variant: "warning" },
  late_early_leave: { label: "Đi trễ - Về sớm", variant: "warning" },
  half_day: { label: "Đi trễ - Về sớm", variant: "warning" },
  absent: { label: "Vắng mặt", variant: "destructive" },
  leave: { label: "Nghỉ phép", variant: "info" },
  half_day_leave: { label: "Nghỉ phép nửa ngày", variant: "info" },
  missing_checkout: { label: "Thiếu check-out", variant: "destructive" },
  missing_checkin: { label: "Thiếu check-in", variant: "destructive" },
};

interface StatusBadgeProps {
  status: string;
  checkInTime?: string | null;
  checkOutTime?: string | null;
}

/**
 * Hiển thị trạng thái chấm công. Nếu truyền checkInTime/checkOutTime,
 * sẽ override status khi phát hiện thiếu check-in hoặc check-out.
 */
export function StatusBadge({ status, checkInTime, checkOutTime }: StatusBadgeProps) {
  let effectiveStatus = status;

  // Override nếu thiếu check-in/out (status DB vẫn là present/late nhưng thực tế thiếu)
  if (checkInTime !== undefined && checkOutTime !== undefined) {
    const hasIn = !!checkInTime;
    const hasOut = !!checkOutTime;
    if (hasIn && !hasOut && !["leave", "half_day_leave"].includes(status)) {
      effectiveStatus = "missing_checkout";
    } else if (!hasIn && hasOut && !["leave", "half_day_leave"].includes(status)) {
      effectiveStatus = "missing_checkin";
    }
  }

  const config = STATUS_CONFIG[effectiveStatus] ?? { label: effectiveStatus, variant: "secondary" as const };
  return <Badge variant={config.variant}>{config.label}</Badge>;
}

export function ActiveBadge({ isActive }: { isActive: boolean }) {
  return (
    <Badge variant={isActive ? "success" : "secondary"}>
      {isActive ? "Hoạt động" : "Vô hiệu"}
    </Badge>
  );
}

export function RoleBadge({ role }: { role: string }) {
  const map: Record<string, { label: string; variant: "default" | "secondary" | "info" }> = {
    admin: { label: "Admin", variant: "default" },
    manager: { label: "Quản lý", variant: "info" },
    employee: { label: "Nhân viên", variant: "secondary" },
  };
  const config = map[role] ?? { label: role, variant: "secondary" as const };
  return <Badge variant={config.variant}>{config.label}</Badge>;
}
