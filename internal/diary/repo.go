package diary

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/andre-cmd-rgb/bit-tracker/internal/db"
)

type Repo struct{ s *db.Store }

func NewRepo(s *db.Store) *Repo { return &Repo{s: s} }

var ErrNotFound = errors.New("entry not found")

func scanEntry(row interface {
	Scan(dest ...any) error
}) (Entry, error) {
	var e Entry
	var dateStr, startedStr, updatedStr, tagsStr string
	var projCompleted, isBad, isGood int
	err := row.Scan(
		&e.ID, &dateStr, &startedStr, &updatedStr,
		&e.RawText, &e.CleanedText,
		&e.Mood, &e.StudyMinutes, &e.ScrollMinutes, &e.ProjectMinutes,
		&tagsStr, &e.ProjectName, &e.ProjectNote, &projCompleted,
		&isBad, &isGood,
	)
	if err != nil {
		return Entry{}, err
	}
	e.Date, _ = time.Parse(DateLayout, dateStr)
	e.StartedAt, _ = time.Parse(time.RFC3339, startedStr)
	e.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	e.Tags = SplitTags(tagsStr)
	e.ProjectCompleted = projCompleted == 1
	e.IsBadDay = isBad == 1
	e.IsGoodDay = isGood == 1
	return e, nil
}

const selectCols = `id, entry_date, started_at, updated_at, raw_text, cleaned_text,
	mood, study_minutes, scroll_minutes, project_minutes, tags, project_name,
	project_note, project_completed, is_bad_day, is_good_day`

// GetByDate returns the entry for the given date or ErrNotFound.
func (r *Repo) GetByDate(date time.Time) (Entry, error) {
	d := date.Format(DateLayout)
	row := r.s.DB.QueryRow(`SELECT `+selectCols+` FROM entries WHERE entry_date = ?`, d)
	e, err := scanEntry(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Entry{}, ErrNotFound
	}
	return e, err
}

// GetOrCreateToday returns today's entry, creating it with intro text if needed.
func (r *Repo) GetOrCreateToday(now time.Time) (Entry, bool, error) {
	e, err := r.GetByDate(now)
	if err == nil {
		return e, false, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Entry{}, false, err
	}
	e = Entry{
		Date:      now,
		StartedAt: now,
		UpdatedAt: now,
		RawText:   Intro(now),
	}
	if err := r.Insert(&e); err != nil {
		return Entry{}, false, err
	}
	return e, true, nil
}

func (r *Repo) Insert(e *Entry) error {
	isBad, isGood := Score(*e)
	e.IsBadDay, e.IsGoodDay = isBad, isGood
	res, err := r.s.DB.Exec(
		`INSERT INTO entries(entry_date, started_at, updated_at, raw_text, cleaned_text,
			mood, study_minutes, scroll_minutes, project_minutes, tags, project_name,
			project_note, project_completed, is_bad_day, is_good_day)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.Date.Format(DateLayout),
		e.StartedAt.Format(time.RFC3339),
		e.UpdatedAt.Format(time.RFC3339),
		e.RawText, e.CleanedText,
		e.Mood, e.StudyMinutes, e.ScrollMinutes, e.ProjectMinutes,
		JoinTags(e.Tags), e.ProjectName, e.ProjectNote, boolInt(e.ProjectCompleted),
		boolInt(e.IsBadDay), boolInt(e.IsGoodDay),
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		e.ID = id
	}
	return nil
}

func (r *Repo) Save(e *Entry) error {
	e.UpdatedAt = time.Now()
	isBad, isGood := Score(*e)
	e.IsBadDay, e.IsGoodDay = isBad, isGood
	_, err := r.s.DB.Exec(
		`UPDATE entries SET updated_at=?, raw_text=?, cleaned_text=?, mood=?,
			study_minutes=?, scroll_minutes=?, project_minutes=?, tags=?,
			project_name=?, project_note=?, project_completed=?, is_bad_day=?, is_good_day=?
		 WHERE id=?`,
		e.UpdatedAt.Format(time.RFC3339),
		e.RawText, e.CleanedText,
		e.Mood, e.StudyMinutes, e.ScrollMinutes, e.ProjectMinutes,
		JoinTags(e.Tags), e.ProjectName, e.ProjectNote, boolInt(e.ProjectCompleted),
		boolInt(e.IsBadDay), boolInt(e.IsGoodDay),
		e.ID,
	)
	return err
}

// List returns entries ordered ascending by date within [from, to] inclusive.
// Zero times mean unbounded.
func (r *Repo) List(from, to time.Time) ([]Entry, error) {
	query := `SELECT ` + selectCols + ` FROM entries`
	var args []any
	var conds []string
	if !from.IsZero() {
		conds = append(conds, `entry_date >= ?`)
		args = append(args, from.Format(DateLayout))
	}
	if !to.IsZero() {
		conds = append(conds, `entry_date <= ?`)
		args = append(args, to.Format(DateLayout))
	}
	if len(conds) > 0 {
		query += ` WHERE ` + strings.Join(conds, ` AND `)
	}
	query += ` ORDER BY entry_date ASC`
	rows, err := r.s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Search matches keyword across raw_text, cleaned_text, tags, project_name.
// Tag filter is applied if non-empty.
func (r *Repo) Search(keyword, tag string) ([]Entry, error) {
	query := `SELECT ` + selectCols + ` FROM entries WHERE 1=1`
	var args []any
	if keyword != "" {
		q := "%" + keyword + "%"
		query += ` AND (raw_text LIKE ? OR cleaned_text LIKE ? OR project_name LIKE ? OR project_note LIKE ?)`
		args = append(args, q, q, q, q)
	}
	if tag != "" {
		query += ` AND tags LIKE ?`
		args = append(args, "%"+tag+"%")
	}
	query += ` ORDER BY entry_date DESC`
	rows, err := r.s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// LogAIMessage records a chat/AI interaction.
func (r *Repo) LogAIMessage(entryID int64, mode, prompt, response string) error {
	var id any
	if entryID > 0 {
		id = entryID
	}
	_, err := r.s.DB.Exec(
		`INSERT INTO ai_messages(entry_id, mode, prompt, response, created_at) VALUES(?,?,?,?,?)`,
		id, mode, prompt, response, time.Now().Format(time.RFC3339),
	)
	return err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
