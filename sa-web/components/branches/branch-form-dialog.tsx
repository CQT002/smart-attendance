"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Plus, Trash2, Wifi } from "lucide-react";
import { Branch, CreateBranchRequest, UpdateBranchRequest } from "@/types/branch";
import { useEffect, useState } from "react";
import {
  useWifiConfigs,
  useCreateWifiConfig,
  useDeleteWifiConfig,
} from "@/hooks/use-wifi-configs";

const schema = z.object({
  code: z.string().min(1, "Bắt buộc").optional().or(z.literal("")),
  name: z.string().min(1, "Bắt buộc"),
  address: z.string().min(1, "Bắt buộc"),
  city: z.string().optional(),
  province: z.string().optional(),
  phone: z.string().min(1, "Bắt buộc"),
  email: z.string().email("Email không hợp lệ"),
  latitude: z.coerce.number().optional().or(z.literal("")),
  longitude: z.coerce.number().optional().or(z.literal("")),
  gps_radius: z.coerce.number().min(100).max(5000).optional().or(z.literal("")),
});

type FormData = z.infer<typeof schema>;

interface WifiEntry {
  ssid: string;
  bssid: string;
  description: string;
}

interface BranchFormDialogProps {
  open: boolean;
  defaultValues?: Branch;
  onClose: () => void;
  onSubmit: (data: CreateBranchRequest | UpdateBranchRequest) => Promise<void>;
  loading?: boolean;
}

