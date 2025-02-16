package database

import (
	"AirBnb/internal/models"
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

type User struct {
	ID         uint
	Email      string
	Password   string
	Name       string
	Avatar     string
	IsActive   bool
	IsStaff    bool
	Token      string
	IsVerified bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Property struct {
	ID            uint
	Title         string
	Description   string
	PricePerNight int
	Bedrooms      int
	Bathrooms     int
	Guests        int
	Country       string
	CountryCode   string
	Category      string
	Image         string
	Landlord      User
	LandlordID    uint
	FavoritedBy   []User
	Reservation   []models.Reservation
}

type Reservation struct {
	ID             uint
	StartDate      time.Time
	EndDate        time.Time
	NumberOfNights int
	Guests         int
	TotalPrice     float64
	CreatedAt      time.Time
	CreatedByID    uint
	PropertyID     uint
	CreatedBy      User
	Property       Property
}

type ConversationMessage struct {
	ID             uint
	Body           string
	ConversationID uint
	CreatedByID    uint
	SenderToID     uint
	Conversation   models.Conversation
	CreatedID      User
	SentTo         User
}

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	//---------------------- Find---------------------------
	FindUserByEmail(email string) (*models.User, error)
	FindUserByToken(token string) (*models.User, error)
	FindPropertyById(Id int) (*models.Property, error)
	FindAllProperties(page int, limit int) ([]models.Property, int64, error)
	FindReservationById(id uint) (*models.Reservation, error)
	FindOrCreateConversation(senderID, receiverID uint) (*models.Conversation, error)
	//-----------------------Create ------------------------
	CreateUser(user User) (*models.User, error)
	CreateProperty(property Property) (*models.Property, error)
	CreateReservation(reservation Reservation) (*models.Reservation, error)
	AddFavoriteProperty(propertyId uint, userId uint) (*models.Property, error)
	CreateConversationMessage(message ConversationMessage) (*models.ConversationMessage, error)
	// --------------------Verify --------------------------
	VerifyUserAndUpdate(token string) (*models.User, error)
	// --------------------Delete---------------------------
	DeleteUser(id string) (*models.User, error)
	DeleteProperty(Id uint) (*models.Property, error)
	DeleteReservation(id uint) (*models.Reservation, error)
	DeleteFavoriteProperty(propertyId uint, userId uint) (*models.Property, error)
	// ---------------------Update--------------------------
	UpdateProperty(Id uint, property Property) (*models.Property, error)

	Close() error
	GetDB() *gorm.DB // Add this method
}

// --------------------------------------------------------------
// --------------------------- Find ------------------------------
// --------------------------------------------------------------
func (s *service) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *service) FindUserByToken(token string) (*models.User, error) {
	var user models.User
	result := s.db.Where("token = ?", token).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *service) FindReservationById(id uint) (*models.Reservation, error) {
	var reservation models.Reservation
	result := s.db.Preload("CreatedBy").
		Preload("Property").
		Where("id = ?", id).First(&reservation)
	if result.Error != nil {
		return nil, result.Error
	}
	return &reservation, nil
}

func (s *service) FindPropertyById(Id int) (*models.Property, error) {
	var property models.Property

	result := s.db.Preload("Landlord").
		Preload("FavoritedBy").
		Preload("Reservations").
		Where("id = ?", Id).First(&property)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return &property, nil
}

func (s *service) FindAllProperties(page int, limit int) ([]models.Property, int64, error) {
	var properties []models.Property
	var total int64

	query := s.db.Model(&models.Property{})
	query.Count(&total)

	if limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	result := query.Preload("Landlord").
		Preload("FavoritedBy").
		Preload("Reservations").
		Find(&properties)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return properties, total, nil
}

// ---------------------------------------
// ----------------- Create ---------------
// ----------------------------------------

func (s *service) CreateConversationMessage(message database.ConversationMessage) (*models.ConversationMessage, error) {
	NewMessage := &models.ConversationMessage{
		Body:           message.Body,
		ConversationID: message.ConversationID,
		CreatedByID:    message.CreatedByID,
		SentToID:       message.SentToID, // Ensure this is set
	}

	result := s.db.Create(NewMessage)
	if result.Error != nil {
		return nil, result.Error
	}
	return NewMessage, nil
}

func (s *service) CreateUser(user User) (*models.User, error) {
	newUser := &models.User{
		Email:    user.Email,
		Password: user.Password,
		Name:     user.Name,
		Avatar:   user.Avatar,
		Token:    user.Token,
	}

	result := s.db.Create(newUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return newUser, nil
}

func (s *service) CreateProperty(property Property) (*models.Property, error) {
	newProperty := &models.Property{
		Title:         property.Title,
		Description:   property.Description,
		PricePerNight: property.PricePerNight,
		Bedrooms:      property.Bedrooms,
		Bathrooms:     property.Bathrooms,
		Guests:        property.Guests,
		Country:       property.Country,
		CountryCode:   property.CountryCode,
		Category:      property.Category,
		Image:         property.Image,
		LandlordID:    property.LandlordID,
	}

	result := s.db.Create(newProperty)
	if result.Error != nil {
		return nil, result.Error
	}

	// Fetch the complete property with associations
	return s.FindPropertyById(int(newProperty.ID))
}

// AddFavoriteProperty adds a property to a user's list of favorite properties.
func (s *service) AddFavoriteProperty(propertyId uint, userId uint) (*models.Property, error) {
	// Step 1: Check if the property exists
	var property models.Property
	result := s.db.Preload("Landlord").Preload("FavoritedBy").First(&property, propertyId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("property with ID %d not found", propertyId)
		}
		return nil, result.Error
	}

	// Step 2: Check if the user exists
	var user models.User
	result = s.db.First(&user, userId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with ID %d not found", userId)
		}
		return nil, result.Error
	}

	// Step 3: Check if the property is already in the user's favorites
	count := s.db.Model(&property).Association("FavoritedBy").Count()
	if count > 0 {
		return nil, fmt.Errorf("property is already in favorites")
	}

	// Step 4: Add the property to the user's favorites
	if err := s.db.Model(&property).Association("FavoritedBy").Append(&user); err != nil {
		return nil, fmt.Errorf("failed to add property to favorites: %v", err)
	}

	// Step 5: Return the updated property
	return &property, nil
}

