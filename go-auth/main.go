package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
  ID       uuid.UUID   `json:"id" bson:"_id"`
  Email    string `json:"email" bson:"email"`
  Password string `json:"password" bson:"password"`
}

var jwtSecret = []byte("your_jwt_secret")

func main() {
  // Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("test").Collection("trainers")

  indexModel := mongo.IndexModel{
    Keys:    bson.D{{Key: "email", Value: 1}}, // Create index on the "email" field
    Options: options.Index().SetUnique(true),    // Ensure the index is unique
}

// Create the index
_, err = collection.Indexes().CreateOne(context.TODO(), indexModel)
if err != nil {
    log.Printf("could not create index: %v", err)
}

  router := gin.Default()

  router.GET("/", func(c *gin.Context) {
    c.IndentedJSON(200, gin.H{
      "message": "Welcome to the Go Authentication and Authorization tutorial!",
    })
  })

  // Get all users from the mongoDB connection  
  router.GET("/users", func(c *gin.Context) {
    var users []User 
    cursor, err := collection.Find(context.Background(), bson.M{})
    if err != nil {
      c.IndentedJSON(500, gin.H{"error": "Internal server error"})
      return
    }
    defer cursor.Close(context.Background())
    for cursor.Next(context.Background()) {
      var user User
      cursor.Decode(&user)
      users = append(users, user)
    }
    c.IndentedJSON(200, users)
  }) 

  router.POST("/register", func(c *gin.Context) {
    
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
      c.IndentedJSON(400, gin.H{"error": "Invalid request payload"})
      return
    }
  
    // User registration logic
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
      c.IndentedJSON(500, gin.H{"error": "Internal server error"})
      return
    }

    user.ID = uuid.New()
    user.Password = string(hashedPassword)
    // Insert user into the database
    _, err = collection.InsertOne(context.Background(), user)
    if err != nil {
      if mongo.IsDuplicateKeyError(err) {
        // return error message with status code 400 if user already exists
        c.IndentedJSON(400, gin.H{"error": "User already exists"})
        return
      }
      c.IndentedJSON(500, gin.H{"error": "Internal server error"})
      return
    }
    
    // return success message with status code 200 and user details
    c.IndentedJSON(200, user)
  })

  
router.POST("/login", func(c *gin.Context) {
  var user User
  if err := c.ShouldBindJSON(&user); err != nil {
    c.IndentedJSON(400, gin.H{"error": "Invalid request payload"})
    return
  }

  // TODO: Implement user login logic
  var result User
  err := collection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&result)
  if err != nil {
    c.IndentedJSON(400, gin.H{"error": "User not found"})
    return
  }
  // check for password match
  err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
  if err != nil {
    c.IndentedJSON(400, gin.H{"error": "Invalid credentials"})
    return
  }

  // Create a new token
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": result.ID,
    "email": result.Email,
  })

  jwtToken, err := token.SignedString(jwtSecret)
  if err != nil {
    c.JSON(500, gin.H{"error": "Internal server error"})
    return
  }

    c.JSON(200, gin.H{"message": "User logged in successfully", "token": jwtToken})
  })

  router.GET("/secure", AuthMiddleware(), func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "This is a secure route"})
  })

  router.Run()
}


func AuthMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    // JWT validation logic
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
      c.JSON(401, gin.H{"error": "Authorization header is required"})
      c.Abort()
      return
    }

    authParts := strings.Split(authHeader, " ")
    if len(authParts) != 2 || strings.ToLower(authParts[0]) != "bearer" {
      c.JSON(401, gin.H{"error": "Invalid authorization header"})
      c.Abort()
      return
    }

    token, err := jwt.Parse(authParts[1], func(token *jwt.Token) (interface{}, error) {
      if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
      }

      return jwtSecret, nil
    })

    if err != nil || !token.Valid {
      c.JSON(401, gin.H{"error": "Invalid JWT"})
      c.Abort()
      return
    }
    

    c.Next()
  }
}