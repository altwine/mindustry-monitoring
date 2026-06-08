package main

import (
	"database/sql"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/altwine/go-mindustry-ping/pkg/serverinfo"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
)

const (
	MaxAgeHours = 12
)

type HistoryRecord struct {
	Timestamp int64
	Players   int
	Wave      int
}

type Address struct {
	Host string
	Port int
}

type Server struct {
	Name    string
	Address []Address
}

var servers = []Server{
	{"Mindustry Trio", []Address{{"209.25.141.19", 1185}, {"209.25.141.19", 1184}}},
	{"DontVin", []Address{{"167.172.89.177", 6567}}},
	{"Everyone Create", []Address{{"9666.fun", 11542}}},
	{"BE4HOCTb", []Address{{"play.ve4ka.space", 6567}}},
	{"RawHex", []Address{{"104.128.134.201", 6567}}},
	{"汐洛界域", []Address{{"mdt-game.silo-clouds.cn", 6567}}},
	{"EscoDrama", []Address{
		{"2.26.63.50", 6567}, {"2.26.63.50", 6568},
		{"46.149.69.119", 10020}, {"46.149.69.119", 10029},
		{"46.149.69.119", 10030}, {"46.149.69.119", 10031},
		{"46.149.69.119", 10037}, {"play.larzed.icu", 6569},
		{"play.larzed.icu", 6570},
	}},
	{"M-ind Fans", []Address{{"47.236.249.154", 16567}}},
	{"Maxdustry", []Address{
		{"78.24.219.168", 1001},
		{"78.24.219.168", 1002},
		{"78.24.219.168", 1003},
		{"78.24.219.168", 1004},
	}},
	{"gods field", []Address{{"mdt.shenyugame.cn", 20524}}},
	{"Abyss Icefire", []Address{{"9666.fun", 18564}}},
	{"DELTA", []Address{{"45.61.49.88", 2500}, {"45.61.49.88", 7500}}},
	{"MEIQIU", []Address{{"mindustry.fun", 6567}, {"t1.mindustry.fun", 6567}, {"t2.mindustry.fun", 6567}}},
	{"Mindustry Tool", []Address{
		{"server.mindustry-tool.com", 10000},
		{"server.mindustry-tool.com", 10001},
		{"server.mindustry-tool.com", 10002},
		{"server.mindustry-tool.com", 10003},
		{"server.mindustry-tool.com", 10004},
		{"server.mindustry-tool.com", 10005},
		{"server.mindustry-tool.com", 10006},
		{"server.mindustry-tool.com", 10007},
		{"server.mindustry-tool.com", 10008},
		{"server.mindustry-tool.com", 10009},
		{"server.mindustry-tool.com", 10010},
	}},
	{"IndustryCats", []Address{{"80.74.28.57", 55072}, {"80.74.28.57", 7623}}},
	{"LazyCats", []Address{{"pivo.pivomind.pro", 1000}, {"pivo.pivomind.pro", 1001}, {"pivo.pivomind.pro", 1002}, {"pivo.pivomind.pro", 1003}}},
	{"homecloud", []Address{{"m163.pivads.cfd", 6567}}},
	{"Tendhost", []Address{{"tendhost.ru", 7577}}},
	{"OmniCorp", []Address{
		{"omnidustry.ru", 6567},
		{"omnidustry.ru", 6568},
		{"omnidustry.ru", 6569},
		{"omnidustry.ru", 6571},
		{"omnidustry.ru", 6572},
		{"omnidustry.ru", 6573},
		{"omnidustry.ru", 6574},
		{"omnidustry.ru", 6575},
		{"omnidustry.ru", 6576},
		{"omnidustry.ru", 6577},
	}},
	{"EscoCorp Servers", []Address{{"95.215.56.128", 6574}}},
	{"Erepulo", []Address{
		{"95.84.198.97", 5411},
		{"95.84.198.97", 5412},
		{"95.84.198.97", 5413},
		{"95.84.198.97", 5414},
		{"95.84.198.97", 5415},
	}},
	{"404ru", []Address{
		{"144.31.2.197", 6554},
		{"144.31.2.197", 6555},
		{"185.250.46.71", 6587},
		{"185.250.46.71", 6588},
		{"185.250.46.71", 6589},
		{"185.250.46.71", 6590},
		{"185.250.46.71", 6591},
		{"185.250.46.71", 6592},
		{"185.250.46.71", 6593},
	}},
	{"ArmyOFUkraine", []Address{
		{"202.181.188.253", 27510},
		{"202.181.188.253", 27511},
		{"202.181.188.253", 27512},
		{"202.181.188.253", 27513},
	}},
	{"Snow", []Address{{"46.23.90.167", 6567}}},
	{"routerchain", []Address{{"v8.baseduser.eu.org", 6567}, {"v8.baseduser.eu.org", 6568}}},
	{"TinyLake", []Address{{"cn.mindustry.top", 6567}, {"cn.mindustry.top", 40542}}},
	{"Mindurka", []Address{
		{"147.45.230.117", 3050},
		{"147.45.230.117", 3051},
		{"147.45.230.117", 3052},
		{"147.45.230.117", 3053},
		{"147.45.230.117", 3054},
		{"147.45.230.117", 3055},
		{"147.45.230.117", 3056},
	}},
	{"Gadgetroch", []Address{
		{"mindustry.gadgetroch.com", 6567},
		{"mindustry.gadgetroch.com", 6568},
		{"mindustry.gadgetroch.com", 6569},
	}},
	{"XCore", []Address{
		{"62.30.47.117", 7001},
		{"62.30.47.117", 7002},
		{"62.30.47.117", 7003},
		{"62.30.47.117", 7004},
		{"62.30.47.117", 7005},
		{"62.30.47.117", 7006},
		{"62.30.47.117", 7007},
		{"62.30.47.117", 7008},
		{"62.30.47.117", 7009},
	}},
	{"Eradication Mindustry", []Address{
		{"130.61.22.183", 8000},
		{"130.61.22.183", 8001},
		{"144.24.196.119", 8000},
		{"144.24.196.119", 8001},
		{"140.238.246.78", 8000},
		{"140.238.246.78", 8001},
		{"62.30.47.116", 8000},
		{"62.30.47.116", 8001},
		{"122.180.249.217", 8000},
		{"129.80.53.21", 8000},
		{"82.114.229.170", 8000},
	}},
	{"Chaotic Neutral", []Address{
		{"n4.xpdustry.com", 50010},
		{"n4.xpdustry.com", 50011},
		{"n4.xpdustry.com", 50012},
		{"n4.xpdustry.com", 50013},
		{"n4.xpdustry.com", 50014},
		{"n4.xpdustry.com", 50015},
		{"n4.xpdustry.com", 50016},
		{"n4.xpdustry.com", 50018},
		{"n4.xpdustry.com", 50019},
	}},
	{"Featured Servers", []Address{{"n4.xpdustry.com", 50022}}},
	{"MeowIsland", []Address{
		{"meowisland.ru", 6000},
		{"meowisland.ru", 6001},
		{"meowisland.ru", 6002},
		{"meowisland.ru", 6003},
	}},
	{"The Devil", []Address{{"new.xem8k5.top", 6567}}},
	{"Charophyceae", []Address{{"46.4.114.111", 6567}, {"ns.charo.qzz.io", 25029}}},
	{"MAL", []Address{{"78.47.238.87", 6567}, {"78.47.238.87", 6568}, {"78.47.238.87", 6569}}},
	{"io", []Address{{"148.251.184.58", 6567}, {"148.251.184.58", 1000}, {"148.251.184.58", 2000}, {"148.251.184.58", 3000}, {"148.251.184.58", 4000}}},
	{"SynapseOS", []Address{{"46.149.69.116", 5555}, {"185.137.233.193", 25583}, {"185.137.233.193", 25985}}},
	{"Spark", []Address{{"mindustry.net.cn", 6567}, {"new.mindustry.net.cn", 16666}, {"sub.mindustry.net.cn", 16567}}},
	{"Lett's Server Network", []Address{{"sandbox.lettsn.org", 6567}}},
	{"Pure PVP", []Address{{"mindustry.purepvp.org", 6567}, {"open.purepvp.org", 40010}, {"proxy.purepvp.org", 40010}}},
	{"Apricot Alliance", []Address{{"apricotalliance.org", 6567}}},
	{"Novice", []Address{{"play.simpfun.cn", 37144}, {"8.145.45.252", 6567}, {"play.simpfun.cn", 13558}}},
	{"Aetherial", []Address{{"aetherial.my-craft.cc", 25582}, {"aetherial.my-craft.cc", 25630}, {"aetherial.my-craft.cc", 25646}}},
	{"Rosemoncorp Servers", []Address{{"194.164.245.234", 6567}, {"194.164.245.234", 6568}}},
	{"grass server", []Address{{"mindustry.org.cn", 6567}}},
	{"Fish", []Address{
		{"162.248.101.95", 6567},
		{"162.248.100.98", 6567},
		{"162.248.102.101", 6567},
		{"162.248.101.53", 6567},
		{"162.248.100.133", 6567},
		{"162.248.101.116", 6567},
	}},
	{"Omnidust Servers", []Address{{"147.185.221.31", 21206}}},
	{"Router Pi", []Address{{"a55c81b7c4d6e6d604651a93f8af5cd83.asuscomm.com", 6567}, {"a55c81b7c4d6e6d604651a93f8af5cd83.asuscomm.com", 6568}}},
	{"TWS", []Address{{"tws.mlokis.dev", 6567}, {"tws.mlokis.dev", 9999}}},
	{"Classic Lemon", []Address{{"lemon.mindustry.icu", 6567}, {"ksang.mindustry.icu", 6567}, {"cn.mindustry.asia", 6567}}},
	{"Dim.ST", []Address{{"play.simpfun.cn", 14648}}},
	{"abcxyz vietnam offical", []Address{
		{"15.235.173.7", 14658},
		{"15.235.173.7", 15350},
		{"15.235.173.7", 15323},
		{"15.235.173.7", 15356},
	}},
	{"Industrial", []Address{{"95.215.56.61", 6567}, {"95.215.56.61", 6568}}},
	{"Power line", []Address{{"43.136.177.37", 6567}, {"43.136.177.37", 6568}}},
	{"ThePullerCell's Server Network", []Address{
		{"thepullercell.pp.ua", 7100},
		{"thepullercell.pp.ua", 7300},
		{"thepullercell.pp.ua", 7000},
		{"thepullercell.pp.ua", 7200},
	}},
	{"MDN", []Address{
		{"mindustry.ddns.net", 1000},
		{"mindustry.ddns.net", 2000},
		{"mindustry.ddns.net", 3000},
		{"mindustry.ddns.net", 4000},
		{"mindustry.ddns.net", 5000},
	}},
	{"Cyandustry", []Address{{"185.56.162.20", 4000}}},
	{"Alex Multiverse", []Address{
		{"alexmindustryv7.servegame.com", 25588},
		{"172.245.187.143", 6869},
		{"172.234.80.96", 6768},
		{"139.162.41.78", 6767},
		{"172.245.187.143", 6868},
		{"92.119.127.171", 6888},
	}},
	{"STNG", []Address{{"mindustry.asia", 6567}, {"new.mindustry.asia", 6567}}},
	{"WPAS", []Address{{"15.235.173.7", 30866}, {"15.235.173.7", 17662}, {"15.235.173.7", 9904}}},
	{"Moon", []Address{{"15.235.160.173", 26010}, {"srv49.godlike.club", 26010}}},
	{"Korea", []Address{{"server.mindustry.kr", 6567}}},
	{"起始物语", []Address{
		{"monthzifang.top", 6567},
		{"monthzifang.top", 6568},
		{"monthzifang.top", 6569},
		{"monthzifang.top", 6570},
	}},
	{"pvp", []Address{{"mindustry.cc", 6567}}},
	{"Directory", []Address{{"newsletter-focuses.gl.at.ply.gg", 1752}}},
	{"EV's ranked PvP server", []Address{{"evrankedpvp.ddns.net", 6567}}},
	{"Conservatory", []Address{{"165.232.152.103", 6567}, {"165.232.152.103", 6568}}},
	{"Blyss Mindustry", []Address{{"31.58.77.143", 6567}}},
	{"thedimas", []Address{
		{"77.42.41.140", 6567},
		{"77.42.41.140", 6503},
		{"77.42.41.140", 6505},
		{"77.42.41.140", 6507},
		{"77.42.41.140", 6508},
		{"77.42.41.140", 6509},
		{"77.42.41.140", 6510},
		{"77.42.41.140", 6511},
		{"77.42.41.140", 6513},
	}},
	{"Sakura", []Address{
		{"162.43.36.78", 24527},
		{"162.43.36.78", 25527},
		{"162.43.36.78", 27527},
		{"162.43.36.78", 28527},
	}},
	{"Foundation PvP", []Address{{"188.126.61.232", 6567}}},
	{"Mindustry TOKYO", []Address{
		{"mindustry-tokyo.xvps.jp", 6567},
		{"mindustry-tokyo.xvps.jp", 6568},
		{"mindustry-tokyo.xvps.jp", 6569},
	}},
	{"Lost island", []Address{{"zxs.squi2rel.top", 6567}}},
	{"MDT DO", []Address{{"43.248.117.226", 40149}}},
	{"Sectorized", []Address{{"sectorized.freeddns.org", 6567}}},
}

