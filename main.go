package main

import (
    // "context"
    // "crypto/ecdsa"
    "encoding/json"
    "log"
    "net/http"
    // "os"
	"fmt"

    // "github.com/ethereum/go-ethereum"
    // "github.com/ethereum/go-ethereum/common"
    // "github.com/ethereum/go-ethereum/crypto"
    // "github.com/ethereum/go-ethereum/ethclient"
    "github.com/gorilla/mux"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// User data structure
type User struct {
    gorm.Model
    Address   string `json:"userAddress" gorm:"uniqueIndex"`
    DataHash  string `json:"userData"`
    AccessCode string `json:"userAccessCode"`
}

// CORS middleware to allow cross-origin requests
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "https://idquick-satodas-dev.apps.sandbox-m2.ll9k.p1.openshiftapps.com") // allow requests from React frontend
		// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // allow requests from React frontend

		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Check if it's a preflight request and handle it
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Next
		next.ServeHTTP(w, r)
	})
}

// Connect to the PostgreSQL database
func initDB() *gorm.DB {
	// psql -U newuser -d satyajit

    dsn := "host=postgresql user=newuser password=password dbname=satyajit port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	// dsn = "host=localhost user=newuser password=password dbname=satyajit port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to connect database: %v", err)
    }

    // Migrate the schema
    db.AutoMigrate(&User{})
    return db
}

// // Connect to Ethereum client
// func initEthClient() *ethclient.Client {
//     client, err := ethclient.Dial(("https://mainnet.infura.io/v3/01824eaff1034d75a36503b74af45477"))
//     if err != nil {
//         log.Fatalf("Failed to connect to the Ethereum client: %v", err)
//     }
//     return client
// }

// Register user data
func registerUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("About to register user")
	// todo show the code for this and react to ChatGPT and ask to correct as currently the user isn't getting stored!
    var user User
	fmt.Println(r.Body)
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Check if user already exists
    var existingUser User
    db.First(&existingUser, "address = ?", user.Address)
    if existingUser.ID != 0 {
        user.ID = existingUser.ID // Use existing record
    }
	fmt.Println("About to register user", user.Address)
    // Here you would hash the user data and send it to the smart contract
    // For this example, we're using the hash as is and just storing it in the database
    result := db.Save(&user)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}

// Generate access code for a user
func generateAccessCode(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("generating access code!!!")
    // This endpoint assumes that the access code generation logic is handled by the smart contract
    // Here we'll simulate generating an access code and saving it in the database
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Simulate access code generation and update the user record
    user.AccessCode = "simulated-access-code" // Replace with real access code generation logic
    db.Model(&User{}).Where("address = ?", user.Address).Update("access_code", user.AccessCode)

    json.NewEncoder(w).Encode(map[string]string{"accessCode": user.AccessCode})
}

// Fetch user data with an access code
func fetchUserData(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
    accessCode := r.URL.Query().Get("accessCode")
    if accessCode == "" {
        http.Error(w, "Access code is required", http.StatusBadRequest)
        return
    }

    var user User
    db.Where("access_code = ?", accessCode).First(&user)
    if user.ID == 0 {
        http.Error(w, "Invalid access code", http.StatusNotFound)
        return
    }

    // The actual user data would be encrypted and stored off-chain
    // Here we just return the data hash
    json.NewEncoder(w).Encode(map[string]string{"dataHash": user.DataHash})
}

func main() {
    db := initDB()
    // ethClient := initEthClient()

    // // Unused in this example, but included for completeness
    // _ = ethClient

    router := mux.NewRouter()

    router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
        registerUser(db, w, r)
    }).Methods("POST")

    router.HandleFunc("/generate-access-code", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("generate access code called")
        generateAccessCode(db, w, r)
    }).Methods("POST")

    router.HandleFunc("/fetch-data", func(w http.ResponseWriter, r *http.Request) {
        fetchUserData(db, w, r)
    }).Methods("GET")

	// Wrap the router with the CORS middleware
    http.Handle("/", enableCORS(router))

    log.Println("Server started on port 8080")

    log.Fatal(http.ListenAndServe(":8080", enableCORS(router)))
}
