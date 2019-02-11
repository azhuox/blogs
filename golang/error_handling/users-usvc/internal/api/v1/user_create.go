package v1

import (
	"net/http"
	"fmt"
	"log"
	"encoding/json"

	"github.com/gorilla/mux"

	userV1 "github.com/aaronzhuo1990/blogs/golang/error_handling/users-usvc/internal/user/v1"
)

// CreateUserAPIHandler is the API handler for creating a site. It uses builtin errors to do the error handling
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    var user := &struct{
        Firstname string   `json:"firstname"`
        Lastname  string   `json:"lastname"`
        Pasword string `json:"phone"`
        Email   string `json:"email"`
    }

    // Parse args
    params := mux.Vars(r)
    if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, fmt.Sprintf("Error decoding request params, err: %s", err.Error()), http.StatusBadRequest)
        return
    }

    // Create a user manager
    userManager, err := userV1.NewManager(...)
    if err != nil {
        log.Printf("[user_create_v1] error creating user manager, err: %s", err.Error())
        http.Error(w, fmt.Sprintf("Internal server error, please retry later"), http.StatusBadRequest)
        return
    }

    // Use the user manager to create a user with given parameters
    ID, err := userManager.Create(user.FirstName, user.LastName, user.Pasword, user.Email)
	if err != nil {
        log.Printf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())
        http.Error(w, fmt.Sprintf("Internal server error, please retry later"), http.StatusInternalServerError)
        return
	   }
	// Return ID
    json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
    w.WriteHeader(http.StatusOK)
    return
}


// CreateUserAPIHandler is the API handler for creating a site. It uses the first solution to do the error handling.
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	... // A bunch of operations are omitted

   // Use the user manager to create a user with given parameters
   ID, err := userManager.Create(user.FirstName, user.LastName, user.Pasword, user.Email)
   if err != nil {
		log.Printf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())
		if _, ok := err.(*userV1.BadRequestErr); ok {
			http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusBadRequest)
		} else if _, ok := err.(*userV1.EmailHasBeenUsedErr); ok {
			http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusConflict)
		} else if _, ok := err.(*userV1.InternelServerErr); ok {
			http.Error(w, fmt.Sprintf("Internal server error, please retry later."), http.StatusInternalServerError)
		} else {
			// This should never happen
           http.Error(w, fmt.Sprintf("Unknown error, please retry later."), http.StatusInternalServerError)
       }
       return
   }
   // Return ID
   json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
   w.WriteHeader(http.StatusOK)
   return
}

// CreateUserAPIHandler is the API handler for creating a site. It uses the second solution to do the error handling.
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	... // A bunch of operations are omitted

    // Use the user manager to create a user with given parameters
	ID, err := userManager.Create(user.FirstName, user.LastName, user.Pasword, user.Email); 
	if err != nil {
		log.Printf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())

		if userManErr, ok := err.(*userV1.Error); ok {
			// Upgrade an `error` interface to a `userV1.Error` interface so that we can use the `Type()` method to get the error type
			switch userManErr.Type() {
			case userV1.ErrTypeBadRequest:
				http.Error(w, fmt.Sprintf("Bad request: %s", userManErr.Error()), http.StatusBadRequest)
			case userV1.ErrTypeEmailHasBeenUsed:
				http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusConflict)
			case userV1.ErrTypeInternalServerErr:
				http.Error(w, fmt.Sprintf("Internal server error, please retry later."), http.StatusInternalServerError)
			default:
				http.Error(w, fmt.Sprintf("Unknown error, please retry later."), http.StatusInternalServerError) 
			}
		} else {
			// This should never happen
			http.Error(w, fmt.Sprintf("Unknown error, please retry later."), http.StatusInternalServerError)
		}
	}
	// Return ID
    json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
    w.WriteHeader(http.StatusOK)
    return
}
