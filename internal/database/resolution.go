package database

import (
	"database/sql"
	"goaway/internal/database/models"
	"strings"
)

func FetchResolutions(db *sql.DB) ([]models.Resolution, error) {
	query := ("SELECT ip, domain FROM resolution")

	rows, err := db.Query(query)
	if err != nil {
		log.Error("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var resolutions []models.Resolution
	for rows.Next() {
		var resolution models.Resolution

		if err := rows.Scan(
			&resolution.IP, &resolution.Domain,
		); err != nil {
			return nil, err
		}

		resolutions = append(resolutions, resolution)
	}

	return resolutions, nil
}

func FetchResolution(db *sql.DB, domain string) (string, error) {
	query := "SELECT ip FROM resolution WHERE domain = ?"
	domain = strings.TrimSuffix(domain, ".")

	var foundDomain string
	err := db.QueryRow(query, domain).Scan(&foundDomain)
	if err != nil {
		return "", err
	}

	return foundDomain, nil
}

func CreateNewResolution(db *sql.DB, ip, domain string) {
	tx, err := db.Begin()
	if err != nil {
		log.Error("Could not start database transaction %v", err)
		return
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			log.Warning("DB commit error %v", err)
		}
	}()

	query := "INSERT INTO resolution (ip, domain) VALUES (?, ?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Error("Could not create a prepared statement for resolution %v", err)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(ip, domain); err != nil {
		log.Error("Could not save resolution. Reason: %v", err)
		return
	}
}
