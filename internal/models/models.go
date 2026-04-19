package models

import "time"

type Goal struct {
	ID          int64
	Title       string
	Description string
	Category    string
	Progress    int
	Status      string
	TargetDate  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GoalProgressPoint struct {
	ID         int64
	GoalID     int64
	Progress   int
	RecordedAt time.Time
}

type MoodLog struct {
	ID       int64
	Score    int
	Note     string
	LoggedAt time.Time
}

type JournalEntry struct {
	ID        int64
	Date      time.Time
	Content   string
	WordCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Todo struct {
	ID          int64
	Text        string
	Done        bool
	GoalID      *int64
	Priority    int
	DueDate     *time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type ChatMessage struct {
	ID        int64
	Role      string
	Content   string
	CreatedAt time.Time
}
