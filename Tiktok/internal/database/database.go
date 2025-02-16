package database

import (
	"Tiktok/internal/models"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID            uint
	Name          string
	Bio           string
	Avatar        string
	Email         string
	EmailVerified bool
	Password      string
	Token         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Posts         []models.Post
	Comments      []models.Comment
	Likes         []models.Like
}

type Post struct {
	ID        uint
	UserID    User
	User      User
	Text      string
	Video     string
	Duration  float64
	IsPrivate bool
	Music     string
	Location  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Comments  []models.Comment
	Likes     []models.Like
}

type Like struct {
	ID        uint
	UserID    uint
	PostID    uint
	User      User
	Post      Post
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Comment struct {
	ID        uint
	UserID    uint
	PostID    uint
	User      User
	Post      Post
	Text      string
	CreatedAt time.Time
}

type Follow struct {
	ID          uint
	FollowerID  uint
	FollowingID uint
	Follower    User
	Following   User
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

type Gift struct {
	ID           uint
	UserID       uint
	LivestreamID uint
	User         User
	LiveStream   LiveStream
	GiftType     string
	Amount       float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Hashtag struct {
	ID       uint
	Name     string
	Post     []Post
	CreateAt time.Time
	UpdateAt time.Time
}

type LiveStream struct {
	ID          uint
	UserID      uint
	User        User
	Title       string
	Description string
	StreamKey   string
	Status      string
	ViewerCount uint
	StartedAt   time.Time
	EndedAt     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Comment     []Comment
	Gift        []Gift
}

type LiveStreamComment struct {
	ID           uint
	UserID       uint
	LiveStreamID uint
	User         User
	LiveStream   LiveStream
	Text         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Notification struct {
	ID        uint
	UserID    uint
	FromID    uint
	User      User
	From      User
	Type      string
	Content   string
	Read      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
	GetDB() *gorm.DB // Add this method

	//---------------------- Find---------------------------
	FindUserByEmail(email string) (*models.User, error)
	FindUserByToken(token string) (*models.User, error)
	FindUserById(id uint) (*models.User, error)
	FindPostById(id uint) (*models.Post, error)
	FindFollowByUsers(followerID, followingID uint) (*models.Follow, error)
	FindLikeByUserAndPost(userID, postID uint) (*models.Like, error)
	FindCommentById(id uint) (*models.Comment, error)

	//-----------------------Create ------------------------
	CreateUser(user User) (*models.User, error)
	CreatePost(post Post, hashtags []string) (*models.Post, error)
	CreateLike(like Like) (*models.Like, error)
	CreateLiveStream(stream *models.LiveStream) error
	CreateFollow(follow Follow) (*models.Follow, error)
	CreateComment(comment Comment) (*models.Comment, error)
	// --------------------Verify --------------------------
	VerifyUserAndUpdate(token string) (*models.User, error)
	// --------------------Delete---------------------------
	DeleteUser(id string) (*models.User, error)
	DeletePost(postID uint) error
	DeleteLike(likeID uint) error
	DeleteComment(commentID uint) error
	DeleteFollow(followID uint) error
	// --------------------Update---------------------------
	UpdateUser(user models.User) (*models.User, error)
	UpdatePost(post Post, hashtags []string) (*models.Post, error)

	UpdateComment(comment models.Comment) (*models.Comment, error)
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

func (s *service) FindFollowByUsers(followerID, followingID uint) (*models.Follow, error) {
	var follow models.Follow
	err := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&follow).Error
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (s *service) FindUserByToken(token string) (*models.User, error) {
	var user models.User
	result := s.db.Where("token = ?", token).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *service) FindUserById(id uint) (*models.User, error) {

	var user models.User
	result := s.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (s *service) FindPostById(id uint) (*models.Post, error) {
	var post models.Post

	result := s.db.Preload("User").
		Preload("Comments").      // Changed from "Comment" to "Comments"
		Preload("Comments.User"). // Add this to load comment authors
		Preload("Likes").         // Changed from "Like" to "Likes"
		Preload("Likes.User").    // Add this to load like authors
		Preload("Hashtags").      // Changed from "Hashtag" to "Hashtags"
		Where("id = ?", id).
		First(&post)

	if result.Error != nil {
		return nil, result.Error
	}

	return &post, nil
}

func (s *service) FindLikeByUserAndPost(userID, postID uint) (*models.Like, error) {
	var like models.Like
	err := s.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (s *service) FindCommentById(id uint) (*models.Comment, error) {
	var Comment models.Comment
	err := s.db.Where("id = ?", id).First(&Comment).Error
	if err != nil {
		return nil, err
	}
	return &Comment, nil
}

// ---------------------------------------
// ----------------- Create ---------------
// ----------------------------------------

func (s *service) CreateFollow(follow Follow) (*models.Follow, error) {
	newFollow := &models.Follow{
		FollowerID:  follow.FollowerID,
		FollowingID: follow.FollowingID,
	}

	if err := s.db.Create(newFollow).Error; err != nil {
		return nil, err
	}
	return newFollow, nil
}

func (s *service) CreateUser(user User) (*models.User, error) {
	newUser := &models.User{
		Email:    user.Email,
		Password: user.Password,
		Name:     user.Name,
		Avatar:   user.Avatar,
		Token:    user.Token,
		Bio:      user.Bio,
	}

	result := s.db.Create(newUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return newUser, nil
}

// 📝 Add this implementation
func (s *service) CreatePost(post Post, hashtags []string) (*models.Post, error) {
	// Start a transaction - because we're serious about data!
	tx := s.db.Begin()

	// 📦 Create the new posts
	newPost := &models.Post{
		UserID:    post.UserID.ID,
		Text:      post.Text,
		Video:     post.Video,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(newPost).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 🏷️ Handle the hashtags
	for _, tag := range hashtags {
		var hashtag models.Hashtag

		// Try to find existing hashtag or create new one
		if err := tx.Where("name = ?", tag).FirstOrCreate(&hashtag, models.Hashtag{
			Name: tag,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// 🔗 Link post and hashtag
		if err := tx.Exec("INSERT INTO post_hashtags (post_id, hashtag_id) VALUES (?, ?)",
			newPost.ID, hashtag.ID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 🎉 Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.FindPostById(newPost.ID)
}

func (s *service) CreateLike(like Like) (*models.Like, error) {
	newLike := &models.Like{
		UserID: like.UserID,
		PostID: like.PostID,
	}

	if err := s.db.Create(newLike).Error; err != nil {
		return nil, err
	}
	return newLike, nil
}

func (s *service) CreateComment(comment Comment) (*models.Comment, error) {
	newComment := &models.Comment{
		UserID: comment.UserID,
		PostID: comment.PostID,
		Text:   comment.Text,
	}

	if err := s.db.Create(newComment).Error; err != nil {
		return nil, err
	}
	return newComment, nil
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
		"EmailVerified": true,
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
	if err := s.db.Select("Post").Delete(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *service) DeleteFollow(followID uint) error {
	return s.db.Delete(&models.Follow{}, followID).Error
}

func (s *service) DeletePost(postID uint) error {
	tx := s.db.Begin()

	// Delete associated records first
	if err := tx.Where("post_id = ?", postID).Delete(&models.Comment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("post_id = ?", postID).Delete(&models.Like{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Remove hashtag associations
	if err := tx.Model(&models.Post{ID: postID}).Association("Hashtags").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	// Delete the post
	if err := tx.Delete(&models.Post{}, postID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *service) DeleteLike(likeID uint) error {
	return s.db.Delete(&models.Like{}, likeID).Error
}

func (s *service) DeleteComment(commentID uint) error {
	return s.db.Delete(&models.Comment{}, commentID).Error
}

// ---------------------------------------------------------
// -----------------Update ---------------------------------
// ---------------------------------------------------------
func (s *service) UpdateUser(user models.User) (*models.User, error) {
	result := s.db.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *service) UpdateComment(comment models.Comment) (*models.Comment, error) {
	result := s.db.Save(&comment)

	if result.Error != nil {
		return nil, result.Error
	}

	return &comment, nil
}

func (s *service) UpdatePost(post Post, hashtags []string) (*models.Post, error) {
	tx := s.db.Begin()

	// Update main post data
	if err := tx.Model(&models.Post{}).Where("id = ?", post.ID).Updates(map[string]interface{}{
		"text":       post.Text,
		"is_private": post.IsPrivate,
		"music":      post.Music,
		"location":   post.Location,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Remove existing hashtag associations
	if err := tx.Model(&models.Post{ID: post.ID}).Association("Hashtags").Clear(); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Add new hashtags
	for _, tagName := range hashtags {
		var hashtag models.Hashtag

		if err := tx.Where(models.Hashtag{Name: tagName}).FirstOrCreate(&hashtag).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := tx.Model(&models.Post{ID: post.ID}).Association("Hashtags").Append(&hashtag); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.FindPostById(post.ID)
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
		&models.Like{},
		&models.Post{},
		&models.Comment{},
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