var (
	infoObjects = make(map[string]*serverinfo.ServerInfo)

	db *sql.DB

	pollInterval    = 60 * time.Second
	cleanupInterval = 60
	cycleCounter    = 0
)

func main() {
	go router()

	if err := initDB(); err != nil {
		log.Fatal("ошибка инициализации бд:", err)
	}
	defer db.Close()

	initInfoObjects()
	go pollLoop()
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
		wave INTEGER NOT NULL
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

func insertRecord(serverName, address string, rec HistoryRecord) error {
	_, err := db.Exec(
		"INSERT INTO server_stats(server_name, address, timestamp, players, wave) VALUES(?, ?, ?, ?, ?)",
		serverName, address, rec.Timestamp, rec.Players, rec.Wave,
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
			"SELECT timestamp, players, wave FROM server_stats WHERE address = ? AND timestamp >= ? ORDER BY timestamp ASC",
			address, cutoff,
		)
	} else {
		rows, err = db.Query(
			"SELECT timestamp, players, wave FROM server_stats WHERE address = ? ORDER BY timestamp ASC",
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
		if err := rows.Scan(&rec.Timestamp, &rec.Players, &rec.Wave); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func initInfoObjects() {
	for _, srv := range servers {
		for _, addr := range srv.Address {
			key := addr.Host + ":" + itoa(addr.Port)
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
				record.Wave = 0
			} else {
				record.Players = info.Players
				record.Wave = info.Waves
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
			if addr.Host+":"+itoa(addr.Port) == addrKey {
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

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	abs := n
	if n < 0 {
		abs = -n
	}
	for abs > 0 {
		digits = append([]byte{byte('0' + abs%10)}, digits...)
		abs /= 10
	}
	if n < 0 {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

const (
	width      = 1200
	height     = 1200
	cardWidth  = 1100
	cardHeight = 1100
)

const (
	mainBgColor = "#0F0F0F"
	// cardBgColor      = "#1E1E1E"
	// cardOutlineColor = "#00A896"
	textPrimary    = "#FFFFFF"
	accentColor    = "#FF6B6B"
	secondaryColor = "#4CAF7A"
	lineColor      = "#66CCFF"
	axisColor      = "#3A3A3A"
	labelColor     = "#AAAAAA"
)

const icons = "\uE80A\u26A0\uE800\uE801\uE802\uE803\uE804\uE805\uE806\uE807\uE808\uE809\uE80B\uE80D\uE80E\uE80F\uE810\uE811\uE812\uE813\uE814\uE815\uE816\uE817\uE818\uE819\uE81A\uE81B\uE81C\uE81D\uE81E\uE822\uE823\uE824\uE825\uE826\uE827\uE829\uE82A\uE82B\uE82C\uE82D\uE830\uE833\uE834\uE835\uE836\uE837\uE839\uE83A\uE83B\uE83D\uE83E\uE83F\uE842\uE844\uE845\uE84C\uE84D\uE852\uE853\uE85B\uE85C\uE85D\uE85E\uE85F\uE861\uE864\uE865\uE867\uE868\uE869\uE86B\uE86C\uE86D\uE86E\uE86F\uE870\uE871\uE872\uE873\uE874\uE875\uE876\uE877\uE878\uE879\uE87B\uE87C\uE88A\uE88B\uE88C\uE88D\uE88E\uE88F\uF029\uF0B0\uF0F6\uF120\uF129\uF12D\uF15B\uF15C\uF181\uF1C5\uF281\uF300\uF308"

// todo(altwine): refactor this lol
// assumes that current font is already set and size if `fontSize`
func DrawMindustryFormatString(dc *gg.Context, text string, x, y float64, fontSize int) {
	textArr := []rune(text)
	xSum := 0.0
	isBuildingColor := false
	stepByStepColor := ""
	for i := 0; i < len(textArr); i += 1 {
		r := string(textArr[i])
		if r == "[" {
			stepByStepColor = ""
			isBuildingColor = true
			continue
		}
		if r == "]" {
			isBuildingColor = false
			if stepByStepColor == "" {
				dc.SetHexColor(textPrimary)
				continue
			}
			if stepByStepColor[0] == '#' && len(stepByStepColor) <= 7 {
				dc.SetHexColor(stepByStepColor + strings.Repeat("0", 7-len(stepByStepColor)))
				continue
			} else {
				mc, is_valid := serverinfo.MINDUSTRY_COLORS[stepByStepColor]
				if !is_valid {
					log.Printf("lol invalid color: %v (%s), just skippin", mc, stepByStepColor)
					continue
				}
				dc.SetColor(color.RGBA(color.RGBA{
					R: uint8(mc.R),
					G: uint8(mc.G),
					B: uint8(mc.B),
					A: uint8(mc.A),
				}))
				continue
			}
		}
		if isBuildingColor {
			stepByStepColor += r
			continue
		}

		dc.DrawString(r, x+xSum, y)
		w, _ := dc.MeasureString(r)
		xSum += w
	}
}

func measureMindustryString(dc *gg.Context, text string, fontSize int) float64 {
	textArr := []rune(text)
	isBuildingColor := false
	xSum := 0.0
	for i := 0; i < len(textArr); i += 1 {
		r := string(textArr[i])
		if r == "[" {
			isBuildingColor = true
			continue
		}
		if r == "]" {
			isBuildingColor = false
			continue
		}
		if isBuildingColor {
			continue
		}
		w, _ := dc.MeasureString(r)
		xSum += w
	}
	return xSum
}

func genImage(dc *gg.Context, si serverinfo.ServerInfo, hr []HistoryRecord) {
	dc.SetHexColor(mainBgColor)
	dc.Clear()
	cardX := float64(width-cardWidth) / 2
	cardY := float64(height-cardHeight) / 2
	// dc.SetHexColor(cardBgColor)
	// dc.DrawRoundedRectangle(cardX, cardY, cardWidth, cardHeight, 20)
	// dc.Fill()
	// dc.SetHexColor(cardOutlineColor)
	// dc.SetLineWidth(3)
	// dc.DrawRoundedRectangle(cardX, cardY, cardWidth, cardHeight, 20)
	// dc.Stroke()

	const yAxisWidth = 65.0
	const rightMargin = 0.0
	const rightMarginText = 20.0*2 + 23.0

	xMinCard := cardX
	xMaxCard := cardX + cardWidth
	xMinGraph := xMinCard + yAxisWidth
	xMaxGraph := xMaxCard - rightMargin

	d := 200.0
	yTop := 120.0 + d
	yBottom := 700.0 + d
	chartHeight := yBottom - yTop

	maxPlayers := 0
	validCount := 0
	for _, r := range hr {
		if r.Players > maxPlayers {
			maxPlayers = r.Players
		}
		if r.Players != -1 {
			validCount++
		}
	}
	dataMax := float64(((maxPlayers + 4) / 5) * 5)
	if dataMax == 0 {
		dataMax = 5
	}

	mapY := func(players int) float64 {
		val := float64(players)
		if val < 0 {
			val = 0
		}
		if val > dataMax {
			val = dataMax
		}
		ratio := (dataMax - val) / dataMax
		return yTop + ratio*chartHeight
	}

	axisX := xMinGraph - 10
	loadFont(dc, 24)
	for v := 0.0; v <= dataMax; v += 5 {
		y := yTop + (dataMax-v)/dataMax*chartHeight
		label := fmt.Sprintf("%.0f", v)
		w, h := dc.MeasureString(label)

		dc.SetHexColor(axisColor)
		dc.SetLineWidth(1)
		dc.DrawLine(axisX, y, cardX+cardWidth, y)
		dc.Stroke()

		dc.SetHexColor(labelColor)
		dc.DrawString(label, axisX-5-w, y+h/3)
	}

	dc.SetLineWidth(2)
	dc.SetHexColor(lineColor)
	for i := 0; i < len(hr)-1; i++ {
		if hr[i].Players == -1 || hr[i+1].Players == -1 {
			continue
		}
		x1 := xMinGraph + (float64(i)/float64(len(hr)-1))*(xMaxGraph-xMinGraph)
		y1 := mapY(hr[i].Players)
		x2 := xMinGraph + (float64(i+1)/float64(len(hr)-1))*(xMaxGraph-xMinGraph)
		y2 := mapY(hr[i+1].Players)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}

	const maxTimeLabels = 12
	if len(hr) > 0 {
		loadFont(dc, 24)
		dc.SetHexColor(labelColor)
		for i := 0; i < maxTimeLabels; i++ {
			t := float64(i) / float64(maxTimeLabels-1)
			idx := int(math.Round(t * float64(len(hr)-1)))
			record := hr[idx]
			timeStr := time.Unix(record.Timestamp, 0).Format("15:04")
			w, _ := dc.MeasureString(timeStr)
			x := xMinGraph + t*(xMaxGraph-xMinGraph-rightMarginText)
			dc.DrawString(timeStr, x-w/2, yBottom+40)
		}
	}

	dc.SetHexColor(textPrimary)
	currSize := 128
	yDec := 0.0
	for {
		loadFont(dc, currSize)
		strWidth := measureMindustryString(dc, si.Host, currSize)
		if strWidth+cardX > cardWidth {
			currSize -= 8
			yDec += 4.0
		} else {
			break
		}
	}
	DrawMindustryFormatString(dc, si.Host, cardX, cardY+105-yDec, currSize)

	loadFont(dc, 58)
	dc.SetHexColor(labelColor)
	dc.DrawString(si.Address, cardX, cardY+185)

	// line1Y := cardY + 185 + 20 + 22
	// dc.SetHexColor(cardOutlineColor)
	// dc.SetLineWidth(3.0)
	// dc.DrawLine(cardX, line1Y, cardX+cardWidth, line1Y)
	// dc.Stroke()

	// online
	var sumPlayers int
	var countValid int
	for _, r := range hr {
		if r.Players != -1 {
			sumPlayers += r.Players
			countValid++
		}
	}
	var avgStr string
	if countValid == 0 {
		avgStr = "0"
	} else {
		averagePlayers := float64(sumPlayers) / float64(countValid)
		avgStr = fmt.Sprintf("%.1f", averagePlayers)
	}

	yCommon := 65.0 //200.0

	// line2Y := yBottom + 25 + 30
	// dc.SetHexColor(cardOutlineColor)
	// dc.SetLineWidth(3.0)
	// dc.DrawLine(cardX, line2Y, cardX+cardWidth, line2Y)
	// dc.Stroke()

	loadFont(dc, 128)
	dc.SetHexColor(secondaryColor)
	dc.DrawString(avgStr, cardX+40, yCommon+yBottom+140+20)

	loadFont(dc, 32)
	dc.SetHexColor(textPrimary)
	dc.DrawString("средний онлайн:", cardX+40, yCommon+yBottom+50)

	loadFont(dc, 128)
	dc.SetHexColor(secondaryColor)
	str2 := itoa(maxPlayers)
	wStr2, _ := dc.MeasureString(str2)
	dc.DrawString(str2, cardX+cardWidth/2-wStr2, yCommon+yBottom+140+20)

	loadFont(dc, 32)
	dc.SetHexColor(textPrimary)
	dc.DrawString("макс онлайн:", cardX+cardWidth/2-wStr2, yCommon+yBottom+50)

	loadFont(dc, 128)
	dc.SetHexColor(secondaryColor)
	str3 := "---"
	wStr3, _ := dc.MeasureString(str3)
	dc.DrawString(str3, cardWidth-cardX-wStr3/2-40, yCommon+yBottom+140+20)

	loadFont(dc, 32)
	dc.SetHexColor(textPrimary)
	dc.DrawString("средний пинг:", cardWidth-cardX-wStr3/2-40, yCommon+yBottom+50)
}

func loadFont(dc *gg.Context, fontSize int) {
	err := dc.LoadFontFace("./fonts/mindustry.ttf", float64(fontSize))
	if err != nil {
		log.Printf("ршибка загрузки шрифта: %v", err)
	}
}

func router() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("/image", generateAndServeImage)
	router.Run(":8080")
}

func generateAndServeImage(c *gin.Context) {
	address := c.DefaultQuery("address", "none")
	if address == "none" {
		c.Status(http.StatusBadRequest)
		return
	}

	stats, err := getStatsByAddress(address, 12)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// serverName := "none"
	var serverInfo *serverinfo.ServerInfo
	for _, srv := range servers {
		for _, addr := range srv.Address {
			key := addr.Host + ":" + itoa(addr.Port)
			if key == address {
				serverInfo = infoObjects[address]
				// serverName = srv.Name
			}
		}
	}

	if serverInfo == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	dc := gg.NewContext(width, height)
	genImage(dc, *serverInfo, stats)
	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		c.String(http.StatusInternalServerError, "ошибка генерации изображения: %v", err)
		return
	}
	c.Data(http.StatusOK, "image/png", buf.Bytes())
}
