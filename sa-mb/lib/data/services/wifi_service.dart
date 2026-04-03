import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';
import 'package:network_info_plus/network_info_plus.dart';
import 'package:permission_handler/permission_handler.dart';

class WifiInfo {
  final String ssid;
  final String bssid;

  WifiInfo({required this.ssid, required this.bssid});
}

class WifiService {
  final NetworkInfo _networkInfo = NetworkInfo();

  /// WiFi hardcode cho debug/simulator
  static const _debugSsid = 'Long';
  static const _debugBssid = '72:D0:1F:20:BE:26';

  Future<bool> checkPermission() async {
    if (Platform.isMacOS || Platform.isWindows || Platform.isLinux) {
      if (Platform.isMacOS) {
        LocationPermission permission = await Geolocator.checkPermission();
        if (permission == LocationPermission.denied) {
          permission = await Geolocator.requestPermission();
        }
        return permission == LocationPermission.whileInUse ||
            permission == LocationPermission.always;
      }
      return true;
    }
    final status = await Permission.locationWhenInUse.request();
    return status.isGranted;
  }

  Future<WifiInfo?> getCurrentWifiInfo() async {
    // Debug mode: luôn trả về WiFi hardcode để test trên simulator/desktop
    if (kDebugMode) {
      return WifiInfo(ssid: _debugSsid, bssid: _debugBssid);
    }

    final hasPermission = await checkPermission();
    if (!hasPermission) return null;

    try {
      final ssid = await _networkInfo.getWifiName();
      final bssid = await _networkInfo.getWifiBSSID();

      if (ssid == null || bssid == null) return null;

      final cleanSsid = ssid.replaceAll('"', '');
      return WifiInfo(ssid: cleanSsid, bssid: bssid);
    } catch (_) {
      return null;
    }
  }
}
