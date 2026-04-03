import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';

class LocationService {
  /// GPS hardcode cho debug/simulator — toạ độ chi nhánh TP.HCM
  static const _debugLat = 15.8102;
  static const _debugLng = 105.879;

  Future<bool> checkPermission() async {
    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) return false;

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) return false;
    }

    if (permission == LocationPermission.deniedForever) return false;

    return true;
  }

  Position _debugPosition() => Position(
        latitude: _debugLat,
        longitude: _debugLng,
        timestamp: DateTime.now(),
        accuracy: 10.0,
        altitude: 0.0,
        altitudeAccuracy: 0.0,
        heading: 0.0,
        headingAccuracy: 0.0,
        speed: 0.0,
        speedAccuracy: 0.0,
      );

  Future<Position?> getCurrentPosition() async {
    // Debug mode: luôn trả về tọa độ hardcode để test trên simulator/desktop
    if (kDebugMode) {
      return _debugPosition();
    }

    final hasPermission = await checkPermission();
    if (!hasPermission) return null;

    return await Geolocator.getCurrentPosition(
      desiredAccuracy: LocationAccuracy.high,
    );
  }

  /// Check if position is within radius (meters) of target
  bool isWithinGeofence({
    required double currentLat,
    required double currentLng,
    required double targetLat,
    required double targetLng,
    required double radiusMeters,
  }) {
    final distance = Geolocator.distanceBetween(
      currentLat,
      currentLng,
      targetLat,
      targetLng,
    );
    return distance <= radiusMeters;
  }

  /// Check if mock/fake GPS is enabled
  Future<bool> isMockLocation() async {
    if (kDebugMode) return false;

    try {
      final position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
      );
      return position.isMocked;
    } catch (_) {
      return false;
    }
  }
}
