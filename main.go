package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kibuns/AuthService/DAL"
	"github.com/Kibuns/AuthService/Models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var SECRET = []byte("super-secret-auth-key")
const api_key_required = false;
var api_key = "1234"

func main() {
    // create a channel to receive signals to stop the application
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    // start the goroutine to receive messages from the queue
    go receive()

    // start the goroutine to handle API requests
    go handleRequests()

    // wait for a signal to stop the application
    <-stop
}

func returnUser(w http.ResponseWriter, r *http.Request) {
	var usernameParam string = mux.Vars(r)["username"]
	result, err := DAL.SearchUser(usernameParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func CreateJWT(username string) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["username"] = username

	tokenStr, err := token.SignedString(SECRET)

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return tokenStr, nil
}

func ValidateJWT(next func(w http.ResponseWriter, r* http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("validating jwt")
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(t *jwt.Token) (interface{}, error) {
				_, ok := t.Method.(*jwt.SigningMethodHMAC)
				if !ok {
					http.Error(w, "invalid token", http.StatusUnauthorized)
				}
				return SECRET, nil
			})

			if err != nil {
				http.Error(w, "error while validating token", http.StatusBadRequest)
			}

			if token.Valid {
				next(w, r)
			}
		} else {
			http.Error(w, "token header not present", http.StatusBadRequest)
		}
	})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var idParam string = mux.Vars(r)["user"]
	DAL.DeleteAllOfUser(idParam)
	fmt.Fprintf(w, "deleted everything from user: " + idParam)
}

func GetJwt(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	fmt.Println("Checking credentials")
	// parse the request body into a User struct
	var user Models.User
	err := json.NewDecoder(body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	//find the user credentials in DB and check
	userCred, err := DAL.SearchUser(user.UserName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var encryptedPassword = userCred["password"].(string)

	var hashedUserInput = sha256.Sum256([]byte(user.Password))
	var hashedString = hex.EncodeToString(hashedUserInput[:])

	if err != nil || encryptedPassword != hashedString {
		http.Error(w, "incorrect credentials", http.StatusUnauthorized)
		fmt.Println(encryptedPassword + " does not equal " + hashedString)
		// fmt.Println(err)
		return
	}

	fmt.Println("Credentials are correct!")

	if api_key_required{
		if r.Header["Access"] != nil {
			if r.Header["Access"][0] != api_key {
				fmt.Println("Incorrect api key")
				http.Error(w, "incorrect API key", http.StatusUnauthorized)
				return
			} else {
				token, err := CreateJWT(user.UserName)
				if err != nil {
					return
				}
				fmt.Fprint(w, token)
			}
		} else {
			fmt.Println("No api key found")
			http.Error(w, "no API key found", http.StatusUnauthorized)
		}
	} else{
		token, err := CreateJWT(user.UserName)
		if err != nil {
			return
		}
		fmt.Fprint(w, token)
	}
	
	
}

func getUsernameFromTokenHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getting username")
	// Get the token from the "Token" header
	tokenStr := r.Header.Get("Token")
	if tokenStr == "" {
		http.Error(w, "Token header missing", http.StatusBadRequest)
		return
	}

	// Call the function to retrieve the "username" claim from the token
	username, err := getUsernameFromToken(tokenStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Println(username)

	// Return the "username" claim in the response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(username))
}


func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "super secret area")
}



func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", Home)
	myRouter.Handle("/api", ValidateJWT(Home))
	myRouter.HandleFunc("/jwt", GetJwt)
	myRouter.HandleFunc("/delete/{user}", deleteUser)
	myRouter.HandleFunc("/get/{username}", returnUser)
	myRouter.Handle("/getusername", ValidateJWT(getUsernameFromTokenHandler))

	log.Fatal(http.ListenAndServe(":8083", myRouter))
}

func validateToken(tokenStr string) bool {
	// Parse the token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		// Return the secret used for signing
		return []byte(SECRET), nil
	})

	if err != nil || !token.Valid {
		// Token validation failed
		return false
	}

	return true
}

func getUsernameFromToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(SECRET), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if username, ok := claims["username"].(string); ok {
			return username, nil
		}
	}

	return "", fmt.Errorf("username claim not found in token")
}