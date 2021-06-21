package repository

import (
	"time"

	"github.com/k3forx/booking-app/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(startDate, endDate time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(startDate, endDate time.Time) ([]models.Room, error)
}
