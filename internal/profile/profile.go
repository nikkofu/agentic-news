package profile

type UserProfile struct {
	FocusTopics      []string
	PreferredStyles  []string
	ExplicitFeedback map[string]string
	BehaviorSignals  map[string]float64
	TopicAffinity    map[string]float64
}

func ApplyExplicitFeedback(p UserProfile, topic string, sentiment string) UserProfile {
	if p.TopicAffinity == nil {
		p.TopicAffinity = map[string]float64{}
	}
	switch sentiment {
	case "like":
		p.TopicAffinity[topic] += 5
	case "disagree":
		p.TopicAffinity[topic] -= 5
	default:
		p.TopicAffinity[topic] += 0
	}
	return p
}

func ApplyBehaviorSignals(p UserProfile, topic string, dwellSeconds float64, bookmarked bool) UserProfile {
	if p.TopicAffinity == nil {
		p.TopicAffinity = map[string]float64{}
	}
	p.TopicAffinity[topic] += dwellSeconds / 60.0
	if bookmarked {
		p.TopicAffinity[topic] += 2
	}
	return p
}
