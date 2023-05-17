package main

import (
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

func CreateJWT() (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["username"] = "test"

	tokenStr, err := token.SignedString(SECRET)

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return tokenStr, nil
}

func ValidateJWT(next func(w http.ResponseWriter, r* http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(t *jwt.Token) (interface{}, error) {
				_, ok := t.Method.(*jwt.SigningMethodHMAC)
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("not authorized"))
				}
				return SECRET, nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("not authorized: " + err.Error()))
			}

			if token.Valid {
				next(w, r)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("not authorized"))
		}
	})
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

	//find the user credentials in DB
	userCred, err := DAL.SearchUser(user.UserName)
	if err != nil || userCred["password"].(string) != user.Password {
		http.Error(w, "Incorrect credentials", http.StatusUnauthorized)
		fmt.Println("******ERROR******")
		// fmt.Println(err)
		return
	}

	// //TEMP CODE TEMP CODE TEMP CODE TEMP CODE TEMP
	// if user.UserName != "ninoverhaegh" || user.Password != "1234" {
	// 	http.Error(w, "credentials are incorrect", http.StatusUnauthorized)
	// 	fmt.Println(err)
	// 	return
	// }
	// //TEMP CODE TEMP CODE TEMP CODE TEMP CODE TEMP

	fmt.Println("Credentials are correct!")

	if r.Header["Access"] != nil {
		if r.Header["Access"][0] != api_key {
			return
		} else {
			token, err := CreateJWT()
			if err != nil {
				return
			}
			fmt.Fprint(w, token)
		}
	}
}


func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "super secret area")
}



func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", Home)
	myRouter.Handle("/api", ValidateJWT(Home))
	myRouter.HandleFunc("/jwt", GetJwt)
	myRouter.HandleFunc("/get/{username}", returnUser)

	log.Fatal(http.ListenAndServe(":3500", myRouter))
}