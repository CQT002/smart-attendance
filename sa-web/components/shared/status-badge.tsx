import { Badge } from "@/components/ui/badge";
import { AttendanceStatus } from "@/types/attendance";

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
};

export function StatusBadge({ status }: { status: string }) {
  const config = STATUS_CONFIG[status] ?? { label: status, variant: "secondary" as const };
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
