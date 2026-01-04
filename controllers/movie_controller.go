package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/database"
	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// GIN FRAMEWORK EXPLANATION (coming from Express.js):
// ===================================================
// Gin = Express for Go (web framework for building APIs)
//
// Express.js equivalent:
// app.get('/movies', (req, res) => {
//   res.json({ movies: [] });
// });
//
// gin.Context (the 'c' parameter):
// - Combines Express's 'req' + 'res' into one object
// - c.JSON() = res.json()
// - c.Param("id") = req.params.id
// - c.Query("search") = req.query.search
// - c.BindJSON(&movie) = req.body parsing
//
// gin.HandlerFunc:
// - Express equivalent: (req, res) => {}
// - Go signature: func(c *gin.Context)
//
// Why return a function?
// Go pattern for middleware/configuration - like:
// const getMovies = () => (req, res) => { /* handler logic */ };

// CONTEXT, TIMEOUT, AND CURSOR EXPLANATION (coming from Node.js):
// ===============================================================
// In Node.js:
// const movies = await Movie.find(); // Simple! No timeouts/contexts needed
//
// In Go: Much more explicit about:
// 1. CONTEXT = request lifecycle management
// 2. TIMEOUTS = preventing hanging operations
// 3. CURSORS = manual result iteration (no auto-parsing)
// 4. DEFER = cleanup guarantee (like finally block)

var movieCollection *mongo.Collection = database.OpenCollection("Movie")
var movieValidate = validator.New()

func GetMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GOROUTINE + CHANNEL PATTERN (async-like behavior):
		// ================================================
		// This pattern mimics async/await from Node.js
		// goroutine = async function, channels = await mechanism

		// Create channels to receive results
		moviesChan := make(chan []models.Movie, 1)
		errorChan := make(chan error, 1)

		// Start goroutine (like async function)
		go func() {
			// CONTEXT EXPLANATION:
			// ctx = context (request lifecycle, NOT gin.Context)
			// cancel = function to stop the operation
			// WithTimeout = "kill this operation after 100 seconds"
			// Node.js equivalent: setTimeout to cancel operation
			// both etx and cancel are returned here
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

			// DEFER EXPLANATION:
			// defer = run this when function exits (like finally block)
			// Ensures cancel() ALWAYS runs even if error occurs
			// breaks the timer to make sure it isnt running
			// Node.js equivalent: try/finally { cancel() }
			defer cancel()

			// Prepare array to store results
			// Node.js equivalent: let movies = [];
			var movies []models.Movie

			// FIND OPERATION:
			// Find returns a CURSOR (pointer to results), not actual data
			// bson.M{} = empty filter (find all)
			// Node.js equivalent: Movie.find({}) but returns cursor instead of data
			cursor, err := movieCollection.Find(ctx, bson.M{})

			if err != nil {
				errorChan <- err // Send error to channel
				return
			}

			// CURSOR CLEANUP:
			// defer ensures cursor.Close() runs when function exits
			// Prevents memory leaks - VERY important!
			// closes connection to mongodb
			// Node.js equivalent: MongoDB driver handles this automatically
			defer cursor.Close(ctx)

			// CURSOR.ALL EXPLANATION:
			// cursor.All() = convert cursor to actual data array
			// &movies = pass memory address so function can fill it
			// Node.js equivalent: const movies = await cursor.toArray();
			// shorthand for err = cursor.All(ctx, &movies)
			//if err != nil {
			// Handle error
			//}
			if err = cursor.All(ctx, &movies); err != nil {
				errorChan <- err // Send error to channel
				return
			}

			moviesChan <- movies // Send result to channel
		}()

		// CHANNEL SELECT (like await):
		// ===========================
		// select = wait for first channel to receive data
		// Like Promise.race() in Node.js
		select {
		case movies := <-moviesChan:
			// Success case - got movies from channel
			c.JSON(200, movies)
		case err := <-errorChan:
			// Error case - got error from channel
			c.JSON(500, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			// Timeout case - neither channel responded in time
			c.JSON(408, gin.H{"error": "request timeout"})
		}
	}
}

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		moviesChan := make(chan []models.Movie, 1)
		errorChan := make(chan error, 1)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			var movies []models.Movie
			id := c.Param("imdb_id")
			cursor, err := movieCollection.Find(ctx, bson.M{"imdb_id": id})
			if err != nil {
				errorChan <- err // Send error to channel
				return
			}

			defer cursor.Close(ctx)

			if err = cursor.All(ctx, &movies); err != nil {
				errorChan <- err // Send error to channel
				return
			}

			moviesChan <- movies // Send result to channel
		}()
		select {
		case movies := <-moviesChan:
			// Success case - got movies from channel
			c.JSON(200, movies)
		case err := <-errorChan:
			// Error case - got error from channel
			c.JSON(500, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			// Timeout case - neither channel responded in time
			c.JSON(408, gin.H{"error": "request timeout"})
		}
	}
}

func GetTopRatedMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		moviesChan := make(chan []models.Movie, 1)
		errorChan := make(chan error, 1)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Get top rated movies (rating >= 7, sorted by rating desc)
			filter := bson.M{"ranking.ranking_value": bson.M{"$gte": 7}}
			opts := options.Find().SetSort(bson.D{{Key: "ranking.ranking_value", Value: -1}}).SetLimit(20)

			cursor, err := movieCollection.Find(ctx, filter, opts)
			if err != nil {
				errorChan <- err
				return
			}
			defer cursor.Close(ctx)

			var movies []models.Movie
			if err = cursor.All(ctx, &movies); err != nil {
				errorChan <- err
				return
			}

			moviesChan <- movies
		}()

		select {
		case movies := <-moviesChan:
			c.JSON(200, gin.H{
				"top_rated_movies": movies,
				"total_found":      len(movies),
				"minimum_rating":   7,
			})
		case err := <-errorChan:
			c.JSON(500, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			c.JSON(408, gin.H{"error": "request timeout"})
		}
	}
}

func GetMoviesByGenre() gin.HandlerFunc {
	return func(c *gin.Context) {
		genreName := c.Param("genre")
		if genreName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Genre parameter required"})
			return
		}

		moviesChan := make(chan []models.Movie, 1)
		errorChan := make(chan error, 1)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Find movies with the specified genre
			filter := bson.M{"genre.genre_name": bson.M{"$regex": genreName, "$options": "i"}}
			opts := options.Find().SetLimit(20)

			cursor, err := movieCollection.Find(ctx, filter, opts)
			if err != nil {
				errorChan <- err
				return
			}
			defer cursor.Close(ctx)

			var movies []models.Movie
			if err = cursor.All(ctx, &movies); err != nil {
				errorChan <- err
				return
			}

			moviesChan <- movies
		}()

		select {
		case movies := <-moviesChan:
			c.JSON(200, gin.H{
				"movies":      movies,
				"genre":       genreName,
				"total_found": len(movies),
			})
		case err := <-errorChan:
			c.JSON(500, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			c.JSON(408, gin.H{"error": "request timeout"})
		}
	}
}

func MakeMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		var movie models.Movie
		movieMade := make(chan bool, 1) // Add buffer to prevent blocking
		errorChan := make(chan error, 1)

		// Parse JSON body into movie struct
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON"})
			return
		}

		if err := movieValidate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			_, err := movieCollection.InsertOne(ctx, movie)
			if err != nil {
				errorChan <- err // Send error to channel
				return
			}

			movieMade <- true // Send success signal
		}()

		// Use select to handle channels properly (like your other functions)
		select {
		case <-movieMade:
			// Success case
			c.JSON(201, gin.H{"message": "Movie created successfully"}) // 201 = Created
		case err := <-errorChan:
			// Error case
			c.JSON(500, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			// Timeout case
			c.JSON(408, gin.H{"error": "request timeout"})
		}
	}
}

func AdminReviewUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		movieId := c.Param("imdb_id")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID required"})
			return
		}

		var req struct {
			AdminReview string `json:"admin_review" validate:"required"`
			Rating      int    `json:"rating" validate:"required,min=1,max=10"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		if err := movieValidate.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		// Simple rating mapping
		rankingName := getRatingName(req.Rating)

		filter := bson.M{"imdb_id": movieId}
		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_value": req.Rating,
					"ranking_name":  rankingName,
				},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := movieCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update movie"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Review updated successfully",
			"admin_review": req.AdminReview,
			"rating":       req.Rating,
			"ranking_name": rankingName,
		})
	}
}

// Simple rating name mapping - no AI needed!
func getRatingName(rating int) string {
	switch {
	case rating >= 9:
		return "excellent"
	case rating >= 7:
		return "good"
	case rating >= 5:
		return "average"
	case rating >= 3:
		return "poor"
	default:
		return "terrible"
	}
}

// Get user by ID helper function
func getUserById(userID string) (*models.User, error) {
	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Access userCollection from user_controller package
	userColl := database.OpenCollection("User")
	err := userColl.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Real recommendation system based on user preferences
func GetRecommendedMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
			return
		}

		// Get user's favorite genres
		user, err := getUserById(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Extract genre names from user's favorites
		var genreNames []string
		for _, genre := range user.FavouriteGenres {
			genreNames = append(genreNames, genre.GenreName)
		}

		if len(genreNames) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message":     "No favorite genres set for user",
				"movies":      []models.Movie{},
				"user_genres": []string{},
			})
			return
		}

		// Find movies in user's favorite genres with good ratings
		moviesChan := make(chan []models.Movie, 1)
		errorChan := make(chan error, 1)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Filter movies by user's favorite genres and good ratings
			filter := bson.M{
				"genre.genre_name":      bson.M{"$in": genreNames},
				"ranking.ranking_value": bson.M{"$gte": 6}, // Only recommend good movies
			}

			// Sort by rating (highest first) and limit results
			opts := options.Find().SetSort(bson.D{{Key: "ranking.ranking_value", Value: -1}}).SetLimit(10)
			cursor, err := movieCollection.Find(ctx, filter, opts)
			if err != nil {
				errorChan <- err
				return
			}
			defer cursor.Close(ctx)

			var movies []models.Movie
			if err = cursor.All(ctx, &movies); err != nil {
				errorChan <- err
				return
			}

			moviesChan <- movies
		}()

		select {
		case movies := <-moviesChan:
			c.JSON(200, gin.H{
				"recommended_movies": movies,
				"based_on_genres":    genreNames,
				"total_found":        len(movies),
			})
		case err := <-errorChan:
			c.JSON(500, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			c.JSON(408, gin.H{"error": "request timeout"})
		}
	}
}
