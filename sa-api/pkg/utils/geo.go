package utils

import "math"

const earthRadiusKm = 6371.0

// HaversineDistance tính khoảng cách (mét) giữa 2 tọa độ GPS
// Sử dụng công thức Haversine để tính khoảng cách trên mặt cầu
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := toRadians(lat2 - lat1)
	dLng := toRadians(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distanceKm := earthRadiusKm * c

	return distanceKm * 1000 // Chuyển về mét
}

// IsWithinGeofence kiểm tra tọa độ có nằm trong vùng geofence không
// lat/lng: tọa độ cần kiểm tra
// centerLat/centerLng: tâm của geofence
// radiusMeters: bán kính cho phép (mét)
func IsWithinGeofence(lat, lng, centerLat, centerLng, radiusMeters float64) bool {
	distance := HaversineDistance(lat, lng, centerLat, centerLng)
	return distance <= radiusMeters
}

// IsValidCoordinate kiểm tra tọa độ GPS có hợp lệ không
func IsValidCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// IsZeroCoordinate kiểm tra tọa độ có bằng 0 không (dấu hiệu GPS giả hoặc không có GPS)
func IsZeroCoordinate(lat, lng float64) bool {
	return lat == 0 && lng == 0
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
