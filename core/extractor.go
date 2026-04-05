package core

import (
	"regexp"
	"strings"
)

// Precompiled regex patterns for performance.
var (
	sentenceSplitRE = regexp.MustCompile(`[.!?]+\s*`)
	entityMatchRE  = regexp.MustCompile(`\b[A-Z][a-z]+\b`)
)

// FactType represents the type of atomic fact.
type FactType string

const (
	FactTypeEpisodic FactType = "EPISODIC" // Event/action facts
	FactTypeSemantic FactType = "SEMANTIC" // General knowledge
	FactTypeOpinion  FactType = "OPINION"  // Subjective statements
	FactTypeTemporal FactType = "TEMPORAL" // Time-related facts
)

// Extractor handles atomic fact extraction from text.
type Extractor struct{}

// NewExtractor creates a new Extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractedFact represents a single atomic fact.
type ExtractedFact struct {
	Content   string
	FactType  FactType
	Entities  []string
	ParentKey string
}

// ExtractFacts splits input text into atomic facts.
func (e *Extractor) ExtractFacts(text, parentKey string) ([]*ExtractedFact, error) {
	// 1. Split into sentences using punctuation
	sentences := e.splitSentences(text)

	// 2. Classify each sentence
	facts := make([]*ExtractedFact, 0, len(sentences))
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		factType := e.classifyFactType(sentence)
		entities := e.extractEntities(sentence)

		facts = append(facts, &ExtractedFact{
			Content:   sentence,
			FactType:  factType,
			Entities:  entities,
			ParentKey: parentKey,
		})
	}

	return facts, nil
}

// splitSentences splits text into sentences using punctuation.
func (e *Extractor) splitSentences(text string) []string {
	parts := sentenceSplitRE.Split(text, -1)
	return parts
}

// classifyFactType classifies a sentence into a fact type.
func (e *Extractor) classifyFactType(sentence string) FactType {
	lower := strings.ToLower(sentence)

	// Check temporal markers
	temporalMarkers := []string{
		"yesterday", "today", "tomorrow", "last week", "next week",
		"on monday", "on tuesday", "on wednesday", "on thursday", "on friday",
		"on saturday", "on sunday", "in the morning", "in the afternoon", "in the evening",
		"at 3pm", "at 2am", "every day", "every week", "previously", "before", "after",
		"when", "during", "while", "once", "already", "just", "recently", "last month",
		"next month", "last year", "next year", "ago", "since", "until",
	}

	for _, marker := range temporalMarkers {
		if strings.Contains(lower, marker) {
			return FactTypeTemporal
		}
	}

	// Check opinion markers
	opinionMarkers := []string{
		"i think", "i believe", "i feel", "i like", "i hate",
		"i prefer", "probably", "maybe", "perhaps", "might", "could be", "seems",
		"appears", "wonderful", "terrible", "amazing", "awful", "love", "hate",
		"best", "worst", "great", "bad", "good", "better", "worse", "favorite",
		"dislike", "enjoy", "recommend", "suggest", "avoid",
	}

	for _, marker := range opinionMarkers {
		if strings.Contains(lower, marker) {
			return FactTypeOpinion
		}
	}

	// Check episodic markers (action verbs)
	episodicMarkers := []string{
		"created", "deleted", "updated", "modified", "fixed",
		"completed", "started", "finished", "implemented", "designed", "built",
		"wrote", "read", "went", "came", "met", "talked", "called", "sent", "received",
		"bought", "sold", "moved", "changed", "added", "removed", "installed",
		"configured", "deployed", "released", "merged", "opened", "closed", "approved",
	}

	for _, marker := range episodicMarkers {
		if strings.Contains(lower, marker) {
			return FactTypeEpisodic
		}
	}

	// Default to semantic
	return FactTypeSemantic
}

// extractEntities extracts entities from a sentence.
// Uses simple regex patterns for capitalized words.
func (e *Extractor) extractEntities(sentence string) []string {
	matches := entityMatchRE.FindAllString(sentence, -1)

	// Deduplicate
	seen := make(map[string]bool)
	entities := make([]string, 0)
	for _, m := range matches {
		if !seen[m] && len(m) > 2 { // Filter short words
			seen[m] = true
			entities = append(entities, m)
		}
	}

	return entities
}
