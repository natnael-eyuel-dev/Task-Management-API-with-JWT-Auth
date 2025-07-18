package models

type User struct {
	ID           string      `bson:"_id,omitempty" json:"id"`         // mongodb's unique identifier for users 
	Username     string      `bson:"username" json:"username"`        // username 
	Password     string      `bson:"password" json:"password"`        // password (hashed before storage)
	Role         string      `bson:"role" json:"role"`                // user role (role/user)
}

type Credentials struct {
	Username 	 string      `json:"username" binding:"required"`     // login username (required field)
    Password 	 string 	 `json:"password" binding:"required"`     // login password (required field)
}