// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'briefing.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DailyBriefing _$DailyBriefingFromJson(Map<String, dynamic> json) =>
    DailyBriefing(
      topInsight: json['top_insight'] == null
          ? null
          : TopInsight.fromJson(json['top_insight'] as Map<String, dynamic>),
      butlerSuggestion: json['butler_suggestion'] == null
          ? null
          : ButlerSuggestion.fromJson(
              json['butler_suggestion'] as Map<String, dynamic>,
            ),
      curatedArticles: (json['curated_articles'] as List<dynamic>?)
          ?.map((e) => CuratedArticle.fromJson(e as Map<String, dynamic>))
          .toList(),
      quickReads: (json['quick_reads'] as List<dynamic>?)
          ?.map((e) => QuickRead.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$DailyBriefingToJson(DailyBriefing instance) =>
    <String, dynamic>{
      'top_insight': instance.topInsight,
      'butler_suggestion': instance.butlerSuggestion,
      'curated_articles': instance.curatedArticles,
      'quick_reads': instance.quickReads,
    };

TopInsight _$TopInsightFromJson(Map<String, dynamic> json) => TopInsight(
  title: json['title'] as String,
  summary: json['summary'] as String,
  imageUrl: json['image_url'] as String,
);

Map<String, dynamic> _$TopInsightToJson(TopInsight instance) =>
    <String, dynamic>{
      'title': instance.title,
      'summary': instance.summary,
      'image_url': instance.imageUrl,
    };

ButlerSuggestion _$ButlerSuggestionFromJson(Map<String, dynamic> json) =>
    ButlerSuggestion(
      endurancePct: (json['endurance_pct'] as num).toInt(),
      progressMins: (json['progress_mins'] as num).toInt(),
      targetMins: (json['target_mins'] as num).toInt(),
    );

Map<String, dynamic> _$ButlerSuggestionToJson(ButlerSuggestion instance) =>
    <String, dynamic>{
      'endurance_pct': instance.endurancePct,
      'progress_mins': instance.progressMins,
      'target_mins': instance.targetMins,
    };

CuratedArticle _$CuratedArticleFromJson(Map<String, dynamic> json) =>
    CuratedArticle(id: json['id'] as String, title: json['title'] as String);

Map<String, dynamic> _$CuratedArticleToJson(CuratedArticle instance) =>
    <String, dynamic>{'id': instance.id, 'title': instance.title};

QuickRead _$QuickReadFromJson(Map<String, dynamic> json) => QuickRead(
  id: json['id'] as String,
  title: json['title'] as String,
  readTime: (json['read_time'] as num).toInt(),
);

Map<String, dynamic> _$QuickReadToJson(QuickRead instance) => <String, dynamic>{
  'id': instance.id,
  'title': instance.title,
  'read_time': instance.readTime,
};
