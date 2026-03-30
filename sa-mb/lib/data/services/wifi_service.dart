import 'package:network_info_plus/network_info_plus.dart';
import 'package:permission_handler/permission_handler.dart';

class WifiInfo {
  final String ssid;
  final String bssid;

  WifiInfo({required this.ssid, required this.bssid});
}

class WifiService {
  final NetworkInfo _networkInfo = NetworkInfo();

  Future<bool> checkPermission() async {
    // On Android, location permission is needed to get WiFi info
    final status = await Permission.locationWhenInUse.request();
    return status.isGranted;
  }

  Future<WifiInfo?> getCurrentWifiInfo() async {
    final hasPermission = await checkPermission();
    if (!hasPermission) return null;

    try {
      final ssid = await _networkInfo.getWifiName();
      final bssid = await _networkInfo.getWifiBSSID();

      if (ssid == null || bssid == null) return null;

      // Remove quotes from SSID if present
      final cleanSsid = ssid.replaceAll('"', '');

      return WifiInfo(ssid: cleanSsid, bssid: bssid);
    } catch (_) {
      return null;
    }
  }
}
