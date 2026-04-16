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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Loader2 } from "lucide-react";
import { useEffect } from "react";
import {
  CreateHolidayRequest,
  Holiday,
  HolidayType,
  UpdateHolidayRequest,
} from "@/types/holiday";

const schema = z.object({
  name: z.string().min(1, "Bắt buộc"),
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Định dạng YYYY-MM-DD"),
  coefficient: z.coerce.number().min(0).max(10).optional().or(z.literal("")),
  type: z.enum(["national", "company"]),
  is_compensated: z.boolean().optional(),
  compensate_for: z.string().optional(),
  description: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

interface HolidayFormDialogProps {
  open: boolean;
  defaultValues?: Holiday;
  onClose: () => void;
  onSubmit: (data: CreateHolidayRequest | UpdateHolidayRequest) => Promise<void>;
  loading?: boolean;
}

export function HolidayFormDialog({
  open,
  defaultValues,
  onClose,
  onSubmit,
  loading,
}: HolidayFormDialogProps) {
  const isEdit = !!defaultValues;
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: "national",
      is_compensated: false,
    },
  });

  const isCompensated = watch("is_compensated");
  const currentType = watch("type");

  useEffect(() => {
    if (!open) return;
    if (defaultValues) {
      reset({
        name: defaultValues.name,
        date: defaultValues.date?.slice(0, 10) ?? "",
        coefficient: defaultValues.coefficient,
        type: defaultValues.type,
        is_compensated: defaultValues.is_compensated,
        compensate_for: defaultValues.compensate_for?.slice(0, 10) ?? "",
        description: defaultValues.description ?? "",
      });
    } else {
      reset({
        name: "",
        date: "",
        coefficient: undefined,
        type: "national",
        is_compensated: false,
        compensate_for: "",
        description: "",
      });
    }
  }, [open, defaultValues, reset]);

  const handleFormSubmit = async (data: FormData) => {
    const payload: CreateHolidayRequest = {
      name: data.name,
      date: data.date,
      coefficient:
        data.coefficient === "" || data.coefficient === undefined
          ? undefined
          : Number(data.coefficient),
      type: data.type as HolidayType,
      is_compensated: !!data.is_compensated,
      compensate_for:
        data.is_compensated && data.compensate_for
          ? data.compensate_for
          : undefined,
      description: data.description ?? "",
    };
    await onSubmit(payload);
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Chỉnh sửa ngày lễ" : "Thêm ngày lễ mới"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label>Tên ngày lễ *</Label>
            <Input placeholder="VD: Quốc khánh 2/9" {...register("name")} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Ngày *</Label>
              <Input type="date" {...register("date")} />
              {errors.date && <p className="text-xs text-destructive">{errors.date.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>Loại *</Label>
              <Select
                value={currentType}
                onValueChange={(v) => setValue("type", v as HolidayType)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Chọn loại" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="national">Lễ quốc gia</SelectItem>
                  <SelectItem value="company">Lễ công ty</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label>Hệ số lương</Label>
            <Input
              type="number"
              step="0.1"
              placeholder="Để trống để dùng mặc định (quốc gia=3.0, công ty=2.0)"
              {...register("coefficient")}
            />
            <p className="text-xs text-muted-foreground">
              Ví dụ: 2.0 = 200%, 3.0 = 300%, 4.0 = 400%
            </p>
            {errors.coefficient && (
              <p className="text-xs text-destructive">{String(errors.coefficient.message)}</p>
            )}
          </div>

          <div className="flex items-start gap-2 rounded-md border p-3">
            <input
              id="is_compensated"
              type="checkbox"
              className="mt-1"
              {...register("is_compensated")}
            />
            <div className="flex-1 space-y-1">
              <Label htmlFor="is_compensated" className="cursor-pointer">
                Ngày nghỉ bù
              </Label>
              <p className="text-xs text-muted-foreground">
                Khi lễ gốc rơi vào Thứ 7/Chủ nhật — tạo ngày nghỉ bù vào ngày làm việc kế tiếp
              </p>
              {isCompensated && (
                <div className="pt-2 space-y-1">
                  <Label className="text-xs">Ngày lễ gốc đang nghỉ bù cho *</Label>
                  <Input type="date" {...register("compensate_for")} />
                </div>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <Label>Mô tả</Label>
            <Input placeholder="Ghi chú thêm (tuỳ chọn)" {...register("description")} />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Huỷ
            </Button>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEdit ? "Lưu thay đổi" : "Tạo ngày lễ"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
