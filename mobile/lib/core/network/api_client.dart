import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(BaseOptions(
    baseUrl: 'http://localhost:8080/api', // Connects to the Golang backend
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 10),
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    }
  ));

  dio.interceptors.add(InterceptorsWrapper(
    onRequest: (options, handler) {
      // Future: Inject JWT token from Riverpod/SecureStorage here
      return handler.next(options);
    },
    onError: (DioException e, handler) {
      // Global error handling strategy
      return handler.next(e);
    }
  ));

  return dio;
});
