package routes

import (
	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/controllers"
	"github.com/Futuredakster/GoProject/Server/MagicStreamMoviesServer/middleware"
	"github.com/gin-gonic/gin"
)

func MovieRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "Movie Streaming API",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /movies - Get all movies",
				"GET /movies/top-rated - Get highest rated movies",
				"GET /movies/genre/:genre - Get movies by genre",
				"GET /movie/:imdb_id - Get specific movie",
				"GET /movies/recommended/:user_id - Get personalized recommendations",
				"POST /movies - Create new movie (auth required)",
				"PUT /movies/:imdb_id/review - Add admin review (auth required)",
			},
		})
	})

	// Public routes (no authentication needed)
	router.GET("/movies", controllers.GetMovies())
	router.GET("/movies/top-rated", controllers.GetTopRatedMovies())
	router.GET("/movies/genre/:genre", controllers.GetMoviesByGenre())
	router.GET("/movie/:imdb_id", controllers.GetMovie())
	router.GET("/movies/recommended/:user_id", controllers.GetRecommendedMovies())

	// Protected route group
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare())
	{
		protected.POST("/movies", controllers.MakeMovies())
		protected.PUT("/movies/:imdb_id/review", controllers.AdminReviewUpdate())
		// Add more protected routes here as needed
		// protected.PUT("/movies/:id", controllers.UpdateMovie())
		// protected.DELETE("/movies/:id", controllers.DeleteMovie())
	}
}
