package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"minimalytics/model"
	"net/http"
	"strconv"
	"strings"
)

type Message struct {
	Event string
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type StatRequest struct {
	Event string `json:"event"`
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func writeResponse(w http.ResponseWriter, err error, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	var response Response
	var status string = "OK"

	if err != nil {
		status = "ERROR"
		w.WriteHeader(http.StatusBadRequest)
		log.Println(message)
		log.Println(err)
	}

	response = Response{
		Status:  status,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func HandleGraphs(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	trimmedPath := strings.Trim(path, "/")
	parts := strings.Split(trimmedPath, "/")

	if len(parts) == 2 {
		switch r.Method {
		case http.MethodPost:
			var postData model.GraphCreate
			if err := json.NewDecoder(r.Body).Decode(&postData); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

			}

			err := model.CreateGraph(postData)
			if err != nil {
				writeResponse(w, err, err.Error(), nil)

			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		}

	} else if len(parts) == 3 {
		graphId, err := strconv.Atoi(parts[2])
		if err != nil {
			writeResponse(w, err, "Invalid graphId in the request", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			writeResponse(w, nil, "Dashboard Details", model.GetGraph(int64(graphId)))

		case http.MethodPatch:
			var patchData model.GraphUpdate
			if err := json.NewDecoder(r.Body).Decode(&patchData); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
			}

			err = model.UpdateGraph(int64(graphId), patchData)
			if err != nil {
				writeResponse(w, err, err.Error(), nil)
			}

		case http.MethodDelete:
			err = model.DeleteGraph(int64(graphId))
			if err != nil {
				writeResponse(w, err, err.Error(), nil)
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}

}

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	trimmedPath := strings.Trim(path, "/")
	parts := strings.Split(trimmedPath, "/")

	if len(parts) == 2 {
		writeResponse(w, nil, "Dashboards Details", model.GetDashboards())

	} else if len(parts) > 2 {
		dashboardId, err := strconv.Atoi(parts[2])
		if err != nil {
			writeResponse(w, err, "Invalid dashboardId in the request", nil)
			return
		}

		if len(parts) == 3 {
			writeResponse(w, nil, "Dashboard details", model.GetDashboard(int64(dashboardId)))

		} else if len(parts) == 4 {
			writeResponse(w, nil, "Graph details", model.GetDashboardGraphs(int64(dashboardId)))
		} else {
			writeResponse(w, errors.New("Invalid request"), "Invalid request", nil)
		}
	}

}

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeResponse(w, nil, "Request received", nil)
		return
	}

	key := parts[3]
	config := model.GetConfig(key)

	value := config.Value
	writeResponse(w, nil, "Value", value)
}

func HandleEventDefsApi(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, nil, "Events Definitions", model.GetEventDefs())
}

func HandleEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var t Message
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	event := t.Event
	model.InitEvent(event)
	model.SubmitDailyEvent(event)
	model.SubmitHourlyEvent(event)
	model.SubmitMinuteEvent(event)

	io.WriteString(w, "OK")
}

func HandleEvents(w http.ResponseWriter, r *http.Request) {

}

func HandleStat(w http.ResponseWriter, r *http.Request) {
	var statRequest StatRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		writeResponse(w, err, "Unable to read body", nil)

		return
	}
	defer r.Body.Close()

	if len(string(body)) <= 2 {
		writeResponse(w, errors.New("Inavlid body size"), "No event provided in request", nil)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(&statRequest)
	if err != nil {
		writeResponse(w, err, "Invalid Request Body", nil)
	}

	if r.URL.Path == "/api/stat/daily/" {
		writeResponse(w, nil, "Daily Stat New", model.GetDailyStat(statRequest.Event))

	} else if r.URL.Path == "/api/stat/hourly/" {
		writeResponse(w, nil, "Hourly Stat", model.GetHourlyStat(statRequest.Event))

	} else if r.URL.Path == "/api/stat/minutes/" {
		writeResponse(w, nil, "Minute Stat", model.GetMinuteStat(statRequest.Event))

	} else {
		writeResponse(w, nil, "Not implemented", nil)

	}

}

func HandleTest(w http.ResponseWriter, r *http.Request) {
}
