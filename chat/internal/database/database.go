package database

import (
	"chat/internal/models"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserCreate struct {
	Email          string
	Password       string
	Name           string
	Token          string
	Avatar         string
	Email_Verified string
	Is_admin       bool
}

type UserUpdate struct {
	Email          string
	Password       string
	Name           string
	Token          string
	Avatar         string
	Email_Verified string
	Is_admin       bool
}

type GroupData struct {
	Name          string
	Description   string
	Image         string
	OwnerId       int
	LastMassageId int
}

type ConversationData struct {
	UserId1       int
	UserId2       int
	LastMassageId int
}

type MessageData struct {
	SenderId   int
	ReceiverId int
	Message    string
	GroupId    int
}

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string
	// ---------create-------------------------------
	CreateUser(user UserCreate) (*models.User, error)
	CreateGroup(group GroupData) (*models.Group, error)
	// --------------------Verify -------------------
	VerifyUserAndUpdate(token string) (*models.User, error)
	// -----------------Delete-----------------------
	DeleteUser(id int) (*models.User, error)
	DeleteGroupMessages(groupId int) error // Add this line
	DeleteGroup(id int) error              // Update return type
	// ---------------------Find----------------------
	FindUserByEmail(email string, password string) (*models.User, error)
	FindUserByEmailOnly(email string) (*models.User, error)
	FindUserByToken(token string) (*models.User, error)
	FindGroupById(id int) (*models.Group, error)
	FindAllGroups(page, limit int) ([]models.Group, int64, error)
	// --------------------Update---------------------------
	UpdateUser(id int, userData UserUpdate) (*models.User, error)
	UpdateUserToken(id int, token string) (*models.User, error)
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	AutoMigrate() error // Add this line
	Close() error
}

type service struct {
	db *gorm.DB
}

// -----------------------------------------------------
// -----------------FindUserByEmail --------------------
// -----------------------------------------------------
func (s *service) FindUserByEmail(email string, password string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil
	}
	return &user, nil
}

func (s *service) FindUserByEmailOnly(email string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// ------------------------------------------------
// -----------------CreateUser --------------------
// ------------------------------------------------
func (s *service) CreateUser(userData UserCreate) (*models.User, error) {
	user := &models.User{
		Email:    userData.Email,
		Password: userData.Password,
		Name:     userData.Name,
		Token:    userData.Token,
		Avatar:   userData.Avatar,
	}

	result := s.db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

// ------------------------------------------------
// -----------------CreateGroup --------------------
// ------------------------------------------------
func (s *service) CreateGroup(groupData GroupData) (*models.Group, error) {
	group := &models.Group{
		Name:        groupData.Name,
		Description: groupData.Description,
		OwnerId:     groupData.OwnerId,
		Image:       groupData.Image,
	}
	result := s.db.Create(group)

	if result.Error != nil {
		return nil, result.Error
	}

	return group, nil
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
		"email_verified": true,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// ---------------------------------------------------------
// -----------------Delete -----------------------------
// ---------------------------------------------------------

func (s *service) DeleteUser(id int) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	if err := s.db.Delete(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Replace existing DeleteGroup function
func (s *service) DeleteGroup(id int) error {
	var group models.Group
	if err := s.db.First(&group, id).Error; err != nil {
		return err
	}

	if err := s.db.Delete(&group).Error; err != nil {
		return err
	}

	return nil
}

// Add new DeleteGroupMessages function
func (s *service) DeleteGroupMessages(groupId int) error {
	result := s.db.Delete(&models.Message{}, "group_id = ?", groupId)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ---------------------------------------------------------
// -----------------FindUserByToken -----------------------------
// ---------------------------------------------------------

func (s *service) FindUserByToken(token string) (*models.User, error) {
	var user models.User
	result := s.db.Where("token = ?", token).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return &user, nil
}

// ---------------------------------------------------------
// -----------------FindGroupByToken -----------------------------
// ---------------------------------------------------------
func (s *service) FindGroupById(id int) (*models.Group, error) {
	var group models.Group
	result := s.db.First(&group, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return &group, nil
}

func (s *service) FindAllGroups(page, limit int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64

	offset := (page - 1) * limit

	// Get total count
	if err := s.db.Model(&models.Group{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated groups
	result := s.db.Offset(offset).Limit(limit).Find(&groups)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return groups, total, nil
}

// ---------------------------------------------------------
// -----------------Update -----------------------------
// ---------------------------------------------------------
func (s *service) UpdateUserToken(id int, token string) (*models.User, error) {
	var user models.User

	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&user).Update("token", token).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *service) UpdateUser(id int, userData UserUpdate) (*models.User, error) {
	var user models.User

	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"email":       userData.Email,
		"password":    userData.Password,
		"name":        userData.Name,
		"profile_url": userData.Avatar,
		"token":       userData.Token,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// ---------------------------------------------------------
// -----------------FindIF -----------------------------
// ---------------------------------------------------------

var (
	dbname     = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbname)

	// Opening a connection to the database using GORM
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Suppress all GORM logs (including SQL)
	})
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) AutoMigrate() error {
	return s.db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.Conversation{},
		&models.Message{},
		&models.MessageAttachment{},
	)
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Get the underlying sql.DB from GORM
	sqlDB, err := s.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("failed to get sql.DB: %v", err)
		log.Fatalf("failed to get sql.DB: %v", err) // Log the error and terminate the program
		return stats
	}

	// Ping the database
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
	log.Printf("Disconnected from database: %s", dbname)
	return sqlDB.Close()
}
