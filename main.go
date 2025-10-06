package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type errorResp struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func dsn() string {
	host := getenv("DB_HOST", "migrate.cluster-c3o6qc26ias1.eu-west-1.rds.amazonaws.com")
	port := getenv("DB_PORT", "3306")
	user := getenv("DB_USER", "lina")
	pass := getenv("DB_PASS", "123456")
	name := getenv("DB_NAME", "ocpp_CSGO")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func ifNullFloat(v sql.NullFloat64) float64 {
	if v.Valid {
		return v.Float64
	}
	return 0
}

func handleRevenue(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	q := r.URL.Query()
	partnerID := q.Get("partner_id")
	if partnerID == "" {
		partnerID = q.Get("organization_id")
	}
	if partnerID == "" {
		writeJSON(w, http.StatusBadRequest, errorResp{Error: "missing partner_id or organization_id"})
		return
	}

	// الأعمدة الصحيحة: total_amount / issued_to / issued_date
	const sqlStmt = `
		SELECT COALESCE(SUM(total_amount), 0) AS total
		FROM Partner_Bill
		WHERE TRIM(issued_to) = TRIM(?)
		  AND issued_date >= CURDATE()
		  AND issued_date <  CURDATE() + INTERVAL 1 DAY
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var total sql.NullFloat64
	if err := db.QueryRowContext(ctx, sqlStmt, partnerID).Scan(&total); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResp{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"partner_id":    partnerID,
		"date":          time.Now().Format("2006-01-02"),
		"total_revenue": ifNullFloat(total),
	})
}

func handleActiveSessions(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	query := `
		SELECT session_id, start_date
		FROM Sessions
		WHERE active = 1 OR charging = 1
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResp{Error: err.Error()})
		return
	}
	defer rows.Close()

	type session struct {
		SessionID string     `json:"session_id"`
		StartedAt *time.Time `json:"started_at,omitempty"`
	}

	var sessions []session
	for rows.Next() {
		var s session
		if err := rows.Scan(&s.SessionID, &s.StartedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResp{Error: err.Error()})
			return
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResp{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":    len(sessions),
		"sessions": sessions,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	err := db.PingContext(ctx)

	dbStatus := "reachable"
	if err != nil {
		dbStatus = "unreachable"
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"db":     dbStatus,
	})
}

func main() {
	// DB pool
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("command") {
		case "revenue":
			handleRevenue(w, r, db)
		case "active_sessions":
			handleActiveSessions(w, r, db)
		case "health":
			handleHealth(w, r, db)
		default:
			writeJSON(w, http.StatusBadRequest, errorResp{Error: "unknown or missing command"})
		}
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
