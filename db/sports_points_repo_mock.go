package db

import (
	"errors"
	"time"
)

type SportsPointsMockRepo struct {
	points  map[string]Point
	created time.Time
}

func NewSportsPointsMockRepo() *SportsPointsMockRepo {
	return &SportsPointsMockRepo{
		points:  make(map[string]Point),
		created: time.Now(),
	}
}

func (s *SportsPointsMockRepo) Ping() error {
	if time.Since(s.created) > time.Second*3 {
		return nil
	}
	return errors.New("not connected")
}

func (s *SportsPointsMockRepo) CreateTablesIfDontExist(sports ...string) error {
	for _, sport := range sports {
		if _, ok := s.points[sport]; !ok {
			s.points[sport] = Point{
				Time:  time.Now(),
				Value: 0,
			}
		}
	}
	return nil
}

func (s *SportsPointsMockRepo) PutPoint(sport string, point Point) error {
	s.points[sport] = point
	return nil
}

func (s *SportsPointsMockRepo) GetCurrent(sport string) (Point, error) {
	return s.points[sport], nil
}
