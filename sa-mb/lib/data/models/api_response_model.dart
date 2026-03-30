class ApiResponse<T> {
  final bool success;
  final T? data;
  final String? message;
  final ApiError? error;

  ApiResponse({
    required this.success,
    this.data,
    this.message,
    this.error,
  });

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(dynamic)? fromJsonT,
  ) {
    return ApiResponse(
      success: json['success'] as bool,
      data: json['data'] != null && fromJsonT != null
          ? fromJsonT(json['data'])
          : null,
      message: json['message'] as String?,
      error: json['error'] != null
          ? ApiError.fromJson(json['error'] as Map<String, dynamic>)
          : null,
    );
  }
}

class ApiError {
  final String code;
  final String message;
  final Map<String, String>? fields;

  ApiError({
    required this.code,
    required this.message,
    this.fields,
  });

  factory ApiError.fromJson(Map<String, dynamic> json) {
    return ApiError(
      code: json['code'] as String? ?? '',
      message: json['message'] as String? ?? '',
      fields: json['fields'] != null
          ? Map<String, String>.from(json['fields'] as Map)
          : null,
    );
  }
}
