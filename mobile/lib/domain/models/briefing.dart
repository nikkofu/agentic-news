import 'package:json_annotation/json_annotation.dart';

part 'briefing.g.dart';

@JsonSerializable()
class DailyBriefing {
  @JsonKey(name: 'top_insight')
  final TopInsight? topInsight;
  
  @JsonKey(name: 'butler_suggestion')
  final ButlerSuggestion? butlerSuggestion;
  
  @JsonKey(name: 'curated_articles')
  final List<CuratedArticle>? curatedArticles;

  @JsonKey(name: 'quick_reads')
  final List<QuickRead>? quickReads;

  DailyBriefing({this.topInsight, this.butlerSuggestion, this.curatedArticles, this.quickReads});

  factory DailyBriefing.fromJson(Map<String, dynamic> json) => _$DailyBriefingFromJson(json);
  Map<String, dynamic> toJson() => _$DailyBriefingToJson(this);
}

@JsonSerializable()
class TopInsight {
  final String title;
  final String summary;
  @JsonKey(name: 'image_url')
  final String imageUrl;

  TopInsight({required this.title, required this.summary, required this.imageUrl});
  factory TopInsight.fromJson(Map<String, dynamic> json) => _$TopInsightFromJson(json);
  Map<String, dynamic> toJson() => _$TopInsightToJson(this);
}

@JsonSerializable()
class ButlerSuggestion {
  @JsonKey(name: 'endurance_pct')
  final int endurancePct;
  @JsonKey(name: 'progress_mins')
  final int progressMins;
  @JsonKey(name: 'target_mins')
  final int targetMins;

  ButlerSuggestion({required this.endurancePct, required this.progressMins, required this.targetMins});
  factory ButlerSuggestion.fromJson(Map<String, dynamic> json) => _$ButlerSuggestionFromJson(json);
  Map<String, dynamic> toJson() => _$ButlerSuggestionToJson(this);
}

@JsonSerializable()
class CuratedArticle {
  final String id;
  final String title;

  CuratedArticle({required this.id, required this.title});
  factory CuratedArticle.fromJson(Map<String, dynamic> json) => _$CuratedArticleFromJson(json);
  Map<String, dynamic> toJson() => _$CuratedArticleToJson(this);
}

@JsonSerializable()
class QuickRead {
  final String id;
  final String title;
  @JsonKey(name: 'read_time')
  final int readTime;

  QuickRead({required this.id, required this.title, required this.readTime});
  factory QuickRead.fromJson(Map<String, dynamic> json) => _$QuickReadFromJson(json);
  Map<String, dynamic> toJson() => _$QuickReadToJson(this);
}
