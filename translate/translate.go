package translate

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Translator struct {
	ruEnWords map[string]string
}

func New() Translator {
	return Translator{
		ruEnWords: map[string]string{
			"утренний": "morning",
			"утренняя": "morning",

			"полуденный": "lunch",
			"полуденная": "lunch",

			"дневной": "afternoon",
			"дневная": "afternoon",

			"вечерний": "evening",
			"вечерняя": "evening",

			"ночной": "night",
			"ночная": "night",

			"забег":      "run",
			"заезд":      "ride",
			"заплыв":     "swim",
			"ходьба":     "walk",
			"велозаезд":  "ride",
			"тренировка": "workout",
		},
	}
}

func (t Translator) ActivityName(name string) string {
	nameLower := strings.ToLower(strings.TrimSpace(name))

	timeOfDay, sportType, ok := strings.Cut(nameLower, " ")
	if !ok {
		return name
	}
	if timeOfDay == "" || sportType == "" {
		return name
	}

	trTimeOfDay, ok := t.ruEnWords[timeOfDay]
	if !ok {
		return name
	}

	trSportType, ok := t.ruEnWords[sportType]
	if !ok {
		return name
	}

	title := cases.Title(language.English)

	return title.String(trTimeOfDay) + " " + title.String(trSportType)
}
