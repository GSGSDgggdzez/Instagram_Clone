package controllers

import (
	models "API/internal/Models"
	"API/internal/config"
	"API/internal/database"
	"API/internal/utils"
	"context"
	"errors"
	"fmt"
	"html"
	"log"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

type AuthController struct {
	db       database.Service    // The database service to interact with the database.
	validate *validator.Validate // Validator instance for validating user inputs.
}

func NewAuthController(db database.Service) *AuthController {
	return &AuthController{
		db:       db,              // Setting the provided database service.
		validate: validator.New(), // Initializing a new validator instance.
	}
}

var (
	Limiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 100) // More permissive for bursts
)

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the registration logic -------------------------
// ---------------------------------------------------------------------------------------------------

type RegisterRequest struct {
	Username string `form:"username" validate:"required,max=30"`
	Name     string `form:"name" validate:"required,max=255"`
	Email    string `form:"email" validate:"required,email,max=255"`
	Password string `form:"password" validate:"required,max=255,min=8"`
	Bio      string `form:"bio" validate:"max=150"`
	Website  string `form:"website" validate:"omitempty,url,max=255"`
	Phone    string `form:"phone" validate:"required,max=255"`
	Language string `form:"language" validate:"required,max=20"`
	Privacy  bool   `form:"privacy" validate:"omitempty,oneof=public private"`
}

func (ac *AuthController) Register(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if !Limiter.Allow() {
		return utils.SendErrorResponse(c, fiber.StatusTooManyRequests, "Too many registration attempts", nil)
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	uploadChan := make(chan struct {
		url string
		err error
	})

	var avatarURL string

	file, err := c.FormFile("avatar")

	if err != nil {
		avatarURL = utils.GetDefaultAvatar()
	} else {
		// Get Cloudinary instance before goroutine
		cld, err := config.InitCloudinary()
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to initialize Cloudinary", err.Error())
		}

		go func(file *multipart.FileHeader) {
			var result struct {
				url string
				err error
			}

			if err := utils.ValidateImageFile(file); err != nil {
				result.err = err
				uploadChan <- result
				return
			}

			fileHeader, err := file.Open()
			if err != nil {
				result.err = err
				uploadChan <- result
				return
			}
			defer fileHeader.Close()

			url, err := utils.UploadToCloudinary(cld, ctx, fileHeader)
			if err != nil {
				result.err = err
				uploadChan <- result
				return
			}

			result.url = url
			uploadChan <- result

		}(file)

		uploadResult := <-uploadChan
		if uploadResult.err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to upload image", uploadResult.err.Error())
		}
		avatarURL = uploadResult.url
	}

	// Use the existing context
	// ctx := context.Background()

	// Handle upload in goroutine

	token, err := utils.GenerateVerificationToken()
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate verification token", err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	createUserData := models.User{
		Name:     html.EscapeString(req.Name),
		Email:    html.EscapeString(req.Email),
		Password: string(hashedPassword),
		Bio:      html.EscapeString(req.Bio),
		Avatar:   avatarURL,
		// Avatar:   utils.GetDefaultAvatar(), // HERE'S THE FIRE ADDITION! 🔥
		Website:  html.EscapeString(req.Website),
		Phone:    html.EscapeString(req.Phone),
		Language: html.EscapeString(req.Language),
		Privacy:  req.Privacy,
		Username: html.EscapeString(req.Username),
		Token:    token,
	}

	var newUser models.User
	// Wrap the user creation in a transaction
	err = ac.db.GetDB().Transaction(func(tx *gorm.DB) error {
		// Create user
		var err error
		var userPtr *models.User
		// Retry mechanism for user creation
		for attempts := 1; attempts <= 3; attempts++ {
			userPtr, err = ac.db.CreateUser(createUserData)
			go func() {
				utils.TrackRegistration(req.Email, true, attempts)
			}()

			if err == nil {
				break
			}
			time.Sleep(time.Second * time.Duration(attempts))
		}
		if err != nil {
			return err
		}
		newUser = *userPtr

		// Send verification email asynchronously
		// go func() {
		// 	if err := utils.SendVerificationEmail(newUser.Email, newUser.Token); err != nil {
		// 		log.Printf("Failed to send verification email: %v", err)
		// 	}
		// }()

		return nil
	})

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Registration failed", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully, please check your email for verification",
		"status":  fiber.StatusCreated,
		"userID":  newUser.Email,
	})

}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the start of the Verify Email logic -------------------------
// --------------------------------------------------------------------------------------------------

