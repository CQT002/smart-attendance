"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Clock, Plus, Pencil, Trash2, Check, X } from "lucide-react";
import { Branch } from "@/types/branch";
import { useShifts, useCreateShift, useUpdateShift, useDeleteShift } from "@/hooks/use-shifts";
import { Shift, DAY_NAMES, CreateShiftRequest, UpdateShiftRequest } from "@/services/shift.service";

interface ShiftConfigDialogProps {
  branch: Branch;
  open: boolean;
  onClose: () => void;
}

const DEFAULT_FORM: CreateShiftRequest = {
  name: "Ca mặc định",
  start_time: "08:00",
  end_time: "17:00",
  late_after: 15,
  early_before: 15,
  work_hours: 8,
  morning_end: "12:00",
  afternoon_start: "13:00",
  regular_end_day: 6,
  regular_end_time: "12:00",
  ot_min_checkin_hour: 17,
  ot_start_hour: 18,
  ot_end_hour: 22,
  is_default: true,
};

export function ShiftConfigDialog({ branch, open, onClose }: ShiftConfigDialogProps) {
  const { data: shifts, isLoading } = useShifts(branch.id);
  const createMut = useCreateShift(branch.id);
  const updateMut = useUpdateShift(branch.id);
  const deleteMut = useDeleteShift(branch.id);

  const [editingId, setEditingId] = useState<number | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState<CreateShiftRequest>(DEFAULT_FORM);

  const resetForm = () => {
    setForm(DEFAULT_FORM);
    setEditingId(null);
    setShowForm(false);
  };

  const handleCreate = async () => {
    await createMut.mutateAsync(form);
    resetForm();
  };

  const handleUpdate = async () => {
    if (!editingId) return;
    const data: UpdateShiftRequest = { ...form };
    await updateMut.mutateAsync({ id: editingId, data });
    resetForm();
  };

  const handleEdit = (shift: Shift) => {
    setForm({
      name: shift.name,
      start_time: shift.start_time,
      end_time: shift.end_time,
      late_after: shift.late_after,
      early_before: shift.early_before,
      work_hours: shift.work_hours,
      morning_end: shift.morning_end,
      afternoon_start: shift.afternoon_start,
      regular_end_day: shift.regular_end_day,
      regular_end_time: shift.regular_end_time,
      ot_min_checkin_hour: shift.ot_min_checkin_hour,
      ot_start_hour: shift.ot_start_hour,
      ot_end_hour: shift.ot_end_hour,
      is_default: shift.is_default,
    });
    setEditingId(shift.id);
    setShowForm(true);
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Xóa ca làm việc này?")) return;
    await deleteMut.mutateAsync(id);
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5 text-blue-600" />
            Ca làm việc — {branch.name}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Danh sách shifts */}
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground text-sm">
              Đang tải...
            </div>
          ) : shifts && shifts.length > 0 ? (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Tên ca</TableHead>
                    <TableHead>Giờ làm</TableHead>
                    <TableHead>Khung chính thức</TableHead>
                    <TableHead>OT</TableHead>
                    <TableHead>Trạng thái</TableHead>
                    <TableHead className="w-[80px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {shifts.map((shift) => (
                    <TableRow key={shift.id}>
                      <TableCell className="font-medium text-sm">
                        {shift.name}
                        {shift.is_default && (
                          <Badge variant="outline" className="ml-2 text-xs">Mặc định</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-sm">
                        {shift.start_time} – {shift.end_time} ({shift.work_hours}h)
                      </TableCell>
                      <TableCell className="text-sm">
                        T2 → {DAY_NAMES[shift.regular_end_day]} {shift.regular_end_time}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {shift.ot_start_hour}:00 – {shift.ot_end_hour}:00
                      </TableCell>
                      <TableCell>
                        <Badge variant={shift.is_active ? "success" : "secondary"}>
                          {shift.is_active ? "Bật" : "Tắt"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleEdit(shift)}>
                            <Pencil className="h-3.5 w-3.5" />
                          </Button>
                          <Button variant="ghost" size="icon" className="h-7 w-7 text-red-500" onClick={() => handleDelete(shift.id)}>
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground text-sm border rounded-md">
              <Clock className="h-8 w-8 mx-auto mb-2 opacity-40" />
              Chưa có ca làm việc nào được cấu hình
            </div>
          )}

          {/* Nút thêm */}
          {!showForm && (
            <Button variant="outline" className="w-full" onClick={() => { resetForm(); setShowForm(true); }}>
              <Plus className="h-4 w-4 mr-2" /> Thêm ca làm việc
            </Button>
          )}

          {/* Form tạo/sửa */}
          {showForm && (
            <div className="border rounded-md p-4 space-y-4 bg-muted/30">
              <h4 className="font-medium text-sm">
                {editingId ? "Chỉnh sửa ca" : "Thêm ca mới"}
              </h4>

              {/* Row 1: Tên + Giờ làm */}
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <Label className="text-xs">Tên ca</Label>
                  <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
                </div>
                <div>
                  <Label className="text-xs">Giờ vào</Label>
                  <Input type="time" value={form.start_time} onChange={(e) => setForm({ ...form, start_time: e.target.value })} />
                </div>
                <div>
                  <Label className="text-xs">Giờ ra</Label>
                  <Input type="time" value={form.end_time} onChange={(e) => setForm({ ...form, end_time: e.target.value })} />
                </div>
              </div>

              {/* Row 2: Khung chính thức */}
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <Label className="text-xs">Ngày cuối tuần làm việc</Label>
                  <Select
                    value={String(form.regular_end_day ?? 6)}
                    onValueChange={(v) => setForm({ ...form, regular_end_day: Number(v) })}
                  >
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      {DAY_NAMES.map((name, idx) => (
                        <SelectItem key={idx} value={String(idx)}>{name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label className="text-xs">Giờ kết thúc ngày cuối</Label>
                  <Input type="time" value={form.regular_end_time ?? "12:00"} onChange={(e) => setForm({ ...form, regular_end_time: e.target.value })} />
                </div>
                <div>
                  <Label className="text-xs">Số giờ làm chuẩn</Label>
                  <Input type="number" step="0.5" value={form.work_hours ?? 8} onChange={(e) => setForm({ ...form, work_hours: Number(e.target.value) })} />
                </div>
              </div>

              {/* Row 3: Nghỉ trưa + Tolerance */}
              <div className="grid grid-cols-4 gap-3">
                <div>
                  <Label className="text-xs">Nghỉ trưa từ</Label>
                  <Input type="time" value={form.morning_end ?? "12:00"} onChange={(e) => setForm({ ...form, morning_end: e.target.value })} />
                </div>
                <div>
                  <Label className="text-xs">Nghỉ trưa đến</Label>
                  <Input type="time" value={form.afternoon_start ?? "13:00"} onChange={(e) => setForm({ ...form, afternoon_start: e.target.value })} />
                </div>
                <div>
                  <Label className="text-xs">Trễ sau (phút)</Label>
                  <Input type="number" value={form.late_after ?? 15} onChange={(e) => setForm({ ...form, late_after: Number(e.target.value) })} />
                </div>
                <div>
                  <Label className="text-xs">Sớm trước (phút)</Label>
                  <Input type="number" value={form.early_before ?? 15} onChange={(e) => setForm({ ...form, early_before: Number(e.target.value) })} />
                </div>
              </div>

              {/* Row 4: OT config */}
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <Label className="text-xs">OT check-in sớm nhất (giờ)</Label>
                  <Input type="number" value={form.ot_min_checkin_hour ?? 17} onChange={(e) => setForm({ ...form, ot_min_checkin_hour: Number(e.target.value) })} />
                </div>
                <div>
                  <Label className="text-xs">OT bắt đầu tính (giờ)</Label>
                  <Input type="number" value={form.ot_start_hour ?? 18} onChange={(e) => setForm({ ...form, ot_start_hour: Number(e.target.value) })} />
                </div>
                <div>
                  <Label className="text-xs">OT kết thúc (giờ)</Label>
                  <Input type="number" value={form.ot_end_hour ?? 22} onChange={(e) => setForm({ ...form, ot_end_hour: Number(e.target.value) })} />
                </div>
              </div>

              {/* Actions */}
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="sm" onClick={resetForm}>
                  <X className="h-4 w-4 mr-1" /> Hủy
                </Button>
                <Button
                  size="sm"
                  onClick={editingId ? handleUpdate : handleCreate}
                  disabled={createMut.isPending || updateMut.isPending}
                >
                  <Check className="h-4 w-4 mr-1" />
                  {editingId ? "Cập nhật" : "Thêm"}
                </Button>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
