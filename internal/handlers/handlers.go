package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/k3forx/booking-app/internal/config"
	"github.com/k3forx/booking-app/internal/driver"
	"github.com/k3forx/booking-app/internal/forms"
	"github.com/k3forx/booking-app/internal/helpers"
	"github.com/k3forx/booking-app/internal/models"
	"github.com/k3forx/booking-app/internal/render"
	"github.com/k3forx/booking-app/internal/repository"
	"github.com/k3forx/booking-app/internal/repository/dbrepo"
)

// Repo is the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(a, db.SQL),
	}
}

// NewTestRepo creates a new repository for testing
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the about page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		StringMap: stringMap,
		Data:      data,
	})
}

// PostReservation renders the make a reservation page and displays form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("cannot get from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	m.App.InfoLog.Println("Get reservation from context in session")

	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("cannot parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	m.App.InfoLog.Println("parsed form")

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		m.App.InfoLog.Println("the form is invalid")
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// send notifications
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s: <br>
		This is confirm your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "me@here.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.tmpl",
	}

	m.App.MailChan <- msg

	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong><br>
		A reservation has been made for %s from %s to %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:      "me@here.com",
		From:    "me@here.com",
		Subject: "Reservation Notification",
		Content: htmlMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability renders the make a reservation page and displays form
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("can't parse form'")
		// m.App.Session.Put(r.Context(), "error", "can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	// 2020-01-02 -- 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		// no available rooms
		m.App.InfoLog.Println("NO available rooms")
		m.App.Session.Put(r.Context(), "error", "No available rooms found")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles requests for availability and send JSON responses
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("Search availability")

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.ErrorLog.Printf("can't parse start date: %s", sd)
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.InfoLog.Println("Successfully parsed stat date and end date")

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		m.App.InfoLog.Println("failed to connect to database")
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	m.App.InfoLog.Println("Return JSON response for available room")

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get error from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
	})
}

// ChooseRoom displays list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes ULR parameters, builds a sessional variable, and takes user to make reservation screen
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var res models.Reservation

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("Post login information")

	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("failed to parse login form")
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		m.App.InfoLog.Println("email or password is missing")
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.ErrorLog.Println("failed to authenticate")
		m.App.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminNewReservations starts")
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		m.App.ErrorLog.Println("failed to get all reservations")
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	m.App.InfoLog.Println("AdminNewReservations is finished")

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminReservations shows all reservations in admin tool
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminALlReservations starts")
	reservations, err := m.DB.AllReservations()
	if err != nil {
		m.App.ErrorLog.Println("failed to get all reservations")
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	m.App.InfoLog.Println("AdminALlReservations is finished")

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows the reservation in admin tool
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminShowReservation starts")
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		m.App.ErrorLog.Panicln("cannot get ID from request URI")
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")

	stringMap["year"] = year
	stringMap["month"] = month

	m.App.InfoLog.Printf("Search a reservation by ID: %d", id)

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		m.App.ErrorLog.Printf("failed to get a reservation by ID: %d, err: %s", id, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation

	m.App.InfoLog.Printf("Show a reservation by ID: %d", id)
	m.App.InfoLog.Println("AdminShowReservation is finished")
	render.Template(w, r, "admin-reservation-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminPostShowReservation updates the reservation in admin tool
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminPostShowReservation starts")
	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("cannot parse form")
		http.Redirect(w, r, "/admin/dashboard", http.StatusTemporaryRedirect)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		m.App.ErrorLog.Panicln("cannot get ID from request URI")
		return
	}

	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src

	m.App.InfoLog.Printf("Search a reservation by ID: %d", id)

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		m.App.ErrorLog.Printf("failed to get a reservation by ID: %d, err: %s", id, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		m.App.ErrorLog.Printf("failed to update a reservation by ID: %d, err: %s", id, err)
		return
	}
	m.App.InfoLog.Printf("Updated a reservation")

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?year=%s&month=%s", year, month), http.StatusSeeOther)
	}

	m.App.InfoLog.Println("AdminPostShowReservation is finished")
}