func (ac *AuthController) VerifyEmail(c *fiber.Ctx) error {
	GodTOken := html.EscapeString(c.Params("token"))

	var updateResult models.User
	err := ac.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var err error
		var userPtr *models.User

		for attempts := 1; attempts <= 3; attempts++ {
			userPtr, err = ac.db.VerifyUserAndUpdate(GodTOken)
			utils.TrackFindUserByToken(userPtr.Email, true, attempts)
			if err == nil {
				break
			}
			time.Sleep(time.Second * time.Duration(attempts))
		}
		if err != nil {
			return err
		}

		updateResult = *userPtr

		return nil
	})

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Bad Verification token", err.Error())
	}

	JWT, err := utils.GenerateToken(updateResult.ID, updateResult.Email, updateResult.Name, updateResult.Username, updateResult.Avatar, updateResult.Token, updateResult.Bio, updateResult.Website, updateResult.Phone, updateResult.Language, updateResult.EmailVerified, updateResult.Privacy, updateResult.EmailVerified, updateResult.FollowingCount, updateResult.FollowerCount)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate JWT token", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"detail": "User have been verify successfully",
		"status": fiber.StatusOK,
		"JWT":    JWT,
		"user": fiber.Map{
			"User": updateResult,
		},
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the Verify Email logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the Start of the LoginRequest logic -------------------------
// --------------------------------------------------------------------------------------------------

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`    // Must be a valid email
	Password string `json:"password" validate:"required,max=255,min=8"` // Cannot be empty
}

func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if !Limiter.Allow() {
		return utils.SendErrorResponse(c, fiber.StatusTooManyRequests, "Too many registration attempts", nil)
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	var user *models.User
	err := ac.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var err error

		for attempts := 1; attempts <= 3; attempts++ {
			user, err = ac.db.FindUserByEmail(html.EscapeString(req.Email))
			utils.TrackFindUserByEMAIL(req.Email, true, attempts)
			if err == nil {
				break
			}
			time.Sleep(time.Second * time.Duration(attempts))
		}

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error during login", utils.FormatValidationErrors(err))
	}

	if user == nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error during login", utils.FormatValidationErrors(err))
	}
	JWT, err := utils.GenerateToken(user.ID, user.Email, user.Name, user.Username, user.Avatar, user.Token, user.Bio, user.Website, user.Phone, user.Language, user.EmailVerified, user.Privacy, user.EmailVerified, user.FollowingCount, user.FollowerCount)

	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate JWT token", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"detail": "User have been verify successfully",
		"status": fiber.StatusOK,
		"JWT":    JWT,
		"user": fiber.Map{
			"User": user,
		},
	})
}

// --------------------------------------------------------------------------------------------------
//------------------------------ these is the End of the LoginRequest logic -------------------------
// --------------------------------------------------------------------------------------------------

// --------------------------------------------------------------------------------------------------
// ------------------------------ these is the start of the ForgotPasswordRequest logic -------------------------
// --------------------------------------------------------------------------------------------------

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"` // Must be a valid email
}

func (ac *AuthController) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest

	if !Limiter.Allow() {
		return utils.SendErrorResponse(c, fiber.StatusTooManyRequests, "Too many registration attempts", nil)
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	var user *models.User
	err := ac.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var err error

		for attempts := 1; attempts <= 3; attempts++ {
			user, err = ac.db.FindUserByEmail(html.EscapeString(req.Email))
			utils.TrackFindUserByEMAIL(req.Email, true, attempts)
			if err == nil {
				break
			}
			time.Sleep(time.Second * time.Duration(attempts))
		}

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to check user existence", err.Error())
	}

	go func() {
		// Assume you have a function that sends the email
		err := utils.SendVerificationPassword(user.Email, user.Token)
		if err != nil {
			// Return an error response if sending the verification email fails.
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to send verification email", // Error message.
				"status": fiber.StatusInternalServerError,     // Internal server error status code.
				"TO":     user.Email,
			})
		}
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Reset Password Email have been send please check you email", // Success message.
		"status":  fiber.StatusCreated,                                          // Created status code.
	})
}

// -----------------------------------------------------------------------------------------------------
// ------------------------------ these is the End of the ForgotPasswordRequest logic -------------------------
// --------------------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the start  of the RestPasswordRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

func (ac *AuthController) RestPassword(c *fiber.Ctx) error {

	Token := html.EscapeString(c.Params("Token"))

	var user *models.User
	err := ac.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var err error

		for attempts := 1; attempts <= 3; attempts++ {
			user, err = ac.db.FindUserByToken(Token)
			utils.TrackFindUserByToken(user.Email, true, attempts)
			if err == nil {
				break
			}
			time.Sleep(time.Second * time.Duration(attempts))
		}

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil && err != gorm.ErrRecordNotFound {
		// Return an error response if there is an error other than record not found.
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if user == nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	JWT, err := utils.GenerateToken(user.ID, user.Email, user.Name, user.Username, user.Avatar, user.Token, user.Bio, user.Website, user.Phone, user.Language, user.EmailVerified, user.Privacy, user.EmailVerified, user.FollowingCount, user.FollowerCount)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate JWT token", err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{

		"detail": "User have been verify successfully",
		"status": fiber.StatusOK,
		"JWT":    JWT,
		"user": fiber.Map{
			"User": user,
		},
	})

}

// ----------------------------------- -------------------------------------------------------------------------
// ------------------------------ these is the End of the RestPasswordRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the Start of the DeleteRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