func (s *service) CreateReservation(reservation Reservation) (*models.Reservation, error) {
	newReservation := &models.Reservation{
		StartDate:      reservation.StartDate,
		EndDate:        reservation.EndDate,
		NumberOfNights: reservation.NumberOfNights,
		Guests:         reservation.Guests,
		TotalPrice:     reservation.TotalPrice,
		CreatedByID:    reservation.CreatedByID,
		PropertyID:     reservation.PropertyID,
	}

	result := s.db.Create(newReservation)
	if result.Error != nil {
		return nil, result.Error
	}

	// Fetch the complete reservation with associations
	var createdReservation models.Reservation
	if err := s.db.Preload("CreatedBy").Preload("Property").First(&createdReservation, newReservation.ID).Error; err != nil {
		return nil, err
	}

	return &createdReservation, nil
}

func (s *service) FindOrCreateConversation(senderID, receiverID uint) (*models.Conversation, error) {
	var conversation models.Conversation

	// Check if a conversation already exists between the two users
	err := s.db.Joins("JOIN conversation_users ON conversation_users.conversation_id = conversations.id").
		Where("conversation_users.user_id IN (?, ?)", senderID, receiverID).
		Group("conversations.id").
		Having("COUNT(DISTINCT conversation_users.user_id) = 2").
		First(&conversation).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// If no conversation exists, create a new one
	if err == gorm.ErrRecordNotFound {
		conversation = models.Conversation{
			Users: []models.User{
				{ID: senderID},
				{ID: receiverID},
			},
		}
		if err := s.db.Create(&conversation).Error; err != nil {
			return nil, err
		}
	}

	return &conversation, nil
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
// -----------------Delete ---------------------------------
// ---------------------------------------------------------

func (s *service) DeleteUser(id string) (*models.User, error) {
	var user models.User

	// Find the user by ID
	result := s.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	// Delete the user
	if err := s.db.Select("FavoritedBy", "Reservations").Delete(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *service) DeleteReservation(id uint) (*models.Reservation, error) {
	var reservation models.Reservation

	result := s.db.Where("id = ?", id).First(&reservation)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	if err := s.db.Delete(reservation).Error; err != nil {
		return nil, err
	}

	return &reservation, nil
}

func (s *service) DeleteProperty(Id uint) (*models.Property, error) {
	var property models.Property

	// Find the user by email
	result := s.db.Where("id = ?", Id).First(&property)
	if result.Error != nil {
		// If no user is found, return nil and error
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // User not found, return nil user
		}
		return nil, result.Error // Return other database errors
	}

	// Delete the user
	if err := s.db.Select("FavoritedBy", "Reservations").Delete(&property).Error; err != nil {
		return nil, err
	}

	// Return the deleted user and nil error
	return s.FindPropertyById(int(property.ID))
}

func (s *service) DeleteFavoriteProperty(propertyId uint, userId uint) (*models.Property, error) {
	// Step 1: Check if the property exists
	var property models.Property
	result := s.db.Preload("Landlord").Preload("FavoritedBy").First(&property, propertyId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("property with ID %d not found", propertyId)
		}
		return nil, result.Error
	}

	// Step 2: Check if the user exists
	var user models.User
	result = s.db.First(&user, userId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user with ID %d not found", userId)
		}
		return nil, result.Error
	}

	// Step 3: Check if the property is in user's favorites
	count := s.db.Model(&property).Association("FavoritedBy").Count()
	if count == 0 {
		return nil, fmt.Errorf("property is not in favorites")
	}

	// Step 4: Remove the property from user's favorites
	if err := s.db.Model(&property).Association("FavoritedBy").Delete(&user); err != nil {
		return nil, fmt.Errorf("failed to remove property from favorites: %v", err)
	}

	// Step 5: Return the updated property
	return &property, nil
}

// --------------------------------------------------------------------------
// ----------------------------------Update----------------------------------
// --------------------------------------------------------------------------

func (s *service) UpdateProperty(Id uint, property Property) (*models.Property, error) {
	var FindProperty models.Property

	// Find the property by ID
	result := s.db.Where("id = ?", Id).First(&FindProperty)
	if result.Error != nil {
		// If no property is found, return nil and error
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Property not found, return nil
		}
		return nil, result.Error // Return other database errors
	}

	// Update the fields using a map
	updates := map[string]interface{}{
		"title":           property.Title,
		"description":     property.Description,
		"price_per_night": property.PricePerNight,
		"bedrooms":        property.Bedrooms,
		"bathrooms":       property.Bathrooms,
		"guests":          property.Guests,
		"country":         property.Country,
		"country_code":    property.CountryCode,
		"category":        property.Category,
		"image":           property.Image,
		"landlord_id":     property.LandlordID,
	}

	// Update the property in the database
	if err := s.db.Model(&FindProperty).Updates(updates).Error; err != nil {
		return nil, err // Return error if update fails
	}

	// Return the updated property
	return &FindProperty, nil
}

type service struct {
	db *gorm.DB
}

func (s *service) GetDB() *gorm.DB {
	return s.db
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
		&models.Conversation{},
		&models.Property{},
		&models.Reservation{},
		&models.ConversationMessage{},
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
