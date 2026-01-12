package filter

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/nkn/unifi-cli/internal/api"
)

// Filter applies SQL WHERE clause to clients using JSON storage
type Filter struct {
	db          *sql.DB
	whereClause string
}

// NewFilter creates in-memory SQLite database and returns filter
func NewFilter(whereClause string) (*Filter, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table and view
	if _, err := db.Exec(clientTableSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Filter{db: db, whereClause: whereClause}, nil
}

// Apply filters clients using SQL WHERE clause
func (f *Filter) Apply(clients []api.Client) ([]api.Client, error) {
	// Insert clients as JSON
	if err := f.insertClients(clients); err != nil {
		return nil, err
	}

	// Query with WHERE clause
	return f.queryClients()
}

// insertClients inserts all clients as JSON into the database
func (f *Filter) insertClients(clients []api.Client) error {
	tx, err := f.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO clients (data) VALUES (?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, client := range clients {
		jsonData, err := json.Marshal(client)
		if err != nil {
			return fmt.Errorf("failed to marshal client: %w", err)
		}

		if _, err := stmt.Exec(string(jsonData)); err != nil {
			return fmt.Errorf("failed to insert client: %w", err)
		}
	}

	return tx.Commit()
}

// queryClients executes SELECT with WHERE clause on the view
func (f *Filter) queryClients() ([]api.Client, error) {
	query := fmt.Sprintf("SELECT data FROM clients_view WHERE %s", f.whereClause)

	rows, err := f.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var result []api.Client
	for rows.Next() {
		var jsonData string
		if err := rows.Scan(&jsonData); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var client api.Client
		if err := json.Unmarshal([]byte(jsonData), &client); err != nil {
			return nil, fmt.Errorf("failed to unmarshal client: %w", err)
		}

		result = append(result, client)
	}

	return result, rows.Err()
}

// Close cleans up database connection
func (f *Filter) Close() error {
	return f.db.Close()
}
