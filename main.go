package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/u-root/u-root/pkg/curl"
)

var db *sql.DB

type Message struct {
	Event string
}

type StatRequest struct {
	Event string
}

type DateEvent struct {
	Id int64
	Timestamp int64
	Date string
	Event string
	Count int64
}

type HourEvent struct {
	Id int64
	Timestamp int64
	Hour string
	Event string
	Count int64
}

type MinuteEvent struct {
	Id int64
	Minute string
	Event string
	Count int64
}

type DateStat struct {
	Date string `json:"date"`
	Count int64 `json:"count"`
}

type HourStat struct {
	Hour string `json:"hour"`
	Count int64 `json:"count"`
}

type MinuteStat struct {
	Minute string `json:"minute"`
	Count int64 `json:"count"`
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func dailyEvent(event string) {
	currentTime := time.Now()
	date := currentTime.Format("2006-01-02")

	row := db.QueryRow("select * from days where date = ? and event = ?", date, event)

	var eventRow DateEvent
	err := row.Scan(&eventRow.Id, &eventRow.Date, &eventRow.Event, &eventRow.Count)
	if err != nil {
		db.Exec("insert into days (date, event, count) values (?, ?, ?)", date, event, 1)

	} else {

		rowId := eventRow.Id
		nextCount := eventRow.Count + 1
		db.Exec("update days set count = ? where id = ?", nextCount, rowId)
	}
}

func hourEvent(event string) {
	currentTime := time.Now()

	hour := currentTime.Format("2006-01-02 15:00:00")
	row := db.QueryRow("select * from hours where hour = ? and event = ?", hour, event)

	var eventRow HourEvent

	err := row.Scan(&eventRow.Id, &eventRow.Hour, &eventRow.Event, &eventRow.Count)
	if err != nil {
		db.Exec("insert into hours (hour, event, count) values (?, ?, ?)", hour, event, 1)

	} else {
		rowId := eventRow.Id
		nextCount := eventRow.Count + 1
		db.Exec("update hours set count = ? where id = ?", nextCount, rowId)

	}
}

func minuteEvent(event string) {
	currentTime := time.Now()
	minute := currentTime.Format("2006-01-02 15:04:00")

	log.Println(minute)

	row := db.QueryRow("select * from minutes where minute = ? and event = ?", minute, event)

	var eventRow MinuteEvent

	err := row.Scan(&eventRow.Id, &eventRow.Minute, &eventRow.Event, &eventRow.Count)
	if err != nil {
		db.Exec("insert into minutes (minute, event, count) values (?, ?, ?)", minute, event, 1)

	} else {
		rowId := eventRow.Id
		nextCount := eventRow.Count + 1
		db.Exec("update minutes set count = ? where id = ?", nextCount, rowId)
	}
}

func handleEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var t Message
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	event := t.Event

	minuteEvent(event)
	hourEvent(event)
	dailyEvent(event)

	io.WriteString(w, "OK")
}

func getMinuteStat(event string) [60]MinuteStat {

	currentTime := time.Now()
	toMinute := currentTime.Format("2006-01-02 15:04:00")
	fromMinute := currentTime.AddDate(0, 0, -60).Format("2006-01-02 15:04:00")

	rows, err := db.Query("select * from minutes where minute between ? and ? and event = ?", fromMinute, toMinute, event)
	if err != nil {
		panic(err)
	}

	countMap := make(map[string]int64)
	for rows.Next() {
		var eventRow MinuteEvent
		err = rows.Scan(&eventRow.Id, &eventRow.Minute, &eventRow.Event, &eventRow.Count)
		if err != nil {
			panic(err)
		}

		countMap[eventRow.Minute] = eventRow.Count
	}

	var statsArray [60]MinuteStat
	for i := 0; i < 60; i++ {
		// iMinute := currentTime.AddDate(0, 0, -1 * i).Format("2006-01-02")
		iMinute := currentTime.Add(time.Duration(-i) * time.Minute).Format("2006-01-02 15:04:00")
		iCount := int64(0)
		
		realCount, ok := countMap[iMinute]
		if ok {
			iCount = realCount
		}

		iStatItem := MinuteStat{
			Minute: iMinute,
			Count: iCount,
		}

		statsArray[i] = iStatItem
	}

	return statsArray
}

func getHourlyStat(event string) [48]HourStat {
	currentTime := time.Now()
	toHour := currentTime.Format("2006-01-02 15:00:00")
	fromHour := currentTime.AddDate(0, 0, -48).Format("2006-01-02 15:00:00")

	rows, err := db.Query("select * from hours where hour between ? and ? and event = ?", fromHour, toHour, event)
	if err != nil {
		panic(err)
	}

	countMap := make(map[string]int64)
	for rows.Next() {
		var eventRow HourEvent
		err = rows.Scan(&eventRow.Id, &eventRow.Hour, &eventRow.Event, &eventRow.Count)
		if err != nil {
			panic(err)
		}

		countMap[eventRow.Hour] = eventRow.Count
	}

	var statsArray [48]HourStat
	for i := 0; i < 48; i++ {
		// iHour := currentTime.AddDate(0, 0, -1 * i).Format("2006-01-02")
		iHour := currentTime.Add(time.Duration(-i) * time.Hour).Format("2006-01-02 15:00:00")
		iCount := int64(0)
		
		realCount, ok := countMap[iHour]
		if ok {
			iCount = realCount
		}

		iStatItem := HourStat{
			Hour: iHour,
			Count: iCount,
		}

		statsArray[i] = iStatItem
	}

	return statsArray
}

func getDailyStat(event string) [30]DateStat{
	currentTime := time.Now()
	toDate := currentTime.Format("2006-01-02")
	fromDate := currentTime.AddDate(0, 0, -30).Format("2006-01-02")

	rows, err := db.Query("select * from days where date between ? and ? and event = ?", fromDate, toDate, event)
	if err != nil {
		panic(err)
	}

	countMap := make(map[string]int64)
	for rows.Next() {
		var eventRow DateEvent
		err = rows.Scan(&eventRow.Id, &eventRow.Date, &eventRow.Event, &eventRow.Count)
		if err != nil {
			panic(err)
		}

		countMap[eventRow.Date] = eventRow.Count
	}

	var statsArray [30]DateStat
	for i := 0; i < 30; i++ {
		iDate := currentTime.AddDate(0, 0, -1 * i).Format("2006-01-02")
		iCount := int64(0)
		
		realCount, ok := countMap[iDate]
		if ok {
			iCount = realCount
		}

		iStatItem := DateStat{
			Date: iDate,
			Count: iCount,
		}

		statsArray[i] = iStatItem
	}

	return statsArray
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	var statRequest StatRequest
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&statRequest)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")

	if (r.URL.Path == "/api/stat/daily/") {
		stats := getDailyStat(statRequest.Event)

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		}

	} else if (r.URL.Path == "/api/stat/hourly/") {
		stats := getHourlyStat(statRequest.Event)

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		}

	} else if (r.URL.Path == "/api/stat/minutes/") {
		stats := getMinuteStat(statRequest.Event)

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		}

	}


}

func main() {	
	db, _ = sql.Open("sqlite3", "./events.db")

	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/event", handleEvent)
	http.HandleFunc("/", serveTemplate)
	http.HandleFunc("/api/", handleAPI)

	log.Println("Starting server on port 3333")

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
