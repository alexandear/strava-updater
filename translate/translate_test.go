package translate

import (
	"testing"
)

func TestTranslator_ActivityName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Recovery",
			expected: "Recovery",
		},
		{
			input:    "Evening Run",
			expected: "Evening Run",
		},
		{
			input:    "Night run",
			expected: "Night run",
		},
		{
			input:    "Утренний забег",
			expected: "Morning Run",
		},
		{
			input:    "Полуденный забег",
			expected: "Lunch Run",
		},
		{
			input:    "Дневной забег",
			expected: "Afternoon Run",
		},
		{
			input:    "Вечерний забег",
			expected: "Evening Run",
		},
		{
			input:    "Ночной забег",
			expected: "Night Run",
		},
		{
			input:    "Утренний велозаезд",
			expected: "Morning Ride",
		},
		{
			input:    "Вечерний велозаезд",
			expected: "Evening Ride",
		},
		{
			input:    "Ночной заезд",
			expected: "Night Ride",
		},
		{
			input:    "Утренняя тренировка",
			expected: "Morning Workout",
		},
		{
			input:    "Дневная тренировка",
			expected: "Afternoon Workout",
		},
		{
			input:    "Вечерняя тренировка",
			expected: "Evening Workout",
		},
		{
			input:    "Ночная ходьба",
			expected: "Night Walk",
		},
	}

	translator := New()
	for _, tt := range tests {
		got := translator.ActivityName(tt.input)
		if got != tt.expected {
			t.Errorf("Expected %q, but got %q", tt.expected, got)
		}
	}
}
