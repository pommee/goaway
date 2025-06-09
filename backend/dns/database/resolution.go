package database

import (
	"database/sql"
	"errors"
	"fmt"
	"goaway/backend/dns/database/models"
	"strings"
)

func FetchResolutions(db *sql.DB) ([]models.Resolution, error) {
	query := "SELECT ip, domain FROM resolution"

	rows, err := db.Query(query)
	if err != nil {
		log.Error("Database query error: %v", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

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
	var foundDomain string
	domain = strings.TrimSuffix(domain, ".")

	query := "SELECT ip FROM resolution WHERE domain = ?"
	err := db.QueryRow(query, domain).Scan(&foundDomain)
	if err == nil {
		return foundDomain, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		wildcardDomain := "*." + strings.Join(parts[i:], ".")

		err = db.QueryRow(query, wildcardDomain).Scan(&foundDomain)
		if err == nil {
			return foundDomain, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}

	return "", nil
}

func CreateNewResolution(db *sql.DB, ip, domain string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Error("Could not start database transaction %v", err)
		return fmt.Errorf("could not create new resolution")
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
		return fmt.Errorf("could not create new resolution")
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	if _, err := stmt.Exec(ip, domain); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			log.Error("Domain already exists. Reason: %v", err)
			return fmt.Errorf("domain already exists, must me unique")
		}
		log.Error("Could not save resolution. Reason: %v", err)
		return fmt.Errorf("could not create new resolution")
	}

	return nil
}

func DeleteResolution(db *sql.DB, ip, domain string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("could not delete resolution due to db error: %v", err)
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			log.Warning("DB commit error: %v", err)
		}
	}()

	query := "DELETE FROM resolution WHERE ip = ? AND domain = ?"
	stmt, err := tx.Prepare(query)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("could not delete resolution due to db error: %v", err)
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	result, err := stmt.Exec(ip, domain)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("could not delete resolution. Reason: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("could not retrieve affected rows: %v", err)
	}

	if rowsAffected == 0 {
		log.Warning("No resolution found with IP: %s and Domain: %s", ip, domain)
	}

	return int(rowsAffected), nil
}
