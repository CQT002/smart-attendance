"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Branch } from "@/types/branch";
import { useWifiConfigs } from "@/hooks/use-wifi-configs";
import { useShifts } from "@/hooks/use-shifts";
import { DAY_NAMES } from "@/services/shift.service";
import { MapPin, Wifi, WifiOff, Building2, Phone, Mail, Clock } from "lucide-react";
import { formatDate } from "@/lib/utils";

interface BranchDetailDialogProps {
  branch: Branch;
  open: boolean;
  onClose: () => void;
}

export function BranchDetailDialog({ branch, open, onClose }: BranchDetailDialogProps) {
  const { data: wifiConfigs, isLoading: wifiLoading } = useWifiConfigs(branch.id);
  const { data: shifts, isLoading: shiftLoading } = useShifts(branch.id);
  const defaultShift = shifts?.find((s) => s.is_default) ?? shifts?.[0];

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Chi tiết chi nhánh
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-5">
          {/* Basic Info */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold">{branch.name}</h3>
                <p className="text-sm text-muted-foreground">{branch.code}</p>
              </div>
              <Badge variant={branch.is_active ? "success" : "secondary"}>
                {branch.is_active ? "Hoạt động" : "Ngưng"}
              </Badge>
            </div>

            <div className="grid grid-cols-1 gap-2 text-sm">
              <div className="flex items-start gap-2">
                <MapPin className="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
                <div>
                  <p>{branch.address}</p>
                  {(branch.city || branch.province) && (
                    <p className="text-muted-foreground">
                      {[branch.city, branch.province].filter(Boolean).join(", ")}
                    </p>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Phone className="h-4 w-4 text-muted-foreground" />
                <span>{branch.phone}</span>
              </div>
              <div className="flex items-center gap-2">
                <Mail className="h-4 w-4 text-muted-foreground" />
                <span>{branch.email}</span>
              </div>
            </div>
          </div>

          {/* GPS */}
          <div className="rounded-lg border p-3 space-y-2">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <span className="text-green-600">📍</span> Toạ độ (GPS)
            </h4>
            {branch.latitude && branch.longitude ? (
              <div className="text-sm space-y-1 overflow-x-auto">
                <div className="flex items-center justify-between gap-4 whitespace-nowrap">
                  <div>
                    <span className="text-muted-foreground">Vĩ độ: </span>
                    <span className="font-mono">{branch.latitude}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Kinh độ: </span>
                    <span className="font-mono">{branch.longitude}</span>
                  </div>
                  {branch.gps_radius && (
                    <div>
                      <span className="text-muted-foreground">Bán kính: </span>
                      <span className="font-mono">{branch.gps_radius}m</span>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">Chưa cấu hình</p>
            )}
          </div>

          {/* WiFi */}
          <div className="rounded-lg border p-3 space-y-2">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <Wifi className="h-4 w-4 text-blue-500" /> Mạng WiFi
            </h4>
            {wifiLoading ? (
              <p className="text-sm text-muted-foreground">Đang tải...</p>
            ) : wifiConfigs && wifiConfigs.length > 0 ? (
              <div className="space-y-2">
                {wifiConfigs.map((w) => (
                  <div key={w.id} className="flex items-center justify-between rounded-md bg-muted/50 px-3 py-2 text-sm">
                    <div>
                      <span className="font-medium">{w.ssid}</span>
                      {w.bssid && (
                        <span className="text-muted-foreground font-mono ml-2 text-xs">{w.bssid}</span>
                      )}
                    </div>
                    <Badge variant={w.is_active ? "success" : "secondary"} className="text-xs">
                      {w.is_active ? "Bật" : "Tắt"}
                    </Badge>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <WifiOff className="h-4 w-4" />
                Chưa có mạng WiFi nào
              </div>
            )}
          </div>

          {/* Shift */}
          <div className="rounded-lg border p-3 space-y-2">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <Clock className="h-4 w-4 text-blue-500" /> Ca làm việc
            </h4>
            {shiftLoading ? (
              <p className="text-sm text-muted-foreground">Đang tải...</p>
            ) : defaultShift ? (
              <div className="text-sm space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-medium">{defaultShift.name}</span>
                  <Badge variant={defaultShift.is_active ? "success" : "secondary"} className="text-xs">
                    {defaultShift.is_active ? "Bật" : "Tắt"}
                  </Badge>
                </div>
                <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Giờ làm:</span>
                    <span className="font-mono">{defaultShift.start_time} – {defaultShift.end_time}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Số giờ chuẩn:</span>
                    <span>{defaultShift.work_hours}h</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Nghỉ trưa:</span>
                    <span className="font-mono">{defaultShift.morning_end} – {defaultShift.afternoon_start}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Trễ / Sớm:</span>
                    <span>{defaultShift.late_after}p / {defaultShift.early_before}p</span>
                  </div>
                </div>
                <div className="rounded-md bg-blue-50 dark:bg-blue-950/30 px-3 py-2 text-xs space-y-1">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Khung chính thức:</span>
                    <span className="font-medium">T2 → {DAY_NAMES[defaultShift.regular_end_day]} {defaultShift.regular_end_time}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Tăng ca (OT):</span>
                    <span className="font-medium">{defaultShift.ot_start_hour}:00 – {defaultShift.ot_end_hour}:00 (check-in từ {defaultShift.ot_min_checkin_hour}:00)</span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Clock className="h-4 w-4" />
                Chưa có ca làm việc
              </div>
            )}
          </div>

          {/* Metadata */}
          <div className="text-xs text-muted-foreground border-t pt-3 flex gap-4">
            <span>Ngày tạo: {formatDate(branch.created_at)}</span>
            <span>Cập nhật: {formatDate(branch.updated_at)}</span>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
