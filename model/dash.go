package model

import "log"

type Dashboard struct {
	Id        int64
	Name      string
	CreatedOn string
}

func InitDashboards() {
	query := `
		CREATE TABLE IF NOT EXISTS dashboards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			createdOn TEXT
		);`

	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
		return
	}

}

func GetDashboards() []Dashboard {
	rows, err := db.Query("select * from dashboards")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var dashboards []Dashboard
	for rows.Next() {
		var dash Dashboard
		err := rows.Scan(&dash.Id, &dash.Name, &dash.CreatedOn) // Replace with actual fields
		if err != nil {
			// Handle scan error
			panic(err)
		}
		dashboards = append(dashboards, dash)
	}

	return dashboards
}

func GetDashboard(dashboardId int64) Dashboard {
	row := db.QueryRow("select * from dashboards where id = ?", dashboardId)

	var dashboard Dashboard
	err := row.Scan(&dashboard.Id, &dashboard.Name, &dashboard.CreatedOn)
	if err != nil {
		panic("Dashboard not found")
	}

	return dashboard
}
