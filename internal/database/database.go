package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Service interface {
	GetDB() *sql.DB
	GetGORMDB() *gorm.DB
	Close() error
}

type service struct {
	db     *sql.DB
	gormDB *gorm.DB
}

var dbInstance *service

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:hilmi290208@localhost:5432/ezytixweb?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Failed to connect DB:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping DB:", err)
	}

	// Pool configuration
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to start GORM:", err)
	}

	dbInstance = &service{db: db, gormDB: gormDB}
	return dbInstance
}

func (s *service) GetDB() *sql.DB {
	return s.db
}

func (s *service) GetGORMDB() *gorm.DB {
	return s.gormDB
}

func (s *service) Close() error {
	return s.db.Close()
}
