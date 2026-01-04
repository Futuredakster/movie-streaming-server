package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Example of how a movie document will look in MongoDB:
// {
//   "_id": ObjectId("..."),
//   "imdb_id": "tt1234567",
//   "title": "Avengers",
//   "poster_path": "/poster.jpg",
//   "youtube_id": "abc123",
//   "genre": [
//     {"genreID": 1, "genreName": "Action"},
//     {"genreID": 2, "genreName": "Adventure"}
//   ],
//   "admin_review": "Great movie",
//   "ranking": {
//     "rankingValue": 5,
//     "rankingName": "PG-13"
//   }
// }

// BSON vs JSON tags:
// `bson:"_id"` = when saving to MongoDB, call this field "_id"
// `json:"_id"` = when converting to JSON for API, call this field "_id"
// BSON = Binary JSON (MongoDB's internal storage format)
// In Node.js, the driver handles BSON conversion automatically
// In Go, we need explicit tags to handle the conversion

type Genre struct {
	GenreID   int    `bson:"genre_id" json:"genre_id" validate:"required"`
	GenreName string `bson:"genre_name" json:"genre_name" validate:"required,min=2,max=100"`
}

type Ranking struct {
	RankingValue int    `bson:"ranking_value" json:"ranking_value" validate:"required"`
	RankingName  string `bson:"ranking_name" json:"ranking_name" validate:"required"`
}

type Movie struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ImdbID      string        `bson:"imdb_id" json:"imdb_id" validate:"required"`
	Title       string        `bson:"title" json:"title" validate:"required,min=2,max=500"`
	PosterPath  string        `bson:"poster_path" json:"poster_path" validate:"required,url"`
	YouTubeID   string        `bson:"youtube_id" json:"youtube_id" validate:"required"`
	Genre       []Genre       `bson:"genre" json:"genre" validate:"required,dive"`
	AdminReview *string       `bson:"admin_review,omitempty" json:"admin_review,omitempty"`
	Ranking     *Ranking      `bson:"ranking,omitempty" json:"ranking,omitempty"`
}
