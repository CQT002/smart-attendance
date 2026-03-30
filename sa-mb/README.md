# Smart Attendance - Mobile App (Flutter)

Ung dung di dong cho nhan vien HDBBank thuc hien cham cong bang WiFi hoac GPS.

---

## Muc luc

1. [Yeu cau he thong](#1-yeu-cau-he-thong)
2. [Cai dat moi truong](#2-cai-dat-moi-truong)
3. [Khoi tao du an](#3-khoi-tao-du-an)
4. [Cau hinh API Backend](#4-cau-hinh-api-backend)
5. [Cau hinh Firebase (Push Notification)](#5-cau-hinh-firebase-push-notification)
6. [Chay ung dung](#6-chay-ung-dung)
7. [Cau truc source code](#7-cau-truc-source-code)
8. [Giai thich kien truc](#8-giai-thich-kien-truc)
9. [Danh sach dependencies](#9-danh-sach-dependencies)
10. [Cac man hinh chinh](#10-cac-man-hinh-chinh)
11. [Luu y khi phat trien](#11-luu-y-khi-phat-trien)

---

## 1. Yeu cau he thong

| Tool       | Phien ban toi thieu | Kiem tra                  |
|------------|---------------------|---------------------------|
| Flutter    | 3.16+               | `flutter --version`       |
| Dart       | 3.2+                | `dart --version`          |
| Android Studio | 2023.1+         | Hoac VS Code + Extensions |
| Xcode      | 15+ (chi macOS)     | `xcode-select --version`  |
| CocoaPods  | 1.14+ (chi macOS)   | `pod --version`           |

### Kiem tra moi truong Flutter da san sang chua

```bash
flutter doctor
```

Tat ca muc can hien thi dau **[v]**. Neu co **[!]** hoac **[x]**, lam theo huong dan Flutter hien thi de sua.

---

## 2. Cai dat moi truong

### macOS

```bash
# Cai Flutter qua Homebrew
brew install --cask flutter

# Hoac tai tu trang chu: https://docs.flutter.dev/get-started/install/macos

# Cai CocoaPods (can thiet cho iOS)
sudo gem install cocoapods
```

### Windows

```bash
# Tai Flutter SDK tu: https://docs.flutter.dev/get-started/install/windows
# Giai nen va them vao PATH

# Kiem tra
flutter doctor
```

### VS Code Extensions can thiet

- **Flutter** (Dart-Code.flutter)
- **Dart** (Dart-Code.dart-code)

### Android Studio Plugins

- **Flutter Plugin**
- **Dart Plugin**

---

## 3. Khoi tao du an

```bash
# Buoc 1: Di chuyen vao thu muc du an
cd sa-mb

# Buoc 2: Sinh platform folders (android/, ios/, web/, ...)
# Lenh nay chi can chay 1 lan duy nhat
flutter create .

# Buoc 3: Cai dat dependencies
flutter pub get

# Buoc 4 (chi iOS): Cai CocoaPods dependencies
cd ios && pod install && cd ..
```

### Neu gap loi permission khi chay `flutter create .`

```bash
# Tao thu muc config neu chua co
mkdir -p ~/.config/flutter

# Thu lai
flutter create .
```

---

## 4. Cau hinh API Backend

Mo file `lib/core/constants/api_constants.dart` va sua `baseUrl` theo moi truong:

```dart
class ApiConstants {
  // Development - chay local
  static const String baseUrl = 'http://localhost:8080';

  // Development - Android Emulator ket noi localhost
  // static const String baseUrl = 'http://10.0.2.2:8080';

  // Development - Thiet bi that cung mang WiFi
  // static const String baseUrl = 'http://192.168.x.x:8080';

  // Production
  // static const String baseUrl = 'https://api.smartattendance.hdbank.com.vn';
}
```

> **Luu y**: Android Emulator khong truy cap duoc `localhost` cua may host.
> Phai dung `10.0.2.2` (Android Emulator) hoac IP LAN thuc te.

---

## 5. Cau hinh Firebase (Push Notification)

Push Notification yeu cau Firebase. Neu chua can tinh nang nay, bo qua buoc nay.

### Buoc 1: Tao project tren Firebase Console

1. Truy cap https://console.firebase.google.com
2. Tao project moi hoac chon project co san
3. Them app Android va iOS

### Buoc 2: Tai file cau hinh

- **Android**: Tai `google-services.json` -> dat vao `android/app/`
- **iOS**: Tai `GoogleService-Info.plist` -> dat vao `ios/Runner/`

### Buoc 3: Neu chua co Firebase, tam thoi comment lai trong pubspec.yaml

```yaml
# firebase_core: ^2.27.0
# firebase_messaging: ^14.7.0
# flutter_local_notifications: ^17.0.0
```

---

## 6. Chay ung dung

### Kiem tra thiet bi ket noi

```bash
flutter devices
```

### Chay tren Android Emulator

```bash
# Mo emulator
flutter emulators --launch <ten_emulator>

# Chay app
flutter run
```

### Chay tren iOS Simulator (chi macOS)

```bash
# Mo simulator
open -a Simulator

# Chay app
flutter run
```

### Chay tren thiet bi that

```bash
# Ket noi dien thoai qua USB
# Android: Bat Developer Options > USB Debugging
# iOS: Trust thiet bi trong Xcode

flutter run
```

### Chay che do Release (test hieu nang)

```bash
flutter run --release
```

### Build APK (Android)

```bash
flutter build apk --release
# File output: build/app/outputs/flutter-apk/app-release.apk
```

### Build IPA (iOS)

```bash
flutter build ipa --release
# Can Apple Developer Account va Xcode signing
```

---

## 7. Cau truc source code

```
sa-mb/
├── pubspec.yaml                                  # Khai bao dependencies va config
├── README.md                                     # File nay
├── assets/
│   ├── images/                                   # Hinh anh (logo, illustrations)
│   └── icons/                                    # Icon SVG tuy chinh
│
└── lib/                                          # Toan bo source code Dart
    ├── main.dart                                 # Entry point - khoi dong app
    ├── app.dart                                  # Root widget, Dependency Injection, Routing
    │
    ├── core/                                     # --- CORE: Chia se toan app ---
    │   ├── constants/
    │   │   ├── api_constants.dart                # URL endpoints API backend
    │   │   └── app_constants.dart                # Hang so: token keys, status codes
    │   ├── network/
    │   │   └── api_client.dart                   # Dio HTTP client + JWT interceptor
    │   ├── theme/
    │   │   ├── app_colors.dart                   # Bang mau HDBank + status colors
    │   │   └── app_theme.dart                    # Material Design 3: Light + Dark theme
    │   └── utils/
    │       └── date_utils.dart                   # Format ngay/gio, tinh tuan/thang
    │
    ├── data/                                     # --- DATA LAYER: Giao tiep ben ngoai ---
    │   ├── models/                               # Data models (JSON <-> Dart objects)
    │   │   ├── api_response_model.dart           # ApiResponse<T> + ApiError
    │   │   ├── user_model.dart                   # UserModel (khop entity.User backend)
    │   │   ├── branch_model.dart                 # BranchModel (khop entity.Branch)
    │   │   ├── attendance_model.dart             # AttendanceModel (khop entity.AttendanceLog)
    │   │   └── login_response_model.dart         # LoginResponseModel (token + user)
    │   ├── repositories/                         # Hien thuc cac repository interface
    │   │   ├── auth_repository_impl.dart         # Goi API: login, getMe, changePassword
    │   │   └── attendance_repository_impl.dart   # Goi API: check-in, check-out, history
    │   └── services/                             # Platform services (GPS, WiFi, Device)
    │       ├── location_service.dart             # GPS: lay toa do, kiem tra geofence, mock
    │       ├── wifi_service.dart                 # WiFi: lay SSID/BSSID hien tai
    │       ├── device_service.dart               # Device: lay device_id, model, app version
    │       └── security_service.dart             # Anti-fraud: detect VPN + Fake GPS
    │
    ├── domain/                                   # --- DOMAIN LAYER: Business logic ---
    │   └── repositories/                         # Abstract interfaces (khong phu thuoc data)
    │       ├── auth_repository.dart              # Interface: login, logout, getMe
    │       └── attendance_repository.dart        # Interface: checkIn, checkOut, getHistory
    │
    └── presentation/                             # --- PRESENTATION LAYER: UI ---
        ├── blocs/                                # BLoC state management
        │   ├── auth/
        │   │   ├── auth_bloc.dart                # Xu ly: login, logout, check session
        │   │   ├── auth_event.dart               # Events: LoginRequested, LogoutRequested
        │   │   └── auth_state.dart               # States: Authenticated, Unauthenticated
        │   └── attendance/
        │       ├── attendance_bloc.dart           # Xu ly: check-in/out voi anti-fraud logic
        │       ├── attendance_event.dart          # Events: CheckIn, CheckOut, LoadHistory
        │       └── attendance_state.dart          # States: TodayLoaded, HistoryLoaded
        ├── screens/                              # Cac man hinh chinh
        │   ├── login_screen.dart                 # Man hinh dang nhap
        │   ├── home_screen.dart                  # Trang chu + Profile + Bottom Navigation
        │   ├── check_in_screen.dart              # Chon phuong thuc WiFi/GPS va cham cong
        │   └── history_screen.dart               # Lich su cham cong (tuan/thang)
        └── widgets/                              # Widget tai su dung
            ├── attendance_card.dart              # Card hien thi 1 ngay cham cong
            ├── status_badge.dart                 # Badge trang thai (Dung gio/Tre/Vang)
            └── loading_overlay.dart              # Overlay loading khi xu ly
```

---

## 8. Giai thich kien truc

App su dung **Clean Architecture** voi 3 layers:

```
┌─────────────────────────────────────────────┐
│           PRESENTATION (UI)                  │
│  Screens, Widgets, BLoC (State Management)   │
│  Chi biet domain layer, KHONG biet data      │
├─────────────────────────────────────────────┤
│              DOMAIN (Business Logic)         │
│  Repository interfaces                       │
│  KHONG phu thuoc bat ky layer nao khac       │
├─────────────────────────────────────────────┤
│              DATA (External)                 │
│  Models, Repository Impl, API Client         │
│  Services (GPS, WiFi, Device)                │
│  Hien thuc cac interface cua Domain          │
└─────────────────────────────────────────────┘
```

### Luong du lieu khi Check-in

```
User nhan nut Check-in
  -> CheckInScreen chon WiFi hoac GPS
    -> AttendanceBloc nhan AttendanceCheckInRequested
      -> SecurityService kiem tra VPN + Fake GPS
      -> DeviceService lay device_id, model
      -> LocationService lay GPS / WifiService lay SSID+BSSID
      -> AttendanceRepository.checkIn() goi API backend
        -> ApiClient gui POST /api/v1/attendance/check-in voi JWT
          -> Backend validate + luu database
            -> Tra ve AttendanceModel
              -> BLoC emit AttendanceCheckInSuccess
                -> UI cap nhat trang thai
```

### State Management: BLoC Pattern

```
Event (hanh dong) -> BLoC (xu ly) -> State (ket qua) -> UI (hien thi)

Vi du:
AuthLoginRequested -> AuthBloc -> AuthAuthenticated -> HomeScreen
AttendanceCheckInRequested -> AttendanceBloc -> AttendanceCheckInSuccess -> SnackBar
```

---

## 9. Danh sach dependencies

| Package                    | Muc dich                                    |
|----------------------------|---------------------------------------------|
| `flutter_bloc`             | State management theo BLoC pattern          |
| `equatable`                | So sanh objects trong BLoC states/events     |
| `dio`                      | HTTP client goi API backend                 |
| `geolocator`               | Lay toa do GPS, kiem tra geofence           |
| `network_info_plus`        | Lay thong tin WiFi (SSID, BSSID)            |
| `device_info_plus`         | Lay device ID va model dien thoai           |
| `flutter_secure_storage`   | Luu JWT token bao mat (Keychain/Keystore)   |
| `package_info_plus`        | Lay app version                             |
| `firebase_core`            | Firebase SDK core                           |
| `firebase_messaging`       | Push notification                           |
| `flutter_local_notifications` | Hien thi notification local              |
| `google_fonts`             | Font Inter cho UI                           |
| `intl`                     | Format ngay gio, da ngon ngu                |
| `shimmer`                  | Hieu ung loading skeleton                   |
| `cached_network_image`     | Cache anh tu network                        |
| `flutter_svg`              | Hien thi icon SVG                           |
| `shared_preferences`       | Luu cai dat don gian (theme, etc.)          |
| `connectivity_plus`        | Kiem tra ket noi mang                       |
| `permission_handler`       | Xin quyen Location, WiFi                    |

---

## 10. Cac man hinh chinh

### Man hinh Dang nhap (`login_screen.dart`)

- Nhap email va mat khau nhan vien
- Validation form: email hop le, mat khau >= 6 ky tu
- Hien thi loi tu backend (sai mat khau, tai khoan bi khoa, ...)
- Tu dong chuyen sang Home khi dang nhap thanh cong

### Man hinh Trang chu (`home_screen.dart`)

- **Tab Trang chu**: Loi chao theo buoi, thong tin chi nhanh, card hien thi check-in/out hom nay, nut hanh dong
- **Tab Lich su**: Xem lich su cham cong theo tuan/thang
- **Tab Ca nhan**: Thong tin nhan vien, ma NV, email, chi nhanh, nut dang xuat

### Man hinh Check-in/out (`check_in_screen.dart`)

- Chon phuong thuc: **WiFi** (quet SSID/BSSID) hoac **GPS** (Geofencing)
- Thong bao bao mat: he thong se kiem tra VPN, Fake GPS, Device ID
- Nut thuc hien check-in hoac check-out
- Tu dong quay lai Home khi thanh cong

### Man hinh Lich su (`history_screen.dart`)

- 3 tab: Tuan nay / Thang nay / Thang truoc
- Danh sach card cham cong voi: ngay, gio vao/ra, phuong thuc (WiFi/GPS), trang thai, gio lam
- Trang thai hien thi bang badge mau: Dung gio (xanh), Tre (cam), Vang (do)

---

## 11. Luu y khi phat trien

### Doi API URL khi test tren thiet bi that

Android Emulator khong truy cap duoc `localhost`. Sua trong `api_constants.dart`:
- Emulator Android: `http://10.0.2.2:8080`
- Thiet bi that: dung IP LAN cua may chay backend (vd: `http://192.168.1.100:8080`)
- iOS Simulator: `http://localhost:8080` (truy cap duoc binh thuong)

### Quyen can cap (Permissions)

App se tu dong xin quyen khi can. Nguoi dung can cho phep:

| Quyen               | Khi nao can              | Neu tu choi                    |
|----------------------|--------------------------|--------------------------------|
| Location (GPS)       | Check-in bang GPS         | Khong the check-in bang GPS    |
| Location (WiFi scan) | Check-in bang WiFi        | Khong doc duoc SSID/BSSID      |

### Android: Cau hinh them cho WiFi scanning

Them vao `android/app/src/main/AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_WIFI_STATE" />
<uses-permission android:name="android.permission.CHANGE_WIFI_STATE" />
```

### iOS: Cau hinh them cho Location

Them vao `ios/Runner/Info.plist`:

```xml
<key>NSLocationWhenInUseUsageDescription</key>
<string>Can vi tri de xac minh cham cong tai chi nhanh.</string>
<key>NSLocationAlwaysUsageDescription</key>
<string>Can vi tri de xac minh cham cong tai chi nhanh.</string>
```

### Anti-fraud: Cach hoat dong

1. **Fake GPS**: Su dung `position.isMocked` cua `geolocator` - phat hien app gia lap vi tri
2. **VPN**: Kiem tra network interfaces co ten `tun`, `tap`, `ppp`, `ipsec`, `utun`
3. **Device ID**: Gui `device_id` duy nhat moi thiet bi - backend kiem tra chong dung chung tai khoan
4. Ket qua anti-fraud duoc gui len backend, backend quyet dinh cho phep hay tu choi

### Hot Reload khi phat trien

```bash
# Chay app
flutter run

# Trong terminal:
# Nhan 'r' -> Hot Reload (nhanh, giu state)
# Nhan 'R' -> Hot Restart (chay lai tu dau)
# Nhan 'q' -> Thoat
```

### Kiem tra loi code

```bash
flutter analyze
```

### Chay tests

```bash
flutter test
```
