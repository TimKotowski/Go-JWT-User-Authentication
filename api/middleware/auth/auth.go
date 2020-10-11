package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"Go-JWT-Auth/api/cryption"
	"Go-JWT-Auth/api/jwtgenerate"
	"Go-JWT-Auth/models"

	"github.com/go-chi/chi"
	jsoniter "github.com/json-iterator/go"
	"github.com/volatiletech/sqlboiler/queries"
	"golang.org/x/crypto/bcrypt"
)

var (
	db   *sql.DB
	json = jsoniter.ConfigFastest
	ctx  = context.Background()
)

func New(db *sql.DB, r *chi.Mux) {
	r.Post("/api/v1/signup", userSignUp(db))
	r.Post("/api/v1/login", userLogin(db))
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
		// err := user.Insert(ctx, db, boiler.Infer()) //TODO
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
		jwtToken, err := jwtgenerate.GenerateJWT()
		if passErr != nil {
			log.Fatalf("error in JWT token generation %v", err)
		}
		w.Write([]byte(`{"token": "` + jwtToken + `"}`))
	}
}
