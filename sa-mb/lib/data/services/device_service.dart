import 'dart:io';
import 'package:device_info_plus/device_info_plus.dart';
import 'package:package_info_plus/package_info_plus.dart';

class DeviceInfo {
  final String deviceId;
  final String deviceModel;
  final String appVersion;

  DeviceInfo({
    required this.deviceId,
    required this.deviceModel,
    required this.appVersion,
  });
}

class DeviceService {
  final DeviceInfoPlugin _deviceInfo = DeviceInfoPlugin();

  Future<DeviceInfo> getDeviceInfo() async {
    final packageInfo = await PackageInfo.fromPlatform();
    String deviceId;
    String deviceModel;

    if (Platform.isAndroid) {
      final androidInfo = await _deviceInfo.androidInfo;
      deviceId = androidInfo.id;
      deviceModel = '${androidInfo.manufacturer} ${androidInfo.model}';
    } else if (Platform.isIOS) {
      final iosInfo = await _deviceInfo.iosInfo;
      deviceId = iosInfo.identifierForVendor ?? 'unknown';
      deviceModel = iosInfo.utsname.machine;
    } else {
      deviceId = 'unknown';
      deviceModel = 'unknown';
    }

    return DeviceInfo(
      deviceId: deviceId,
      deviceModel: deviceModel,
      appVersion: '${packageInfo.version}+${packageInfo.buildNumber}',
    );
  }
}