// DeleteUser - Clean and secure user deletion
func (ac *AuthController) DeleteUser(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*utils.Claims)
	if !ok || claims == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or missing authentication", nil)
	}

	ID := html.EscapeString(c.Params("ID"))

	if claims.ID != ID {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "You are not authorized to delete this account", nil)
	}

	// Convert string ID to uint
	userID, err := strconv.ParseUint(claims.ID, 10, 32)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format", err.Error())
	}

	// deleteUser, err := ac.db.FindUserById(uint(userID))

	var deleteUser *models.User
	err = ac.db.GetDB().Transaction(func(tx *gorm.DB) error {
		var err error

		for attempts := 1; attempts <= 3; attempts++ {
			deleteUser, err = ac.db.FindUserById(uint(userID))
			utils.TrackFindUserByID(claims.Email, true, attempts)
			if err == nil {
				break
			}
			time.Sleep(time.Second * time.Duration(attempts))
		}

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Database error", err.Error())
	}

	if deleteUser == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Create notifications for followers before deletion
	notification := models.Notification{
		From:     deleteUser.ID,
		To:       deleteUser.ID, // We'll update this for each follower
		Type:     "account_deletion",
		Context:  fmt.Sprintf("%s has deleted their account", deleteUser.Username),
		Priority: 1,
		GroupID:  fmt.Sprintf("deletion_%d", deleteUser.ID),
	}

	var Delete *models.User
	// Transaction to handle deletion and notifications
	err = ac.db.GetDB().Transaction(func(tx *gorm.DB) error {
		// Create notifications for followers
		for attempts := 1; attempts <= 3; attempts++ {
			_, err := ac.db.CreateNotification(*deleteUser, notification)
			if err == nil {
				break
			}
			time.Sleep(time.Second * time.Duration(attempts))
		}

		// Proceed with user deletion
		_, err := ac.db.DeleteUser(claims.ID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())
	}

	// Handle avatar deletion in background if it exists
	if deleteUser.Avatar != "" {
		go func() {
			publicID, err := utils.ExtractPublicID(deleteUser.Avatar)
			if err != nil {
				log.Printf("Error extracting public ID: %v", err)
				return
			}

			cld, err := config.InitCloudinary()
			if err != nil {
				log.Printf("Error initializing Cloudinary: %v", err)
				return
			}

			if err := utils.DeleteImageFromCloudinary(cld, publicID); err != nil {
				log.Printf("Error deleting image from Cloudinary: %v", err)
			}
		}()
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User successfully deleted",
		"user": fiber.Map{
			"User": Delete,
		},
	})
}

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the End of the DeleteRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the Start of the EditRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------

type EditUserRequest struct {
	Name     string `form:"name" validate:"omitempty,max=255"`
	Email    string `form:"email" validate:"omitempty,email,max=255"`
	Password string `form:"password" validate:"omitempty,max=255,min=8"`
	Bio      string `form:"bio" validate:"omitempty,max=255"`
}

func (ac *AuthController) EditUser(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	var req EditUserRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid form data", err.Error())
	}

	if err := ac.validate.Struct(req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
	}

	// Get existing user
	existingUser, err := ac.db.FindUserById(uint(claims.UserID))
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	// Handle avatar upload if provided
	file, err := c.FormFile("avatar")
	if err == nil {
		uploadChan := make(chan struct {
			url string
			err error
		})

		cld, err := config.InitCloudinary()
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to initialize Cloudinary", err.Error())
		}

		ctx := context.Background()

		go func(file *multipart.FileHeader) {
			var result struct {
				url string
				err error
			}

			if err := utils.ValidateImageFile(file); err != nil {
				result.err = err
				uploadChan <- result
				return
			}

			fileHeader, err := file.Open()
			if err != nil {
				result.err = err
				uploadChan <- result
				return
			}
			defer fileHeader.Close()

			// Delete old avatar if exists
			if existingUser.Avatar != "" {
				publicID, err := utils.ExtractPublicID(existingUser.Avatar)
				if err == nil {
					utils.DeleteImageFromCloudinary(cld, publicID)
				}
			}

			url, err := utils.UploadToCloudinary(cld, ctx, fileHeader)
			if err != nil {
				result.err = err
				uploadChan <- result
				return
			}

			result.url = url
			uploadChan <- result
		}(file)

		uploadResult := <-uploadChan
		if uploadResult.err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to upload image", uploadResult.err.Error())
		}
		existingUser.Avatar = uploadResult.url
	}

	// Update user fields if provided
	if req.Name != "" {
		existingUser.Name = html.EscapeString(req.Name)
	}
	if req.Email != "" {
		existingUser.Email = html.EscapeString(req.Email)
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
		}
		existingUser.Password = string(hashedPassword)
	}
	if req.Bio != "" {
		existingUser.Bio = html.EscapeString(req.Bio)
	}

	// Update user in database
	updatedUser, err := ac.db.UpdateUser(*existingUser)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User updated successfully",
		"status":  fiber.StatusOK,
		"user": fiber.Map{
			"id":     updatedUser.ID,
			"name":   updatedUser.Name,
			"email":  updatedUser.Email,
			"bio":    updatedUser.Bio,
			"avatar": updatedUser.Avatar,
		},
	})
}

// ------------------------------------------------------------------------------------------------------------
// ------------------------------ these is the End of the EditRequest logic -------------------------
// ------------------------------------------------------------------------------------------------------------
