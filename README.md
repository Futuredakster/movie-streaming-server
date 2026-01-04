# Movie Streaming Server (Go)

A RESTful API server built with Go and Gin framework for a movie streaming application.

## Features

- 🔐 JWT Authentication & Authorization
- 🎬 Movie CRUD operations
- ⭐ Rating and review system
- 👥 User management
- 🔍 Movie search and recommendations
- 🌐 CORS enabled for frontend integration

## Tech Stack

- **Framework**: Gin (Go web framework)
- **Database**: MongoDB
- **Authentication**: JWT tokens
- **AI Integration**: OpenAI for content processing

## Quick Start

1. **Clone the repository**
2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up environment variables**:
   - Copy `.env.example` to `.env`
   - Update the values with your configurations

4. **Run the server**:
   ```bash
   go run main.go
   ```

## Environment Variables

See `.env.example` for all required environment variables.

## API Endpoints

- `GET /health` - Health check
- `POST /register` - User registration
- `POST /login` - User login
- `GET /movies` - Get all movies
- `GET /movie/:imdb_id` - Get movie by ID
- `POST /movies` - Create movie (Admin)
- `PUT /movie/:imdb_id/admin-review` - Add admin review

## Deployment

This server is ready for deployment on:
- Render
- Railway  
- Heroku
- Digital Ocean
- AWS

Set the `PORT` environment variable and ensure `GIN_MODE=release` for production.