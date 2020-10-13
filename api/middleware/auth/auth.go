package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"Go-JWT-Auth/api/cryption"
	"Go-JWT-Auth/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	jsoniter "github.com/json-iterator/go"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"golang.org/x/crypto/bcrypt"
)

var (
	db     *sql.DB
	json   = jsoniter.ConfigFastest
	ctx    = context.Background()
	jwtKey = []byte("my_secret_key")
)

func New(db *sql.DB, r *chi.Mux) {
	r.Post("/api/v1/signup", userSignUp(db))
	r.Post("/api/v1/login", userLogin(db))
	r.Post("/api/v1/welcome", userWelcome(db))
}

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Email string `boil:"email" json:"email" toml:"email" yaml:"email"`
	jwt.StandardClaims
}

//Above userSignup function is doing the user registration part. Here response body’s json
// data coming from client browser is decoded in a User struct.
//  Password is hashed before storing in Database with
//  below code snippet. For password hasing I have used bcrypt.
func userSignUp(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user := &models.User{}
		json.NewDecoder(r.Body).Decode(user)
		user.Password = cryption.GetHash([]byte(user.Password))
		err := queries.Raw(`INSERT INTO users (firstname, lastname, email, password) VALUES ($1, $2, $3, $4) RETURNING id`,
			user.Firstname, user.Lastname, user.Email, user.Password).Bind(ctx, db, user)
		// err := user.Insert(ctx, db, boil.Infer())
		if err != nil {
			log.Fatalf("\n err in q %v", err)
			return
		}
		userData, err := json.Marshal(user)
		if err != nil {
			log.Fatalf("\n err is now %v", err)
			return
		}
		w.Write(userData)
	}
}

func userLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// two structs one for the user
		// the other one for password to compare
		user := &models.User{}
		dbUser := &models.User{}
		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			log.Fatalf("\n err in 1  %v", err)
			return
		}
		// foundUser, err := models.Users(qm.Where("email", user.Email)).One(ctx, db)
		err := queries.Raw(`SELECT id, firstname, lastname, email, password FROM users WHERE email= $1`,
			user.Email).Bind(ctx, db, dbUser)

		if err != nil {
			log.Fatalf("user not found in databse %v", err)
		}
		fmt.Println("\n user ", user)
		fmt.Println("\n user ", dbUser)

		userPass := []byte(user.Password)
		dbPass := []byte(dbUser.Password)

		// CompareHashAndPassword compares a bcrypt hashed password with its possible
		passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

		if passErr != nil {
			log.Fatalf("error in comapring passwords %v", passErr)
		}
		// Declare the expiration time of the token
		// here, we have kept it as 5 minutes
		expirationTime := time.Now().Add(5 * time.Minute)
		// Create the JWT claims, which includes the username and expiry time
		claims := &Claims{
			Email: dbUser.Email,
			StandardClaims: jwt.StandardClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: expirationTime.Unix(),
			},
		}

		// Declare the token with the algorithm used for signing, and the claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// Create the JWT string
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Finally, we set the client cookie for "token" as the JWT we just generated
		// we also set an expiry time which is the same as the token itself
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})

	}
}

func userWelcome(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				log.Fatalf("err no cookie found %v", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			log.Fatalf("err bad requesr no cookie found %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Get the JWT string from the cookie
		tknStr := c.Value

		// Initialize a new instance of `Claims`
		claims := &Claims{}

		// Parse the JWT string and store the result in `claims`.
		// Note that we are passing the key in this method as well. This method will return an error
		// if the token is invalid (if it has expired according to the expiry time we set on sign in),
		// or if the signature does not match
		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				log.Fatalf("err invalid signature %v", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			log.Fatalf("badrequest %v", err)
			return
		}
		if !tkn.Valid {
			log.Fatalf("no token or invalid token %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Finally, return the welcome message to the user, along with their
		// username given in the token
		w.Write([]byte(fmt.Sprintf("welcome %s", claims.Email)))
	}
}
