package dbrepo

import (
	"github.com/k3forx/booking-app/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	return 0, nil
}

func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	return nil
}
