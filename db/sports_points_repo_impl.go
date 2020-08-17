package db

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // postgres support
	"go.uber.org/zap"
)

// PGSportsPointsRepo is a postgres implementation of SportsPointsRepo
type PGSportsPointsRepo struct {
	db      *gorm.DB
	errors  chan error
	slogger *zap.SugaredLogger
}

// NewPGSportsPointsRepo creates PGSportsPointsRepo with specified logger
func NewPGSportsPointsRepo(slogger *zap.SugaredLogger) *PGSportsPointsRepo {
	return &PGSportsPointsRepo{
		db:      nil,
		errors:  make(chan error, 1),
		slogger: slogger,
	}
}

// Connect connects postgres database
func (p *PGSportsPointsRepo) Connect(PostgresHost, PostgresPort, PostgresUser, PostgresPassword, PostgresName string) error {
	p.slogger.Infow(
		"Connecting postgres database",
		"host", PostgresHost,
		"port", PostgresPort,
		"name", PostgresName,
	)
	args := fmt.Sprintf("host=%s port=%s sslmode=disable user=%s password=%s dbname=%s",
		PostgresHost,
		PostgresPort,
		PostgresUser,
		PostgresPassword,
		PostgresName)
	db, err := gorm.Open("postgres", args)
	p.db = db
	if err != nil {
		p.slogger.Errorw("Error occurred connecting to postgres database", "err", err)
		p.errors <- err
	} else {
		p.slogger.Info("Connected to postgres database")
	}
	return err
}

// Close closes connection to postgres database
func (p *PGSportsPointsRepo) Close() error {
	p.slogger.Info("Closing connecting to postgres database")
	err := p.db.Close()
	if err != nil {
		p.slogger.Error("Failed to close connection to postgres database")
	}
	return err
}

// Notify returns a channel to notify caller about errors
func (p *PGSportsPointsRepo) Notify() <-chan error {
	return p.errors
}

// Ping creates a ping transaction to check if database is accessible
func (p *PGSportsPointsRepo) Ping() error {
	p.slogger.Debug("PGSportsPointsRepo.Ping called")
	if p.db == nil {
		p.slogger.Warn("Ping transaction created before connecting to postgres database")
		return errors.New("PGSportsPointsRepo is not connected to db")
	}

	err := p.db.Transaction(func(tx *gorm.DB) error {
		return nil
	})
	if err != nil {
		p.slogger.Warn("Ping transaction failed")
		p.errors <- err
	}
	return err
}

// CreateTablesIfDontExist creates tables for specified sports
func (p *PGSportsPointsRepo) CreateTablesIfDontExist(sports ...string) error {
	p.slogger.Debug("PGSportsPointsRepo.CreateTablesIfDontExist called")
	err := p.db.Transaction(func(tx *gorm.DB) error {
		for _, sport := range sports {
			if !tx.HasTable(sport) {
				p.slogger.Debug("Creating table", "sport", sport)
				tx.Table(sport).AutoMigrate(&Point{})
			}
		}
		return nil
	})
	if err != nil {
		p.slogger.Error("Failed to create requested tables")
		p.errors <- err
	}
	return err
}

// PutPoint saves data point for sport
func (p *PGSportsPointsRepo) PutPoint(sport string, point Point) error {
	p.slogger.Debug("PGSportsPointsRepo.PutPoint called")
	err := p.db.Transaction(func(tx *gorm.DB) error {
		tx.Table(sport).Create(&point)
		return nil
	})
	if err != nil {
		p.slogger.Errorw("Failed to save point to database", "sport", sport, "point", point)
		p.errors <- err
	}
	return err
}

// GetCurrent returns last data point for sport
func (p *PGSportsPointsRepo) GetCurrent(sport string) (Point, error) {
	p.slogger.Debug("PGSportsPointsRepo.GetCurrent called")
	var point Point
	err := p.db.Transaction(func(tx *gorm.DB) error {
		if !tx.HasTable(sport) {
			p.slogger.Warnw("Sport for which point is requested does not exist in db")
		}
		tx.Table(sport).Last(&point)
		return nil
	})
	if err != nil {
		p.slogger.Errorw("Failed to get latest point from database")
		p.errors <- err
	}
	return point, err
}
