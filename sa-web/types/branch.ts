export interface Branch {
  id: number;
  code: string;
  name: string;
  address: string;
  city: string;
  province: string;
  phone: string;
  email: string;
  latitude: number | null;
  longitude: number | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  // GPS Geofencing
  gps_radius?: number;
  // WiFi config (returned from detail endpoint)
  wifi_networks?: WifiNetwork[];
}

export interface WifiNetwork {
  id: number;
  branch_id: number;
  ssid: string;
  bssid: string;
  is_active: boolean;
}

export interface CreateBranchRequest {
  code: string;
  name: string;
  address: string;
  city?: string;
  province?: string;
  phone: string;
  email: string;
  latitude?: number;
  longitude?: number;
  gps_radius?: number;
}

export interface UpdateBranchRequest {
  name: string;
  address: string;
  city?: string;
  province?: string;
  phone: string;
  email: string;
  latitude?: number;
  longitude?: number;
  gps_radius?: number;
}

export interface BranchFilter {
  search?: string;
  is_active?: boolean;
  page?: number;
  limit?: number;
}
