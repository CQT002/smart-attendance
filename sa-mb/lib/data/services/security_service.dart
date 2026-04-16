import 'dart:io';
import 'location_service.dart';

class SecurityCheckResult {
  final bool isFakeGps;
  final bool isVpn;

  SecurityCheckResult({
    required this.isFakeGps,
    required this.isVpn,
  });

  bool get hasFraud => isFakeGps || isVpn;
}

class SecurityService {
  final LocationService _locationService;

  SecurityService(this._locationService);

  Future<SecurityCheckResult> performSecurityCheck() async {
    final results = await Future.wait([
      _checkFakeGps(),
      _checkVpn(),
    ]);

    return SecurityCheckResult(
      isFakeGps: results[0],
      isVpn: results[1],
    );
  }

  Future<bool> _checkFakeGps() async {
    return await _locationService.isMockLocation();
  }

  Future<bool> _checkVpn() async {
    try {
      final interfaces = await NetworkInterface.list(
        type: InternetAddressType.any,
        includeLinkLocal: false,
        includeLoopback: false,
      );

      for (final interface_ in interfaces) {
        final name = interface_.name.toLowerCase();
        // iOS/macOS luôn có utun0-utunN cho system services (iCloud Private Relay, Hotspot...)
        // → chỉ flag khi interface là VPN thật (tun/tap/ppp/ipsec), bỏ qua utun (Apple system)
        if (name.startsWith('tun') && !name.startsWith('utun')) return true;
        if (name.contains('tap') ||
            name.contains('ppp') ||
            name.startsWith('ipsec')) {
          return true;
        }
      }
      return false;
    } catch (_) {
      return false;
    }
  }
}
