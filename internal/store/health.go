package store

const readinessSQL = `SELECT
		EXISTS(SELECT 1 FROM sessions),
		EXISTS(SELECT project, created_at, deleted_at FROM observations),
		EXISTS(SELECT 1 FROM user_prompts)
	`

// CheckReadiness checks that the core tables are readable through the existing
// generation-fenced connection. Empty tables are healthy. This bounded read
// does not calculate statistics or perform a full integrity or write audit.
func (s *Store) CheckReadiness() error {
	var sessions, observations, prompts bool
	return s.db.QueryRow(readinessSQL).Scan(&sessions, &observations, &prompts)
}
