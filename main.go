package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"log"
	"time"

	"github.com/altwine/go-mindustry-ping/pkg/serverinfo"
)

var (
	infoObjects = make(map[string]*serverinfo.ServerInfo)

	db *sql.DB

	pollInterval    = 60 * time.Second
	cleanupInterval = 60
	cycleCounter    = 0
)

func main() {
	enableEndpoint := flag.Bool("enable-rest-api", false, "Enable REST API endpoints")
	enableDiscordBot := flag.Bool("enable-discord-bot", false, "Enable Discord bot")

	flag.Parse()

	if !*enableDiscordBot && !*enableEndpoint {
		log.Printf("Nothing is launched, use --enable-rest-api or --enable-discord-bot", len(servers))
		return
	}

	initServers()
	log.Printf("loaded %d servers", len(servers))

	initFont()

	if err := initDB(); err != nil {
		log.Fatal("ошибка инициализации бд:", err)
	}
	defer db.Close()

	initInfoObjects()
	go pollLoop()
	if *enableEndpoint {
		go router()
	}
	if *enableDiscordBot {
		go initDiscordBot()
		defer destroyDiscordBot()
	}
	waitForShutdown()

}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./mindustry_stats.db?_journal_mode=WAL")
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS server_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_name TEXT NOT NULL,
		address TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		players INTEGER NOT NULL,
		wave INTEGER NOT NULL,
		ping INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON server_stats(timestamp);
	CREATE INDEX IF NOT EXISTS idx_server_address ON server_stats(server_name, address);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	log.Println("база данных создана")
	return nil
}

func initInfoObjects() {
	for _, srv := range servers {
		for _, addr := range srv.Address {
			key := addr.Host + ":" + strconv.Itoa(addr.Port)
			if _, exists := infoObjects[key]; !exists {
				infoObjects[key] = &serverinfo.ServerInfo{
					Address: addr.Host,
					Port:    addr.Port,
				}
			}
		}
	}
}

func pollLoop() {
	for {
		fetchAndSave()
		time.Sleep(pollInterval)

		cycleCounter++
		if cycleCounter >= cleanupInterval {
			cleanOldRecords()
			cycleCounter = 0
		}
	}
}

func fetchAndSave() {
	var wg sync.WaitGroup

	for key, si := range infoObjects {
		wg.Add(1)
		go func(k string, info *serverinfo.ServerInfo) {
			defer wg.Done()

			err := info.Update()
			record := HistoryRecord{
				Timestamp: time.Now().Unix(),
			}
			if err != nil {
				log.Printf("ошибка обновления (%s): %v", k, err)
				record.Players = -1
				record.Wave = -1
				record.Ping = -1
			} else {
				record.Players = info.Players
				record.Wave = info.Waves
				record.Ping = info.Latency
				log.Printf("%s: %d игроков, волна %d", k, info.Players, info.Waves)
			}

			serverName := findServerNameByAddress(k)
			if serverName == "" {
				log.Printf("не найден сервер по адресу %s", k)
				return
			}

			err = insertRecord(serverName, k, record)
			if err != nil {
				log.Printf("ошибка записи в БД (%s): %v", k, err)
			}
		}(key, si)
	}

	wg.Wait()
}

func findServerNameByAddress(addrKey string) string {
	for _, srv := range servers {
		for _, addr := range srv.Address {
			if addr.Host+":"+strconv.Itoa(addr.Port) == addrKey {
				return srv.Name
			}
		}
	}
	return ""
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("получен сигнал завершения, закрываю бдху")
	if err := db.Close(); err != nil {
		log.Printf("ошибка при закрытии БД: %v", err)
	}
	os.Exit(0)
}
