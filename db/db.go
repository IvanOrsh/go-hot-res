package db

const (
	DBNAME     = "hot-res"
	TestDBNAME = "hot-res-test"
	DBURI      = "mongodb://localhost:27017"
)

type Store struct {
	User  UserStore
	Hotel HotelStore
	Room  RoomStore
}