// AdminReservationsCalendar displays the reservation calendar
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminReservationCalendar starts")
	now := time.Now()

	if r.URL.Query().Get("year") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("year"))
		month, _ := strconv.Atoi(r.URL.Query().Get("month"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.GetAllRooms()
	if err != nil {
		m.App.ErrorLog.Printf("failed to get all rooms: %s", err)
		return
	}

	data["rooms"] = rooms

	for _, room := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		// Initialize reservationMap and blockMap
		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006 01 2")] = 0
			blockMap[d.Format("2006 01 2")] = 0
		}

		// Get all the restrictions from the current room
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(room.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			m.App.ErrorLog.Printf("failed to get room restrictions: %s", err)
			return
		}

		for _, r := range restrictions {
			if r.ReservationID > 0 {
				// It's a reservation
				for d := r.StartDate; d.After(r.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-02")] = r.ReservationID
				}
			} else {
				// This is a block by owner
				blockMap[r.StartDate.Format("2006-01-02")] = r.ID
			}
		}
		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = blockMap
		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	m.App.InfoLog.Println("AdminReservationCalendar is finished")

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminProcessReservation starts")
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		m.App.ErrorLog.Printf("failed to process a reservation: %s", err)
		return
	}

	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")

	m.App.Session.Put(r.Context(), "flash", "Reservation is marked as processed")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?year=%s&month=%s", year, month), http.StatusSeeOther)
	}

	m.App.InfoLog.Println("AdminProcessReservation is finished")
}

// AdminDeleteReservation deletes a reservation
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminDeleteReservation starts")
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.DeleteReservationByID(id)
	if err != nil {
		m.App.ErrorLog.Printf("failed to process a reservation: %s", err)
		return
	}

	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")

	m.App.Session.Put(r.Context(), "flash", "Reservation is deleted")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?year=%s&month=%s", year, month), http.StatusSeeOther)
	}

	m.App.InfoLog.Println("AdminDeleteReservation is finished")
}

// AdminPostReservationsCalendar handles post of reservations calendar
func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	m.App.InfoLog.Println("AdminPostReservationsCalendar starts")

	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Printf("failed to parse form: %s", err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("year"))
	month, _ := strconv.Atoi(r.Form.Get("month"))

	// Process blocks
	rooms, err := m.DB.GetAllRooms()
	if err != nil {
		m.App.ErrorLog.Printf("failed to get rooms: %s", err)
		return
	}

	form := forms.New(r.PostForm)

	for _, room := range rooms {
		// Get the blockMap for the session, loop through entire, map, if we have an entry in the map
		// that does not exist in our posted data, and if th restriction id > 0, then it is a block we need to
		// remove
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", room.ID)).(map[string]int)
		for name, value := range curMap {
			// ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok {
				// only pay attention to values > 0, and that are not in the form post
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", room.ID, name)) {
						// delete the restriction by Id
						err := m.DB.DeleteBlockForRoom(value)
						if err != nil {
							m.App.ErrorLog.Printf("failed to delete room by id: %s", err)
							return
						}
					}
				}
			}
		}
	}

	// Now handle new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, err := strconv.Atoi(exploded[2])
			if err != nil {
				m.App.ErrorLog.Printf("failed to convert roomID to int: %s", err)
				return
			}
			t, _ := time.Parse("2006-01-2", exploded[3])
			if err != nil {
				m.App.ErrorLog.Printf("failed to parse time: %s", err)
				return
			}

			// insert a new block
			m.App.InfoLog.Printf("Insert a new block by roomID=%d, startDate=%s", roomID, t.Format("2006-01-02"))
			err = m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				m.App.ErrorLog.Printf("failed to insert a new room: %s", err)
				return
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?year=%d&month=%d", year, month), http.StatusSeeOther)

	m.App.InfoLog.Println("AdminPostReservationsCalendar starts")
}
