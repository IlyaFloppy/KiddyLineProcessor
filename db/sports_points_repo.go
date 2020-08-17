package db

// SportsPointsRepo is an interface for lines repo
type SportsPointsRepo interface {
	Ping() error
	CreateTablesIfDontExist(sports ...string) error
	PutPoint(sport string, point Point) error
	GetCurrent(sport string) (Point, error)
}
