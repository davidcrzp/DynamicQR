package api

import (
	"fmt"
	"net/http"
	"encoding/json"
	"sync"
	"strconv"
    // "github.com/davidcrzp/dynamicqr/internal/database"
)

// type Server struct {
//     store *database.Store
// }
//
// type Params struct {
// 	name string
// }
//
// type Resp struct {
// 	name string
// }

type User struct {
	Name string `json:"name"`
}

func NewServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("POST /users", handleCreateUser)
	mux.HandleFunc("GET /users/{id}", handleGetUser)
	mux.HandleFunc("DELETE /users/{id}", handleDeleteUser)
	mux.HandleFunc("PUT /users/{id}", handlePutUser)

	mux.HandleFunc("GET /qr/{id}", handleDynamicQR)
	mux.HandleFunc("GET /example", handleExample)

	return mux
}

func StartServer(srv *http.ServeMux, port string) error {
	return http.ListenAndServe(port, srv)
}

var userCache = make(map[int]User)

var cacheMutex sync.RWMutex

func handleRoot(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./dynamic_qr.png")
}

func handleDynamicQR(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", r.PathValue("id"))
}

func handleExample(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/example.html")
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	userCache[len(userCache)+1] = user // database
	cacheMutex.Unlock()


	w.WriteHeader(http.StatusNoContent)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	cacheMutex.RLock()
	user, ok := userCache[id] // database
	cacheMutex.RUnlock()
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("content-type", "application/json")
	j, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(j)
}


func handlePutUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	
	if _, ok := userCache[id]; !ok { // database
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	userCache[id] = user // database
	cacheMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if _, ok := userCache[id]; !ok { // database
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	delete(userCache, id) // database
	cacheMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
