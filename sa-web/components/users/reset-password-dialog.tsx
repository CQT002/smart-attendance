"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2 } from "lucide-react";
import { useResetPassword } from "@/hooks/use-users";

interface ResetPasswordDialogProps {
  open: boolean;
  userId: number;
  userName: string;
  onClose: () => void;
}

export function ResetPasswordDialog({ open, userId, userName, onClose }: ResetPasswordDialogProps) {
  const [password, setPassword] = useState("");
  const resetPassword = useResetPassword();

  const handleSubmit = async () => {
    if (password.length < 8) return;
    await resetPassword.mutateAsync({ id: userId, password });
    setPassword("");
    onClose();
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Reset mật khẩu</DialogTitle>
          <DialogDescription>
            Đặt mật khẩu mới cho <strong>{userName}</strong>
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2">
          <Label>Mật khẩu mới (tối thiểu 8 ký tự)</Label>
          <Input
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Huỷ</Button>
          <Button
            onClick={handleSubmit}
            disabled={password.length < 8 || resetPassword.isPending}
          >
            {resetPassword.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Xác nhận
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
