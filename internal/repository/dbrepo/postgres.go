package dbrepo

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/k3forx/booking-app/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a new reservation into the database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int
	stmt := `INSERT INTO reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id;`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO room_restrictions (start_date, end_date, room_id, reservation_id,
				created_at, updated_at, restriction_id) VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := m.DB.ExecContext(ctx, stmt, r.StartDate, r.EndDate, r.RoomID, r.ReservationID, time.Now(), time.Now(), r.RestrictionID)
	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for roomID, and false if no availability exists
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(startDate, endDate time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `SELECT COUNT(id) FROM room_restrictions WHERE room_id = $1 AND $2 < end_date AND $3 > start_date;`

	row := m.DB.QueryRowContext(ctx, query, roomID, startDate, endDate)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available room, if any, for given data range
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(startDate, endDate time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `SELECT r.id, r.room_name FROM rooms r WHERE r.id NOT IN (
			SELECT room_id FROM room_restrictions rr WHERE $1 < rr.end_date and $2 > rr.start_date);`

	m.App.InfoLog.Println("Search availability for all rooms")
	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	m.App.InfoLog.Println("Successfully find available rooms")
	return rooms, nil
}

// GetRoomByID gets a room by ID
func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `SELECT id, room_name, created_at, updated_at FROM rooms WHERE id = $1;`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		return room, err
	}

	return room, nil
}

func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	query := `SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at FROM users WHERE id = $1;`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.AccessLevel,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		m.App.InfoLog.Println("cannot get user by ID")
		return user, err
	}

	return user, nil
}

func (m *postgresDBRepo) UpdateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE users SET first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5;`

	_, err := m.DB.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.AccessLevel,
		time.Now(),
	)
	if err != nil {
		m.App.ErrorLog.Println("failed to update user")
		return err
	}

	return nil
}

// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(email, textPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		m.App.ErrorLog.Printf("failed to get id and password by email: %s", email)
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(textPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		m.App.ErrorLog.Println("password is wrong")
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

// AllReservations returns a slice of all reservations
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT r.id, r.first_name, r.Last_name, r.email, r.phone, r.start_date, r.end_date,
		r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id = rm.id)
		ORDER BY r.start_date ASC;
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		m.App.ErrorLog.Println("failed to get all reservations")
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.FirstName,
			&reservation.LastName,
			&reservation.Email,
			&reservation.Phone,
			&reservation.StartDate,
			&reservation.EndDate,
			&reservation.RoomID,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.Room.ID,
			&reservation.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, reservation)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// AllNewReservations returns a slice of all reservations
func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT r.id, r.first_name, r.Last_name, r.email, r.phone, r.start_date, r.end_date,
		r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id = rm.id)
		WHERE processed = 0
		ORDER BY r.start_date ASC;
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		m.App.ErrorLog.Println("failed to get all reservations")
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.FirstName,
			&reservation.LastName,
			&reservation.Email,
			&reservation.Phone,
			&reservation.StartDate,
			&reservation.EndDate,
			&reservation.RoomID,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.Processed,
			&reservation.Room.ID,
			&reservation.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, reservation)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservationByID returns a reservation by ID
func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
		SELECT
			r.id, r.first_name, r.last_name, r.email, r.phone,
			r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
			rm.id, rm.room_name
		FROM
			reservations r
		LEFT JOIN
			rooms rm
		ON
			(r.room_id = rm.id)
		WHERE
			r.id = $1;
	`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)

	if err != nil {
		m.App.ErrorLog.Println("failed to map from result to res")
		return res, err
	}

	return res, nil
}

// UpdateReservation updates a reservation in the database
func (m *postgresDBRepo) UpdateReservation(res models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE
			reservations
		SET
			first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
		WHERE
			id = $6;
	`

	_, err := m.DB.ExecContext(ctx, query,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		time.Now(),
		res.ID,
	)
	if err != nil {
		m.App.ErrorLog.Println("failed to update reservation")
		return err
	}

	return nil
}

// DeleteReservationByID deletes a reservation by ID
func (m *postgresDBRepo) DeleteReservationByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM reservations WHERE id = $1;`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		m.App.ErrorLog.Printf("failed to delete a reservation by ID: %d, err: %s", id, err)
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates processed for reservation by ID
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE reservations SET processed = $1 WHERE id = $2;`

	_, err := m.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		m.App.ErrorLog.Printf("failed to update processed of a reservation by ID: %d, err: %s", id, err)
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetAllRooms() ([]models.Room, error) {
	m.App.InfoLog.Printf("Get all rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `SELECT id, room_name, created_at, updated_at FROM rooms ORDER BY room_name;`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		m.App.ErrorLog.Printf("failed to get all rooms: %s", err)
		return rooms, err
	}
	m.App.InfoLog.Println("Get all rooms successfully")

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRestrictionsForRoomByDate returns restrictions for a room by date range
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `
		SELECT
			id, COALESCE(reservation_id, 0), restriction_id, room_id, start_date, end_date
		FROM
			room_restrictions
		WHERE
			$1 < end_date
			AND
				$2 >= start_date
			AND
				room_id = $3;
	`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		m.App.ErrorLog.Printf("failed to get room restrictions: %s", err)
		return restrictions, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)

		if err != nil {
			m.App.ErrorLog.Printf("failed to map to room restriction: %s", err)
			return nil, err
		}

		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		m.App.ErrorLog.Printf("failed to map to room restriction: %s", err)
		return nil, err
	}

	return restrictions, nil
}
