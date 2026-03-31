"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { WifiOff, Wifi } from "lucide-react";
import { Branch } from "@/types/branch";
import { useWifiConfigs } from "@/hooks/use-wifi-configs";

interface WifiConfigDialogProps {
  branch: Branch;
  open: boolean;
  onClose: () => void;
}

export function WifiConfigDialog({ branch, open, onClose }: WifiConfigDialogProps) {
  const { data: configs, isLoading } = useWifiConfigs(branch.id);

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Wifi className="h-5 w-5 text-green-600" />
            Cấu hình WiFi — {branch.name}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground text-sm">
              Đang tải...
            </div>
          ) : configs && configs.length > 0 ? (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>SSID</TableHead>
                    <TableHead>BSSID</TableHead>
                    <TableHead>Mô tả</TableHead>
                    <TableHead>Trạng thái</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {configs.map((config) => (
                    <TableRow key={config.id}>
                      <TableCell className="font-medium text-sm">
                        {config.ssid}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground font-mono">
                        {config.bssid || "—"}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground max-w-[150px] truncate">
                        {config.description || "—"}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={config.is_active ? "success" : "secondary"}
                        >
                          {config.is_active ? (
                            <><Wifi className="h-3 w-3 mr-1" /> Bật</>
                          ) : (
                            <><WifiOff className="h-3 w-3 mr-1" /> Tắt</>
                          )}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground text-sm border rounded-md">
              <WifiOff className="h-8 w-8 mx-auto mb-2 opacity-40" />
              Chưa có mạng WiFi nào được cấu hình
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
