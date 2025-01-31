package database

import (
	"JobBoard/internal/models"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserCreate struct {
	ID         int
	Email      string
	Password   string
	Name       string
	Token      string
	Avatar     string
	IsVerified bool
	Resume     string
	IsEmployer bool
	CompanyID  *uint
}

type CompanyCreate struct {
	Email       string
	Name        string
	Description string
	Website     string
	Location    string
	Logo        string
}

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string
	//---------------------- Find---------------------------
	FindUserByEmail(email string) (*models.User, error)
	FindCompanyByEmail(email string) (*models.Company, error)
	//-----------------------Create ------------------------
	CreateUser(user UserCreate) (*models.User, error)
	CreateCompany(company CompanyCreate) (*models.Company, error)
	// --------------------Verify --------------------------
	VerifyUserAndUpdate(token string) (*models.User, error)
	// -------------------- Update--------------------------
	UpdateUserCompanyID(Id int, companyId uint, userDate UserCreate) error
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
	GetDB() *gorm.DB // Add this method
}

type service struct {
	db *gorm.DB
}

func (s *service) GetDB() *gorm.DB {
	return s.db
}

// ---------------------------------------
// ----------------- Create ---------------
// ----------------------------------------
func (s *service) CreateUser(user UserCreate) (*models.User, error) {
	newUser := &models.User{
		Email:      user.Email,
		Password:   user.Password,
		Name:       user.Name,
		Avatar:     user.Avatar,
		IsEmployer: user.IsEmployer,
		Token:      user.Token,
		CompanyID:  user.CompanyID,
	}

	result := s.db.Create(newUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return newUser, nil
}

func (s *service) CreateCompany(company CompanyCreate) (*models.Company, error) {
	newCompany := &models.Company{
		Email:       company.Email,
		Description: company.Description,
		Name:        company.Name,
		Logo:        company.Logo,
		Location:    company.Location,
		Website:     company.Website,
	}

	result := s.db.Create(newCompany)
	if result.Error != nil {
		return nil, result.Error
	}
	return newCompany, nil
}

// -----------------------------------------
// -------------- Find ---------------------
// -----------------------------------------

func (s *service) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *service) FindCompanyByEmail(email string) (*models.Company, error) {
	var Company models.Company
	result := s.db.Where("email = ?", email).First(&Company)
	if result.Error != nil {
		return nil, result.Error
	}
	return &Company, nil
}

// ---------------------------------------------------------
// -----------------VerifyUserAndUpdate --------------------
// ---------------------------------------------------------

func (s *service) VerifyUserAndUpdate(token string) (*models.User, error) {
	var user models.User

	// Find user by token
	result := s.db.Where("token = ?", token).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	// Update user verification status
	updates := map[string]interface{}{
		"is_verified": true,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// ---------------------------------------------------------
// ---------------------------Update -------------------------
// ---------------------------------------------------------

func (s *service) UpdateUserCompanyID(Id int, companyId uint, userDate UserCreate) error {
	var user models.User

	if err := s.db.First(&user, Id).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"CompanyID":  companyId,
		"IsEmployer": true, // Add this line to set the user as an employer
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	sslmode    = os.Getenv("BLUEPRINT_DB_SSLMODE")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s search_path=%s", host, username, password, database, port, sslmode, schema)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Job{},
		&models.Skill{},
		&models.Application{},
	)
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	sqlDB, err := s.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := sqlDB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	log.Printf("Disconnected from database: %s", database)
	return sqlDB.Close()
}
