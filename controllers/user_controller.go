package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/database"
	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/models"
	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection("User")

func HashPassword(password string) (string, error) {
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(HashPassword), nil
}

func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		userMade := make(chan bool, 1) // Add buffer to prevent blocking
		errorChan := make(chan error, 1)
		if err := c.ShouldBindJSON(&user); err != nil {
			fmt.Printf("JSON binding error: %v\n", err)
			c.JSON(400, gin.H{"error": "Invalid JSON"})
			return
		}

		fmt.Printf("Registration attempt for email: %s\n", user.Email)

		hashedPassword, err := HashPassword(user.Password)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		go func() {

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

			if err != nil {
				errorChan <- err
				return
			}
			if count > 0 {
				errorChan <- fmt.Errorf("User with this email already exists")
				return
			}
			user.UserID = bson.NewObjectID().Hex()
			user.CreatedAt = time.Now()
			user.UpdatedAt = time.Now()
			user.Password = hashedPassword
			var validate = validator.New()
			if err := validate.Struct(user); err != nil {
				errorChan <- err
				return
			}

			_, err = userCollection.InsertOne(ctx, user)
			if err != nil {
				errorChan <- err // Send error to channel
				return
			}
			userMade <- true
		}()
		select {
		case <-userMade:
			// Success case - generate token and return user data
			token, err := utils.GenerateAccessToken(user.Email, user.FirstName, user.LastName, user.Role, user.UserID)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to generate token"})
				return
			}

			userResponse := gin.H{
				"token": token,
				"user": gin.H{
					"id":             user.UserID,
					"username":       user.FirstName + " " + user.LastName,
					"email":          user.Email,
					"role":           user.Role,
					"favoriteGenres": user.FavouriteGenres,
				},
			}
			c.JSON(http.StatusCreated, userResponse)
		case err := <-errorChan:
			// Handle different error types
			if err.Error() == "User with this email already exists" {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(500, gin.H{"error": err.Error()})
			}
		case <-time.After(10 * time.Second):
			// Timeout case
			c.JSON(408, gin.H{"error": "request timeout"})
		}
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user models.User

		var userLogin models.UserLogin

		var response models.UserResponse

		userToken := make(chan models.UserResponse, 1)

		errorChan := make(chan error, 1)

		if err := c.ShouldBindJSON(&userLogin); err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON"})
			return
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			result := userCollection.FindOne(ctx, bson.M{"email": userLogin.Email})
			err := result.Decode(&user)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					// User not found
					errorChan <- fmt.Errorf("Invalid credentials")
					return
				}

			}

			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userLogin.Password))
			if err != nil {
				errorChan <- fmt.Errorf("Wrong Password")
				return
			}
			token, err := utils.GenerateAccessToken(user.Email, user.FirstName, user.LastName, user.Role, user.UserID)
			if err != nil {
				errorChan <- err
				return
			}
			refreshToken, err := utils.GenerateRefreshToken(user.Email, user.FirstName, user.LastName, user.Role, user.UserID)
			if err != nil {
				errorChan <- err
				return
			}
			err = utils.UpdateAllTokens(user.UserID, token, refreshToken)
			if err != nil {
				errorChan <- err
				return
			}
			response.UserID = user.UserID
			response.FirstName = user.FirstName
			response.LastName = user.LastName
			response.Email = user.Email
			response.Role = user.Role
			response.FavouriteGenres = user.FavouriteGenres
			response.Token = token
			response.RefreshToken = refreshToken
			userToken <- response

		}()
		select {
		case response := <-userToken:
			c.JSON(200, gin.H{"response": response})
		case err := <-errorChan:
			c.JSON(401, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			c.JSON(408, gin.H{"error": "timeout"})
		}
	}
}
