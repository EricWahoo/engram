package store

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

func TestReadinessQueryAvoidsAggregateWork(t *testing.T) {
	s := newTestStore(t)
	if err := rejectAggregateBytecode(s.db, readinessSQL); err != nil {
		t.Fatal(err)
	}
	// A scalar aggregate inside EXISTS must still fail this guard: its query
	// plan has no temporary tree, but it scans every observation.
	mutated := strings.Replace(readinessSQL,
		"SELECT project, created_at, deleted_at FROM observations",
		"SELECT SUM(length(created_at)) FROM observations", 1)
	if mutated == readinessSQL || rejectAggregateBytecode(s.db, mutated) == nil {
		t.Fatal("aggregate mutation escaped readiness work guard")
	}
}

func rejectAggregateBytecode(db *sql.DB, query string) error {
	rows, err := db.Query("EXPLAIN " + query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var addr, p1, p2, p3, p5 int
		var opcode string
		var p4, comment sql.NullString
		if err := rows.Scan(&addr, &opcode, &p1, &p2, &p3, &p4, &p5, &comment); err != nil {
			return err
		}
		if strings.HasPrefix(opcode, "Agg") || strings.HasPrefix(opcode, "Sorter") || opcode == "OpenEphemeral" {
			return fmt.Errorf("readiness executes %s at bytecode address %d", opcode, addr)
		}
	}
	return rows.Err()
}
