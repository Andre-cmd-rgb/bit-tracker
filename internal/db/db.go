package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/andre-cmd-rgb/bit-tracker/internal/models"
)

type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS goals (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT DEFAULT '',
	category TEXT DEFAULT 'personal',
	progress INTEGER DEFAULT 0,
	status TEXT DEFAULT 'active',
	target_date DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS goal_progress_history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	goal_id INTEGER NOT NULL,
	progress INTEGER NOT NULL,
	recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(goal_id) REFERENCES goals(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS mood_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	score INTEGER NOT NULL,
	note TEXT DEFAULT '',
	logged_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS journal_entries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date DATE UNIQUE NOT NULL,
	content TEXT DEFAULT '',
	word_count INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS todos (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	text TEXT NOT NULL,
	done INTEGER DEFAULT 0,
	goal_id INTEGER,
	priority INTEGER DEFAULT 2,
	due_date DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	FOREIGN KEY(goal_id) REFERENCES goals(id) ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS chat_history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// --- Goals ---

func (s *Store) CreateGoal(g models.Goal) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO goals(title, description, category, progress, status, target_date) VALUES(?,?,?,?,?,?)`,
		g.Title, g.Description, g.Category, g.Progress, defaultStr(g.Status, "active"), g.TargetDate)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	_, _ = s.db.Exec(`INSERT INTO goal_progress_history(goal_id, progress) VALUES(?,?)`, id, g.Progress)
	return id, nil
}

func (s *Store) UpdateGoalProgress(id int64, progress int) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	status := "active"
	if progress >= 100 {
		status = "done"
	}
	_, err := s.db.Exec(`UPDATE goals SET progress=?, status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, progress, status, id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO goal_progress_history(goal_id, progress) VALUES(?,?)`, id, progress)
	return err
}

func (s *Store) UpdateGoal(g models.Goal) error {
	_, err := s.db.Exec(`UPDATE goals SET title=?, description=?, category=?, target_date=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		g.Title, g.Description, g.Category, g.TargetDate, g.ID)
	return err
}

func (s *Store) DeleteGoal(id int64) error {
	_, err := s.db.Exec(`DELETE FROM goals WHERE id=?`, id)
	return err
}

func (s *Store) ListGoals() ([]models.Goal, error) {
	rows, err := s.db.Query(`SELECT id, title, description, category, progress, status, target_date, created_at, updated_at FROM goals ORDER BY (status='done'), updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Goal
	for rows.Next() {
		var g models.Goal
		var td sql.NullTime
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.Category, &g.Progress, &g.Status, &td, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		if td.Valid {
			t := td.Time
			g.TargetDate = &t
		}
		out = append(out, g)
	}
	return out, nil
}

func (s *Store) GoalProgressHistory(goalID int64) ([]models.GoalProgressPoint, error) {
	rows, err := s.db.Query(`SELECT id, goal_id, progress, recorded_at FROM goal_progress_history WHERE goal_id=? ORDER BY recorded_at ASC`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.GoalProgressPoint
	for rows.Next() {
		var p models.GoalProgressPoint
		if err := rows.Scan(&p.ID, &p.GoalID, &p.Progress, &p.RecordedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// --- Mood ---

func (s *Store) LogMood(score int, note string) error {
	_, err := s.db.Exec(`INSERT INTO mood_logs(score, note) VALUES(?,?)`, score, note)
	return err
}

func (s *Store) RecentMoods(days int) ([]models.MoodLog, error) {
	rows, err := s.db.Query(`SELECT id, score, note, logged_at FROM mood_logs WHERE logged_at >= datetime('now', ?) ORDER BY logged_at ASC`, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.MoodLog
	for rows.Next() {
		var m models.MoodLog
		if err := rows.Scan(&m.ID, &m.Score, &m.Note, &m.LoggedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) TodayMood() (*models.MoodLog, error) {
	row := s.db.QueryRow(`SELECT id, score, note, logged_at FROM mood_logs WHERE date(logged_at)=date('now','localtime') ORDER BY logged_at DESC LIMIT 1`)
	var m models.MoodLog
	if err := row.Scan(&m.ID, &m.Score, &m.Note, &m.LoggedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// --- Journal ---

func (s *Store) UpsertJournal(date time.Time, content string) error {
	wc := wordCount(content)
	d := date.Format("2006-01-02")
	_, err := s.db.Exec(`INSERT INTO journal_entries(date, content, word_count) VALUES(?,?,?)
		ON CONFLICT(date) DO UPDATE SET content=excluded.content, word_count=excluded.word_count, updated_at=CURRENT_TIMESTAMP`, d, content, wc)
	return err
}

func (s *Store) ListJournal(limit int) ([]models.JournalEntry, error) {
	rows, err := s.db.Query(`SELECT id, date, content, word_count, created_at, updated_at FROM journal_entries ORDER BY date DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.JournalEntry
	for rows.Next() {
		var j models.JournalEntry
		if err := rows.Scan(&j.ID, &j.Date, &j.Content, &j.WordCount, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, nil
}

// --- Todos ---

func (s *Store) CreateTodo(t models.Todo) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO todos(text, done, goal_id, priority, due_date) VALUES(?,?,?,?,?)`,
		t.Text, boolToInt(t.Done), t.GoalID, defaultInt(t.Priority, 2), t.DueDate)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ToggleTodo(id int64) error {
	_, err := s.db.Exec(`UPDATE todos SET done = CASE done WHEN 1 THEN 0 ELSE 1 END,
		completed_at = CASE done WHEN 1 THEN NULL ELSE CURRENT_TIMESTAMP END WHERE id=?`, id)
	return err
}

func (s *Store) DeleteTodo(id int64) error {
	_, err := s.db.Exec(`DELETE FROM todos WHERE id=?`, id)
	return err
}

func (s *Store) SetTodoPriority(id int64, p int) error {
	_, err := s.db.Exec(`UPDATE todos SET priority=? WHERE id=?`, p, id)
	return err
}

func (s *Store) LinkTodo(id int64, goalID *int64) error {
	_, err := s.db.Exec(`UPDATE todos SET goal_id=? WHERE id=?`, goalID, id)
	return err
}

func (s *Store) ListTodos() ([]models.Todo, error) {
	rows, err := s.db.Query(`SELECT id, text, done, goal_id, priority, due_date, created_at, completed_at FROM todos ORDER BY done ASC, priority ASC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Todo
	for rows.Next() {
		var t models.Todo
		var gid sql.NullInt64
		var due, comp sql.NullTime
		var done int
		if err := rows.Scan(&t.ID, &t.Text, &done, &gid, &t.Priority, &due, &t.CreatedAt, &comp); err != nil {
			return nil, err
		}
		t.Done = done == 1
		if gid.Valid {
			v := gid.Int64
			t.GoalID = &v
		}
		if due.Valid {
			v := due.Time
			t.DueDate = &v
		}
		if comp.Valid {
			v := comp.Time
			t.CompletedAt = &v
		}
		out = append(out, t)
	}
	return out, nil
}

// --- Chat ---

func (s *Store) AppendChat(role, content string) error {
	_, err := s.db.Exec(`INSERT INTO chat_history(role, content) VALUES(?,?)`, role, content)
	return err
}

func (s *Store) ChatHistory(limit int) ([]models.ChatMessage, error) {
	rows, err := s.db.Query(`SELECT id, role, content, created_at FROM chat_history ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ChatMessage
	for rows.Next() {
		var m models.ChatMessage
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	// reverse
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// --- helpers ---

func (s *Store) Streak() (int, error) {
	rows, err := s.db.Query(`SELECT DISTINCT date(logged_at) FROM mood_logs ORDER BY date(logged_at) DESC`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	streak := 0
	day := time.Now()
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return 0, err
		}
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			return 0, err
		}
		if sameDay(t, day) {
			streak++
			day = day.AddDate(0, 0, -1)
		} else if t.Before(day) {
			break
		}
	}
	return streak, nil
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func defaultStr(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func defaultInt(i, d int) int {
	if i == 0 {
		return d
	}
	return i
}

func wordCount(s string) int {
	count, inWord := 0, false
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' {
			if inWord {
				count++
				inWord = false
			}
		} else {
			inWord = true
		}
	}
	if inWord {
		count++
	}
	return count
}
