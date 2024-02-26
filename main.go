package main

import (
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
	"github.com/gin-contrib/cors"
    "math/rand"
    "net/http"
	"fmt"
	"strconv"
	"strings"
)

var userPoints map[string]int 
var redisClient *redis.Client
var sharedUsername string
var prev int

func init() {
    userPoints = make(map[string]int)


	redisClient = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", 
        Password: "",                
        DB:       0,                 
    })
}


func registerHandler(c *gin.Context) {
    
    var registrationRequest struct {
        Username string `json:"username"`
    }

    if err := c.ShouldBindJSON(&registrationRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    if registrationRequest.Username == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username cannot be empty"})
        return
    }

    _, err := redisClient.Get(redisClient.Context(), registrationRequest.Username).Result()
    if err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
        return
    } else if err != redis.Nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username existence"})
        return
    }

    if err := redisClient.Set(redisClient.Context(), registrationRequest.Username, 0, 0).Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func startGameHandler(c *gin.Context) {
    username := c.Query("username") 

    userPoints[username] = 0

    cardTypes := []string{"Cat", "Defuse", "Shuffle", "Exploding_Kitten"} // 
    drawnCard := cardTypes[rand.Intn(len(cardTypes))]

    switch drawnCard {
    case "Cat":

    }

    
    if err := updateRedisUserPoints(username); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update user points in Redis: %s", err)})
        return
    }

    c.JSON(http.StatusOK, gin.H{"card": drawnCard, "userPoints": userPoints[username]})
}


func drawCardHandler(c *gin.Context) {
    username := c.Query("username") 

    cardTypes := []string{"Cat", "Defuse", "Shuffle", "Exploding_Kitten"}
    drawnCard := cardTypes[rand.Intn(len(cardTypes))]
	
    switch drawnCard {
    case "Cat":
    case "Defuse":
    case "Shuffle":
		startGameHandler(c)
        return
    case "Exploding_Kitten":

    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unexpected card type: %s", drawnCard)})
        return
    }
	
	if err := updateRedisUserPoints(username); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update user points in Redis: %s", err)})
        return
    }

    c.JSON(http.StatusOK, gin.H{"card": drawnCard, "userPoints": userPoints[username]})
}


func leaderboardHandler(c *gin.Context) {
    usernames, err := redisClient.Keys(redisClient.Context(), "*").Result()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve user points from Redis: %s", err)})
        return
    }

    leaderboard := make(map[string]interface{})
    for _, username := range usernames {
        if strings.HasPrefix(username, "user_") {
            continue
        }

        userPointStr, err := redisClient.Get(redisClient.Context(), username).Result()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve user point for %s from Redis: %s", username, err)})
            return
        }

        fmt.Printf("DEBUG - User: %s, User Point (string): %s\n", username, userPointStr)

        leaderboard[username] = userPointStr
    }

    c.JSON(http.StatusOK, leaderboard)
}

func updateRedisUserPoints(username string) error {
    normalizedUsername := strings.Trim(username, `"`)
    userPointsStr := strconv.Itoa(prev)
    if _, err := strconv.Atoi(userPointsStr); err != nil {
        return fmt.Errorf("Invalid user point value for %s: %s", normalizedUsername, userPointsStr)
    }

    if err := redisClient.Set(redisClient.Context(), normalizedUsername, userPointsStr, 0).Err(); err != nil {
        return fmt.Errorf("Failed to update user points in Redis: %s", err)
    }

    return nil
}

func victoryHandler(c *gin.Context) {

    

    var victoryRequest struct {
        Username string `json:"username"`
        
    }
    
    if err := c.ShouldBindJSON(&victoryRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }
       
    
    fmt.Println("User Point is: ", userPoints[victoryRequest.Username])
        userPoints[victoryRequest.Username]++

        prev = userPoints[victoryRequest.Username]+prev

        if err := updateRedisUserPoints(victoryRequest.Username); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update user score: %s", err)})
        return
    }

    c.JSON(http.StatusOK, gin.H{"username": victoryRequest.Username, "message": "Victory recorded successfully"})
    
}


func getRedisUserPoints(username string) (int, error) {
    normalizedUsername := strings.Trim(username, `"`)

    userPointsStr, err := redisClient.Get(redisClient.Context(), normalizedUsername).Result()
    if err != nil {
        return 0, err
    }

    userPoints, err := strconv.Atoi(userPointsStr)
    if err != nil {
        return 0, err
    }

    return userPoints, nil
}


func victoriousPlayerHandler(c *gin.Context) {

    if sharedUsername == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username not found"})
        return
    }

    fmt.Println("Shared Username is: ",sharedUsername)
    c.JSON(http.StatusOK, gin.H{"username": sharedUsername})
    }

func main() {

    r := gin.Default()
    r.Use(cors.Default())

    r.POST("/register", registerHandler)

    r.POST("/start-game", startGameHandler)

    r.GET("/draw-card", drawCardHandler)

    r.GET("/leaderboard", leaderboardHandler)

     r.POST("/victory", victoryHandler)

    r.GET("/victorious-player", victoriousPlayerHandler)

    r.Run(":8080")

}
