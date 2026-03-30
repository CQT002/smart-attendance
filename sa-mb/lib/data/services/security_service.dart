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
        // Common VPN interface names
        if (name.contains('tun') ||
            name.contains('tap') ||
            name.contains('ppp') ||
            name.contains('ipsec') ||
            name.contains('utun')) {
          return true;
        }
      }
      return false;
    } catch (_) {
      return false;
    }
  }
}
