package controllers

// imports
import (
	"net/http";
	"github.com/gin-gonic/gin";
	"github.com/natnael-eyuel-dev/Task-Management-API-with-JWT-Auth/data";
	"github.com/natnael-eyuel-dev/Task-Management-API-with-JWT-Auth/models";
)

type UserController struct {
	userService data.UserService       // service layer for user operations
}

func NewUserController(service data.UserService) *UserController {
	return &UserController{userService: service}         // return new controller instance 
}

func (userContr *UserController) Register(c *gin.Context) {
	
	var user models.User
	err := c.ShouldBindJSON(&user)    // parse request body into user struct
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})     // returns parsing errors
		return
	}

	// create user through service layer
	err = userContr.userService.Register(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message":"user created successfully"})      // success response
}

func (userContr *UserController) Login(c *gin.Context) {

	var credentials models.Credentials

	err := c.ShouldBindJSON(&credentials)       // parse request body into credentials struct
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// authenticate user through service layer
	token, user, err := userContr.userService.Login(&credentials)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error":err.Error()})     
		return
	}

	// return token, user info (excluding sensitive data)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id": 		 user.ID,
			"username":  user.Username,
			"role":      user.Role,
		},
	})
}

func (userContr *UserController) PromoteAdmin(c *gin.Context) {
    
    userID := c.Param("id")       // get user id from request parameter
    
    // promote user through service layer
    err := userContr.userService.PromoteUserToAdmin(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "user promoted to admin successfully"})
}
