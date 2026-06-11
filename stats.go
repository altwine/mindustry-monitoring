package main

import (
	"database/sql"
	"log"
	"time"
)

const (
	MaxAgeHours = 12
)

type HistoryRecord struct {
	Timestamp int64
	Players   int
	Wave      int
	Ping      int
}

func insertRecord(serverName, address string, rec HistoryRecord) error {
	_, err := db.Exec(
		"INSERT INTO server_stats(server_name, address, timestamp, players, wave, ping) VALUES(?, ?, ?, ?, ?, ?)",
		serverName, address, rec.Timestamp, rec.Players, rec.Wave, rec.Ping,
	)
	return err
}

func cleanOldRecords() {
	cutoff := time.Now().Unix() - int64(MaxAgeHours*3600)
	_, err := db.Exec("DELETE FROM server_stats WHERE timestamp < ?", cutoff)
	if err != nil {
		log.Printf("ошибка очистки старых записей: %v", err)
		return
	}
}

func getStatsByAddress(address string, hours int) ([]HistoryRecord, error) {
	var rows *sql.Rows
	var err error

	if hours > 0 {
		cutoff := time.Now().Unix() - int64(hours*3600)
		rows, err = db.Query(
			"SELECT timestamp, players, wave, ping FROM server_stats WHERE address = ? AND timestamp >= ? ORDER BY timestamp ASC",
			address, cutoff,
		)
	} else {
		rows, err = db.Query(
			"SELECT timestamp, players, wave, ping FROM server_stats WHERE address = ? ORDER BY timestamp ASC",
			address,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []HistoryRecord
	for rows.Next() {
		var rec HistoryRecord
		if err := rows.Scan(&rec.Timestamp, &rec.Players, &rec.Wave, &rec.Ping); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}
