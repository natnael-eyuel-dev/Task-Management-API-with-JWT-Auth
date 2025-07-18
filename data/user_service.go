package data

// imports
import (
	"context";
	"errors";
	"fmt";
	"log";
	"time";
	"github.com/dgrijalva/jwt-go";
	"github.com/natnael-eyuel-dev/Task-Management-API-with-JWT-Auth/models";
	"go.mongodb.org/mongo-driver/bson";
	"go.mongodb.org/mongo-driver/bson/primitive";
	"go.mongodb.org/mongo-driver/mongo";
	"golang.org/x/crypto/bcrypt";
)

type UserService struct {
	db  *MongoDBTaskManager     // reuses existing database connection
}

// creates new UserService instance
func NewUserService(db *MongoDBTaskManager)  *UserService {
	return &UserService{db: db}
}

// helper to access users collection
func (userServ *MongoDBTaskManager) UserCollection() *mongo.Collection {
	return userServ.client.Database(userServ.database).Collection(userServ.collection)
}

func (userServ *UserService) Register(user *models.User) error {
    
	collection := userServ.db.UserCollection()      // get user collection 
	
	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)      // set timeout
	defer cancel()

	// validate input
	if user.Username == "" {
		return errors.New("username can not be empty")
	}	
	if user.Password == "" {
		return errors.New("password can not be empty")
	} else if len(user.Password) < 8 {
		return errors.New("password must be 8+ characters")     // vaid password length is 8+ characters
	}

	// check if user already exists
	exist := collection.FindOne(contx, bson.M{
		"username": user.Username,
	})

	if exist.Err() == nil {
		return errors.New("username already exists")
	}

	// set first user role to admin if user collection is empty
	count, err := collection.CountDocuments(contx, bson.D{})
	if err != nil {
		log.Fatalf("failed to check user count : %v", err)
		return errors.New("internal server error")
	}
	
	user.Role = "user"      // default role
	if count == 0 {
		user.Role = "admin"     // first user becomes admin 
	} 

	// hash password securely 
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password : %v", err)
		return errors.New("internal server error")
	}

	user.Password = string(hashed)   // set user password to hashed password

	// save user to database
	_, err = collection.InsertOne(contx, user)
	if err != nil {
		log.Fatalf("failed to create user : %v", err)
		return errors.New("internal server error")
	}

	return nil     // success 
}

// authenticate user
func (userServ *UserService) Login(credentials *models.Credentials) (string, *models.User, error) {
	
	var user models.User
	collection := userServ.db.UserCollection()      // get user collection 
	
	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)      // set timeout
	defer cancel()

	// find user by username
	err := collection.FindOne(contx, bson.M{"username": credentials.Username,}).Decode(&user)
	if err != nil {
        if err == mongo.ErrNoDocuments {
            return "", nil, errors.New("user not found")
        }
        return "", nil, fmt.Errorf("database error: %v", err)
    }

	// verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// generate jwt token
	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
        return "", nil, fmt.Errorf("failed to generate token: %v", err)
    }

	return token, &user, nil       // success
}

// promote a user to admin role (only admin can do this)
func (userServ *UserService) PromoteUserToAdmin(userID string) error {
    
	collection := userServ.db.UserCollection()      // get user collection 
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)      // set timeout
    defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

    // update user's role to admin
    result, err := collection.UpdateOne(
        ctx,
        bson.M{"_id": objID},
        bson.M{"$set": bson.M{"role": "admin"}},
    )
    
    if err != nil {
        log.Printf("Error promoting user: %v", err)
        return errors.New("failed to promote user")
    }
    
    if result.MatchedCount == 0 {
        return errors.New("user not found")
    }
    
    return nil     // success
}

// temporary secret
var jwtSecret = []byte("jwt-auth-secret")

func GenerateToken(userID, username, role string) (string, error){
	// create token with claims 
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,            // user id          
		"username": username,        // username
		"role": role,                // user role (admin/user)
		"exp": time.Now().Add(time.Hour * 24).Unix(),      // expires in 24h
	})

	// sign with secret key
	return token.SignedString(jwtSecret)
}
