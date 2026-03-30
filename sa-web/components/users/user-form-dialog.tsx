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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Loader2 } from "lucide-react";
import { User, CreateUserRequest, UpdateUserRequest } from "@/types/user";
import { Branch } from "@/types/branch";
import { useEffect } from "react";

const createSchema = z.object({
  employee_code: z.string().min(1, "Bắt buộc"),
  name: z.string().min(1, "Bắt buộc"),
  email: z.string().email("Email không hợp lệ"),
  phone: z.string().min(1, "Bắt buộc"),
  password: z.string().min(8, "Tối thiểu 8 ký tự"),
  role: z.enum(["admin", "manager", "employee"]),
  branch_id: z.coerce.number().optional(),
  department: z.string().optional(),
  position: z.string().optional(),
});

const updateSchema = z.object({
  name: z.string().min(1, "Bắt buộc"),
  phone: z.string().min(1, "Bắt buộc"),
  department: z.string().optional(),
  position: z.string().optional(),
  avatar_url: z.string().optional(),
});

type CreateFormData = z.infer<typeof createSchema>;
type UpdateFormData = z.infer<typeof updateSchema>;

interface UserFormDialogProps {
  open: boolean;
  defaultValues?: User;
  branches: Branch[];
  onClose: () => void;
  onSubmit: (data: CreateUserRequest | UpdateUserRequest) => Promise<void>;
  loading?: boolean;
}

export function UserFormDialog({
  open,
  defaultValues,
  branches,
  onClose,
  onSubmit,
  loading,
}: UserFormDialogProps) {
  const isEdit = !!defaultValues;

  const createForm = useForm<CreateFormData>({ resolver: zodResolver(createSchema) });
  const updateForm = useForm<UpdateFormData>({ resolver: zodResolver(updateSchema) });

  useEffect(() => {
    if (open && defaultValues) {
      updateForm.reset({
        name: defaultValues.name,
        phone: defaultValues.phone,
        department: defaultValues.department,
        position: defaultValues.position,
        avatar_url: defaultValues.avatar_url,
      });
    } else if (open) {
      createForm.reset({});
    }
  }, [open, defaultValues]);

  if (isEdit) {
    return (
      <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Chỉnh sửa nhân viên</DialogTitle>
          </DialogHeader>
          <form
            onSubmit={updateForm.handleSubmit((data) => onSubmit(data as UpdateUserRequest))}
            className="space-y-4"
          >
            <div className="space-y-2">
              <Label>Họ tên *</Label>
              <Input {...updateForm.register("name")} />
              {updateForm.formState.errors.name && (
                <p className="text-xs text-destructive">{updateForm.formState.errors.name.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label>Số điện thoại *</Label>
              <Input {...updateForm.register("phone")} />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Phòng ban</Label>
                <Input placeholder="Công nghệ" {...updateForm.register("department")} />
              </div>
              <div className="space-y-2">
                <Label>Chức vụ</Label>
                <Input placeholder="Developer" {...updateForm.register("position")} />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={onClose}>Huỷ</Button>
              <Button type="submit" disabled={loading}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Lưu thay đổi
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Thêm nhân viên mới</DialogTitle>
        </DialogHeader>
        <form
          onSubmit={createForm.handleSubmit((data) => onSubmit(data as CreateUserRequest))}
          className="space-y-4"
        >
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Mã nhân viên *</Label>
              <Input placeholder="NV001" {...createForm.register("employee_code")} />
              {createForm.formState.errors.employee_code && (
                <p className="text-xs text-destructive">{createForm.formState.errors.employee_code.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label>Vai trò *</Label>
              <Select
                defaultValue="employee"
                onValueChange={(v) => createForm.setValue("role", v as any)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="manager">Quản lý</SelectItem>
                  <SelectItem value="employee">Nhân viên</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label>Họ tên *</Label>
            <Input placeholder="Nguyễn Văn A" {...createForm.register("name")} />
            {createForm.formState.errors.name && (
              <p className="text-xs text-destructive">{createForm.formState.errors.name.message}</p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Email *</Label>
              <Input type="email" placeholder="nva@hdbank.com" {...createForm.register("email")} />
              {createForm.formState.errors.email && (
                <p className="text-xs text-destructive">{createForm.formState.errors.email.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label>Số điện thoại *</Label>
              <Input placeholder="0901234567" {...createForm.register("phone")} />
            </div>
          </div>

          <div className="space-y-2">
            <Label>Mật khẩu *</Label>
            <Input type="password" placeholder="••••••••" {...createForm.register("password")} />
            {createForm.formState.errors.password && (
              <p className="text-xs text-destructive">{createForm.formState.errors.password.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label>Chi nhánh</Label>
            <Select onValueChange={(v) => createForm.setValue("branch_id", Number(v))}>
              <SelectTrigger>
                <SelectValue placeholder="Chọn chi nhánh..." />
              </SelectTrigger>
              <SelectContent>
                {branches.map((b) => (
                  <SelectItem key={b.id} value={b.id.toString()}>
                    {b.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Phòng ban</Label>
              <Input placeholder="Công nghệ" {...createForm.register("department")} />
            </div>
            <div className="space-y-2">
              <Label>Chức vụ</Label>
              <Input placeholder="Developer" {...createForm.register("position")} />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>Huỷ</Button>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Tạo nhân viên
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
