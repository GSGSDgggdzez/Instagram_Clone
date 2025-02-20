package database

import (
	models "API/internal/Models"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
	//-----------------------Create ------------------------
	CreateUser(user models.User) (*models.User, error)
	CreateNotification(user models.User, notification models.Notification) (*models.User, error)
	// --------------------Verify --------------------------
	VerifyUserAndUpdate(token string) (*models.User, error)
	// --------------------Delete---------------------------
	DeleteUser(id string) (*models.User, error)
	// --------------------Update---------------------------
	UpdateUser(user models.User) (*models.User, error)
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

func (s *service) FindUserById(id uint) (*models.User, error) {

	var user models.User
	result := s.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// --------------------------------------------------------------
// --------------------------- Create ------------------------------
// --------------------------------------------------------------

func (s *service) CreateUser(user models.User) (*models.User, error) {
	newUser := &models.User{
		Email:          user.Email,
		Password:       user.Password,
		Name:           user.Name,
		Username:       user.Username,
		Avatar:         user.Avatar,
		Token:          user.Token,
		Bio:            user.Bio,
		Website:        user.Website,
		EmailVerified:  user.EmailVerified,
		Phone:          user.Phone,
		FollowerCount:  user.FollowerCount,
		FollowingCount: user.FollowingCount,
		PostCount:      user.PostCount,
		Privacy:        user.Privacy,
		IsVerified:     user.IsVerified,
		Language:       user.Language,
	}

	result := s.db.Create(newUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return newUser, nil
}

func (s *service) CreateNotification(user models.User, notification models.Notification) (*models.User, error) {
	NewNotification := &models.Notification{
		From:     notification.From,
		To:       notification.To,
		Type:     notification.Type,
		Context:  notification.Context,
		Priority: notification.Priority,
		GroupID:  notification.GroupID,
		Read:     false,
	}

	result := s.db.Create(NewNotification)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// --------------------------------------------------------------
// --------------------------- VerifyUserAndUpdate ------------------------------
// --------------------------------------------------------------

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

// --------------------------------------------------------------
// --------------------------- Delete ------------------------------
// --------------------------------------------------------------

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

	// Delete all related data using Select
	if err := s.db.Select([]string{
		"Posts",
		"Likes",
		"Comments",
		"Stories",
		"Highlights",
		"SavedPosts",
		"Follows",
	}).Delete(&user).Error; err != nil {
		return nil, err
	}

	// Clean up any follows where this user is being followed
	if err := s.db.Where("followed_id = ?", id).Delete(&models.Follow{}).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// --------------------------------------------------------------
// --------------------------- Update ------------------------------
// --------------------------------------------------------------

func (s *service) UpdateUser(user models.User) (*models.User, error) {
	result := s.db.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s search_path=%s pool_max_conns=25 pool_min_conns=5", host, username, password, database, port, sslmode, schema)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true, // Speed boost!
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(25) // More realistic
	sqlDB.SetMaxIdleConns(50) // Better idle count
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

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
		&models.Story{},
		&models.Highlight{},
		&models.Follow{},
		&models.Notification{},
		&models.Hashtag{},
	)
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	sqlDB, err := s.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err)
		return stats
	}

	// Ping the database
	err = sqlDB.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats
	dbStats := sqlDB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 {
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
