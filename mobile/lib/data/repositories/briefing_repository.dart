import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/network/api_client.dart';
import '../../domain/models/briefing.dart';

final briefingRepositoryProvider = Provider<BriefingRepository>((ref) {
  final dio = ref.watch(dioProvider);
  return BriefingRepository(dio);
});

final dailyBriefingProvider = FutureProvider.autoDispose<DailyBriefing>((ref) async {
  final repository = ref.watch(briefingRepositoryProvider);
  return repository.getDailyBriefing();
});

class BriefingRepository {
  final Dio _dio;

  BriefingRepository(this._dio);

  Future<DailyBriefing> getDailyBriefing() async {
    try {
      final response = await _dio.get('/feed/daily');
      return DailyBriefing.fromJson(response.data);
    } on DioException catch (e) {
      // 客户端开发环境降级策略（Backend 尚未就绪时可保证 UI 不中断）
      print('API Error: ${e.message}. Using offline fallback map.');
      return DailyBriefing(
        topInsight: TopInsight(
          title: "The Architecture of Silence",
          summary: "Explore how physical environments dictate cognitive depth. A journey through historical libraries and the modern need for 'The Scholarly Atrium.'",
          imageUrl: "https://images.unsplash.com/photo-1541963463532-d68292c34b19?q=80&w=2500&auto=format&fit=crop"
        ),
        butlerSuggestion: ButlerSuggestion(endurancePct: 82, progressMins: 45, targetMins: 60),
        curatedArticles: [
          CuratedArticle(id: "1", title: "The Moral Ghost in the Machine")
        ],
        quickReads: [
          QuickRead(id: "2", title: "Digital Stoicism", readTime: 6)
        ]
      );
    }
  }
}