export function BranchFormDialog({
  open,
  defaultValues,
  onClose,
  onSubmit,
  loading,
}: BranchFormDialogProps) {
  const isEdit = !!defaultValues;
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  // WiFi state for create mode
  const [wifiEntries, setWifiEntries] = useState<WifiEntry[]>([]);
  const [newSsid, setNewSsid] = useState("");
  const [newBssid, setNewBssid] = useState("");
  const [newDesc, setNewDesc] = useState("");

  // WiFi hooks for edit mode
  const { data: existingWifi } = useWifiConfigs(isEdit ? defaultValues!.id : 0);
  const createWifi = useCreateWifiConfig(isEdit ? defaultValues!.id : 0);
  const deleteWifi = useDeleteWifiConfig(isEdit ? defaultValues!.id : 0);

  useEffect(() => {
    if (open) {
      reset(defaultValues ? {
        ...defaultValues,
        latitude: defaultValues.latitude ?? undefined,
        longitude: defaultValues.longitude ?? undefined,
        gps_radius: defaultValues.gps_radius ?? undefined,
      } : {});
      setWifiEntries([]);
      setNewSsid("");
      setNewBssid("");
      setNewDesc("");
    }
  }, [open, defaultValues, reset]);

  const handleAddWifi = () => {
    if (!newSsid.trim()) return;
    if (isEdit) {
      createWifi.mutate({
        ssid: newSsid.trim(),
        bssid: newBssid.trim() || undefined,
        description: newDesc.trim() || undefined,
      });
    } else {
      setWifiEntries((prev) => [
        ...prev,
        { ssid: newSsid.trim(), bssid: newBssid.trim(), description: newDesc.trim() },
      ]);
    }
    setNewSsid("");
    setNewBssid("");
    setNewDesc("");
  };

  const handleRemoveWifi = (index: number) => {
    setWifiEntries((prev) => prev.filter((_, i) => i !== index));
  };

  const handleDeleteExistingWifi = (id: number) => {
    if (confirm("Xác nhận xóa WiFi này?")) {
      deleteWifi.mutate(id);
    }
  };

  const handleFormSubmit = async (data: FormData) => {
    const payload = {
      ...data,
      latitude: data.latitude === "" ? undefined : Number(data.latitude),
      longitude: data.longitude === "" ? undefined : Number(data.longitude),
      gps_radius: data.gps_radius === "" ? undefined : Number(data.gps_radius),
      wifi_configs: isEdit ? undefined : wifiEntries.length > 0 ? wifiEntries : undefined,
    };
    await onSubmit(payload as CreateBranchRequest);
  };

  // Combine existing wifi (edit mode) and pending entries (create mode)
  const wifiList = isEdit ? existingWifi ?? [] : [];

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Chỉnh sửa chi nhánh" : "Thêm chi nhánh mới"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            {!isEdit && (
              <div className="space-y-2">
                <Label>Mã chi nhánh *</Label>
                <Input placeholder="CN001" {...register("code")} />
                {errors.code && <p className="text-xs text-destructive">{errors.code.message}</p>}
              </div>
            )}
            <div className={`space-y-2 ${!isEdit ? "" : "col-span-2"}`}>
              <Label>Tên chi nhánh *</Label>
              <Input placeholder="Chi nhánh Hà Nội" {...register("name")} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
          </div>

          <div className="space-y-2">
            <Label>Địa chỉ *</Label>
            <Input placeholder="Số 1, Phố Đinh Tiên Hoàng..." {...register("address")} />
            {errors.address && <p className="text-xs text-destructive">{errors.address.message}</p>}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Phường</Label>
              <Input placeholder="Tân Định" {...register("city")} />
            </div>
            <div className="space-y-2">
              <Label>Tỉnh/Thành phố</Label>
              <Input placeholder="TP Hồ Chí Minh" {...register("province")} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Số điện thoại *</Label>
              <Input placeholder="024 1234 5678" {...register("phone")} />
              {errors.phone && <p className="text-xs text-destructive">{errors.phone.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>Email *</Label>
              <Input placeholder="cn001@hdbank.com" {...register("email")} />
              {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
            </div>
          </div>

          {/* GPS */}
          <div className="rounded-lg border p-4 space-y-3">
            <h3 className="font-medium text-sm flex items-center gap-2">
              <span className="text-green-600">📍</span> Cấu hình toạ độ (GPS)
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Vĩ độ (Latitude)</Label>
                <Input type="number" step="any" placeholder="21.028511" {...register("latitude")} />
              </div>
              <div className="space-y-2">
                <Label>Kinh độ (Longitude)</Label>
                <Input type="number" step="any" placeholder="105.834160" {...register("longitude")} />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Bán kính (mét)</Label>
              <Input type="number" placeholder="100" {...register("gps_radius")} />
              <p className="text-xs text-muted-foreground">
                Nhân viên phải ở trong bán kính 100m để check-in bằng GPS
              </p>
            </div>
          </div>

          {/* WiFi Config */}
          <div className="rounded-lg border p-4 space-y-3">
            <h3 className="font-medium text-sm flex items-center gap-2">
              <Wifi className="h-4 w-4 text-blue-500" /> Cấu hình WiFi
            </h3>

            {/* Existing WiFi list (edit mode) */}
            {isEdit && wifiList.length > 0 && (
              <div className="space-y-2">
                {wifiList.map((w) => (
                  <div key={w.id} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                    <div>
                      <span className="font-medium">{w.ssid}</span>
                      {w.bssid && (
                        <span className="text-muted-foreground font-mono ml-2 text-xs">{w.bssid}</span>
                      )}
                      {w.description && (
                        <span className="text-muted-foreground ml-2 text-xs">— {w.description}</span>
                      )}
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-destructive hover:text-destructive"
                      onClick={() => handleDeleteExistingWifi(w.id)}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                ))}
              </div>
            )}

            {/* Pending WiFi list (create mode) */}
            {!isEdit && wifiEntries.length > 0 && (
              <div className="space-y-2">
                {wifiEntries.map((w, i) => (
                  <div key={i} className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                    <div>
                      <span className="font-medium">{w.ssid}</span>
                      {w.bssid && (
                        <span className="text-muted-foreground font-mono ml-2 text-xs">{w.bssid}</span>
                      )}
                      {w.description && (
                        <span className="text-muted-foreground ml-2 text-xs">— {w.description}</span>
                      )}
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-destructive hover:text-destructive"
                      onClick={() => handleRemoveWifi(i)}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                ))}
              </div>
            )}

            {/* Add WiFi form */}
            <div className="space-y-2">
              <div className="grid grid-cols-2 gap-2">
                <Input
                  placeholder="SSID (tên mạng) *"
                  value={newSsid}
                  onChange={(e) => setNewSsid(e.target.value)}
                  className="text-sm"
                />
                <Input
                  placeholder="BSSID (MAC address)"
                  value={newBssid}
                  onChange={(e) => setNewBssid(e.target.value)}
                  className="text-sm"
                />
              </div>
              <div className="flex gap-2">
                <Input
                  placeholder="Mô tả (tuỳ chọn)"
                  value={newDesc}
                  onChange={(e) => setNewDesc(e.target.value)}
                  className="text-sm"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAddWifi}
                  disabled={!newSsid.trim()}
                  className="shrink-0"
                >
                  <Plus className="h-4 w-4 mr-1" /> Thêm
                </Button>
              </div>
            </div>
            <p className="text-xs text-muted-foreground">
              Thêm mạng WiFi của chi nhánh để nhân viên check-in bằng WiFi
            </p>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Huỷ
            </Button>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEdit ? "Lưu thay đổi" : "Tạo chi nhánh"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
