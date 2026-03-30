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
import { Loader2 } from "lucide-react";
import { Branch, CreateBranchRequest, UpdateBranchRequest } from "@/types/branch";
import { useEffect } from "react";

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
  gps_radius: z.coerce.number().min(50).max(5000).optional().or(z.literal("")),
});

type FormData = z.infer<typeof schema>;

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

  useEffect(() => {
    if (open) {
      reset(defaultValues ?? {});
    }
  }, [open, defaultValues, reset]);

  const handleFormSubmit = async (data: FormData) => {
    const payload = {
      ...data,
      latitude: data.latitude === "" ? undefined : Number(data.latitude),
      longitude: data.longitude === "" ? undefined : Number(data.longitude),
      gps_radius: data.gps_radius === "" ? undefined : Number(data.gps_radius),
    };
    await onSubmit(payload as CreateBranchRequest);
  };

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
              <Label>Thành phố</Label>
              <Input placeholder="Hà Nội" {...register("city")} />
            </div>
            <div className="space-y-2">
              <Label>Tỉnh/Thành</Label>
              <Input placeholder="Hà Nội" {...register("province")} />
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

          {/* GPS Geofencing */}
          <div className="rounded-lg border p-4 space-y-3">
            <h3 className="font-medium text-sm flex items-center gap-2">
              <span className="text-green-600">📍</span> Cấu hình GPS Geofencing
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Vĩ độ (Latitude)</Label>
                <Input
                  type="number"
                  step="any"
                  placeholder="21.028511"
                  {...register("latitude")}
                />
              </div>
              <div className="space-y-2">
                <Label>Kinh độ (Longitude)</Label>
                <Input
                  type="number"
                  step="any"
                  placeholder="105.834160"
                  {...register("longitude")}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Bán kính (mét)</Label>
              <Input
                type="number"
                placeholder="100"
                {...register("gps_radius")}
              />
              <p className="text-xs text-muted-foreground">
                Nhân viên phải ở trong bán kính này để check-in bằng GPS (50–5000m)
              </p>
            </div>
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
