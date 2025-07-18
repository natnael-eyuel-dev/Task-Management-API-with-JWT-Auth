package middleware

// imports
import (
	"net/http";                          
	"github.com/dgrijalva/jwt-go";        
	"github.com/gin-gonic/gin";          
)

// temporary secret
var jwtSecret = []byte("jwt-auth-secret")

func ValidateToken(token string) (*jwt.Token, error){
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		
		_, ok := token.Method.(*jwt.SigningMethodHMAC)    // check if token uses HMAC signing  
		if !ok {
			return nil, jwt.ErrSignatureInvalid      // block invalid signing 
		}
		return jwtSecret, nil     // return secret to verify signature
	})
}

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenStr := c.GetHeader("Authorization")     // get token from authorization header
		// reject if empty
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}
		
		// validate token structure/signature with error handling 
		token, err := ValidateToken(tokenStr)     
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// if token is valid, extract claims and store in request context
		claims, ok := token.Claims.(jwt.MapClaims)      
		if ok {
			c.Set("userID", claims["sub"])             // user id
			c.Set("username", claims["username"])      // username 
			c.Set("role", claims["role"])              // user role (admin/user)
		}

		c.Next()     // proceed to next handler
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		
		role, exists := c.Get("role")          // get role from context 

		// block if either role doesn't exist in context or role isn't "admin"
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			
			c.Abort()
			return
		}

		c.Next()     // allow admin to proceed
	}
}