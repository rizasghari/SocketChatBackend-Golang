package repositories

import (
	"log"
	"socketChat/internal/errs"
	"socketChat/internal/models"
	"socketChat/internal/utils"

	"gorm.io/gorm"
)

type AuthenticationRepository struct {
	db *gorm.DB
}

func NewAuthenticationRepository(db *gorm.DB) *AuthenticationRepository {
	return &AuthenticationRepository{
		db: db,
	}
}

func (ar *AuthenticationRepository) CreateUser(user *models.User) (*models.User, []error) {
	var errors []error
	result := ar.db.Create(user)
	if result.Error != nil {
		errors = append(errors, result.Error)
		return nil, errors
	}
	if result.RowsAffected == 0 {
		errors = append(errors, errs.ErrUserNotFound)
		return nil, errors
	}
	return user, nil
}

func (ar *AuthenticationRepository) CheckIfUserExists(email string) *models.User {
	var user models.User
	result := ar.db.Where("email = ?", email).First(&user)
	if result.Error == nil && result.RowsAffected > 0 {
		return &user
	}
	return nil
}

func (ar *AuthenticationRepository) Login(login *models.LoginRequestBody) (*models.User, []error) {
	var errors []error
	var user *models.User
	user = ar.CheckIfUserExists(login.Email)
	if user == nil {
		errors = append(errors, errs.ErrUserNotFound)
		return nil, errors
	}
	log.Printf("Password: %v Hash: %v", login.Password, user.PasswordHash)
	if err := utils.CompareHashAndPassword(user.PasswordHash, login.Password); err != nil {
		errors = append(errors, errs.ErrWrongPassword)
		return nil, errors
	}
	return user, nil
}

func (ar *AuthenticationRepository) GetAllUsersWithPagination(page, limit, offset int) (*models.GetUsersResponse, []error) {
	var users []models.User
	var userResponses []models.UserResponse
	var errors []error
	var total int64

	transactionErr := ar.db.Transaction(func(tx *gorm.DB) error {

		log.Println("AuthenticationRepository Page: ", page, " Limit: ", limit, " Offset: ", offset)

		result := tx.
			Offset(offset).
			Limit(limit).
			Select([]string{"ID", "first_name", "last_name", "profile_photo", "is_online", "last_seen"}).
			Find(&users).
			Where("deleted_at IS NULL")

		if err := result.Error; err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return errs.ErrThereIsNoUser
		}
		err := ar.db.Model(&models.User{}).Where("deleted_at IS NULL").Count(&total).Error
		if err != nil {
			return err
		}

		return nil
	})

	if transactionErr != nil {
		log.Println("Transaction error: ", transactionErr)
		errors = append(errors, transactionErr)
		return nil, errors
	}

	for _, user := range users {
		userResponses = append(userResponses, user.ToUserResponse())
	}

	usersResponse := &models.GetUsersResponse{
		Users: userResponses,
		Page:  page,
		Size:  limit,
		Total: total,
	}

	return usersResponse, errors
}
