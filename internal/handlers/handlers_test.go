package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k3forx/booking-app/internal/models"
)

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedstatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},

	// {"post-search-avail", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-31"},
	// }, http.StatusOK},
	// {"post-search-avail-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "2020-01-01"},
	// 	{key: "end", value: "2020-01-31"},
	// }, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)

		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedstatusCode {
			t.Errorf("for %s, expected %d but god %d", e.name, e.expectedstatusCode, resp.StatusCode)
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)

	// WithContext returns a shallow copy of req with its context changed to ctx
	req = req.WithContext(ctx)

	// NewRecorder returns an initialized ResponseRecorder
	// ResponseRecorder is an implementation of http.ResponseWriter that records its mutations for later inspection in tests
	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	// The HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
	// If f is a function with appropriate signature, HandlerFunc(f) is a Handler that calls f.
	handler := http.HandlerFunc(Repo.Reservation)

	// handler.ServeHTTP calls handler(rr, req)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// Test case:
	// where reservation is not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}

	// Test case:
	// Can't find room for a given ID
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusOK)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	// Test case:
	// can't get session from context
	req, _ := http.NewRequest("POST", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test case
	// missing request body
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// // Test case
	// // invalid start date
	// reqBody = "start_date=invalid"
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	// req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// rr = httptest.NewRecorder()

	// handler = http.HandlerFunc(Repo.PostReservation)

	// handler.ServeHTTP(rr, req)
	// if rr.Code != http.StatusTemporaryRedirect {
	// 	t.Errorf("PostReservation handler returned wrong response code for invalid start date: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	// }

	// // Test case
	// // invalid end date
	// reqBody = "start_date=2050-01-01"
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=invalid")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	// req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// rr = httptest.NewRecorder()

	// handler = http.HandlerFunc(Repo.PostReservation)

	// handler.ServeHTTP(rr, req)
	// if rr.Code != http.StatusTemporaryRedirect {
	// 	t.Errorf("PostReservation handler returned wrong response code for invalid end date: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	// }

	// // Test case
	// // invalid room id
	// reqBody = "start_date=2050-01-01"
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=invalid")

	// req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// rr = httptest.NewRecorder()

	// handler = http.HandlerFunc(Repo.PostReservation)

	// handler.ServeHTTP(rr, req)
	// if rr.Code != http.StatusTemporaryRedirect {
	// 	t.Errorf("PostReservation handler returned wrong response code for invalid room id: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	// }

	// // Test case
	// // invalid data
	// reqBody = "start_date=2050-01-01"
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=J")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	// req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// rr = httptest.NewRecorder()

	// handler = http.HandlerFunc(Repo.PostReservation)

	// handler.ServeHTTP(rr, req)
	// if rr.Code != http.StatusSeeOther {
	// 	t.Errorf("PostReservation handler returned wrong response code for invalid data: got %d, wanted %d", rr.Code, http.StatusSeeOther)
	// }

	// // Test case
	// // for failure to insert reservation into database
	// reqBody = "start_date=2050-01-01"
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=2")

	// req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// rr = httptest.NewRecorder()

	// handler = http.HandlerFunc(Repo.PostReservation)

	// handler.ServeHTTP(rr, req)
	// if rr.Code != http.StatusTemporaryRedirect {
	// 	t.Errorf("PostReservation handler failed when trying to fail inerting reservation: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	// }

	// // Test case
	// // for failure to insert room restriction into database
	// reqBody = "start_date=2050-01-01"
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=John")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Smith")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "email=john@smith.com")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1000")

	// req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// rr = httptest.NewRecorder()

	// handler = http.HandlerFunc(Repo.PostReservation)

	// handler.ServeHTTP(rr, req)
	// if rr.Code != http.StatusTemporaryRedirect {
	// 	t.Errorf("PostReservation handler failed when trying to fail inerting room restriction: got %d, wanted %d", rr.Code, http.StatusTemporaryRedirect)
	// }

}

// func TestRepository_AvailabilityJSON(t *testing.T) {
// 	// Test case
// 	// rooms are available
// 	reqBody := "start_date=2050-01-01"
// 	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
// 	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
// 	t.Logf("Request body: %s", reqBody)

// 	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
// 	t.Logf(req.FormValue("start_date"))
// 	ctx := getCtx(req)
// 	req = req.WithContext(ctx)
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	rr := httptest.NewRecorder()

// 	handler := http.HandlerFunc(Repo.AvailabilityJSON)

// 	handler.ServeHTTP(rr, req)

// 	var j jsonResponse
// 	log.Println("Response body", rr.Result().Body)
// 	err := json.Unmarshal([]byte(rr.Body.Bytes()), &j)
// 	t.Log(err)
// 	if err != nil {
// 		t.Error("failed to parse json")
// 	}
// }

func getCtx(req *http.Request) context.Context {
	// Load retrieves the session data for the given token from the session store,
	// and returns a new context.Context containing the session data. If no matching
	// token is found then this will create a new session.
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
