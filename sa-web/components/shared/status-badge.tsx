import { Badge } from "@/components/ui/badge";
import { AttendanceStatus } from "@/types/attendance";

const STATUS_CONFIG: Record<
  AttendanceStatus,
  { label: string; variant: "success" | "warning" | "destructive" | "secondary" | "info" }
> = {
  present: { label: "Đúng giờ", variant: "success" },
  late: { label: "Đi trễ", variant: "warning" },
  early_leave: { label: "Về sớm", variant: "info" },
  absent: { label: "Vắng mặt", variant: "destructive" },
  half_day: { label: "Nửa ngày", variant: "secondary" },
};

export function StatusBadge({ status }: { status: AttendanceStatus }) {
  const config = STATUS_CONFIG[status];
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
