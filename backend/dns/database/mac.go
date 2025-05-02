package database

import (
	"database/sql"
	"errors"
)

func FindVendor(db *sql.DB, mac string) (string, error) {
	var vendor string
	err := db.QueryRow("SELECT vendor FROM mac_addresses WHERE mac = ?", mac).Scan(&vendor)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return vendor, nil
}

func SaveMacEntry(db *sql.DB, clientIP, mac, vendor string) {
	query := "INSERT INTO mac_addresses (ip, mac, vendor) VALUES (?, ?, ?)"
	_, err := db.Exec(query, clientIP, mac, vendor)

	log.Warning("Unable to save new MAC entry %v", err)
}
