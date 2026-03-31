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
import { MapPin, Wifi, WifiOff, Building2, Phone, Mail } from "lucide-react";
import { formatDate } from "@/lib/utils";

interface BranchDetailDialogProps {
  branch: Branch;
  open: boolean;
  onClose: () => void;
}

export function BranchDetailDialog({ branch, open, onClose }: BranchDetailDialogProps) {
  const { data: wifiConfigs, isLoading: wifiLoading } = useWifiConfigs(branch.id);

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
              <div className="text-sm space-y-1">
                <div className="grid grid-cols-3 gap-2">
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
