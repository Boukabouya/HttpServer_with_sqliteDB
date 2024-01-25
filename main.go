package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// Person struct to represent a person in the 'persons' table
type Person struct {
	ID     *int   `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

var db *sql.DB // Declare the variable to hold the database connection

// Function to create the 'persons' table

func CreatePersonsTable(db *sql.DB) error {
	const query = `
	CREATE TABLE IF NOT EXISTS persons (
		id INTEGER NOT NULL PRIMARY KEY,
		name CHAR(40) NOT NULL,
		email CHAR(50),
		mobile CHAR(25) NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating 'persons' table: %v", err)
	}
	return err
}

// Function to insert a person into the 'persons' table
// the id will be insert auto in sqlite without need to insert it 
func InsertPerson(db *sql.DB, person Person) (sql.Result, error) {
	insertDataQuery := `
	INSERT INTO persons (name, email, mobile)
	VALUES (?, ?, ?);`

	result, err := db.Exec(insertDataQuery, person.Name, person.Email, person.Mobile)
	if err != nil {
		log.Printf("Error inserting person: %v", err)
	}
	return result, err
}

// Function to get all persons from the 'persons' table
func GetAllPersons(db *sql.DB) ([]Person, error) {
	const query = `SELECT id, name, email, mobile FROM persons;`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error getting all persons: %v", err)
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.Mobile); err != nil {
			log.Printf("Error scanning person: %v", err)
			return nil, err
		}
		persons = append(persons, p)
	}

	return persons, nil
}

// Function to get a person by ID from the 'persons' table
func GetPersonByID(db *sql.DB, id int) (Person, error) {
	const query = `SELECT id, name, email, mobile FROM persons WHERE id = ?;`

	row := db.QueryRow(query, id)
	var p Person
	err := row.Scan(&p.ID, &p.Name, &p.Email, &p.Mobile)
	if err != nil {
		log.Printf("Error getting person by ID: %v", err)
		return Person{}, err
	}

	return p, nil
}

// Function to update a person in the 'persons' table
func UpdatePerson(db *sql.DB, person Person) (sql.Result, error) {
	updateDataQuery := `
        UPDATE persons
        SET name=?, email=?, mobile=?
        WHERE id=?;`

	// Check if ID is nil
	if person.ID == nil {
		return nil, fmt.Errorf("person ID is nil, cannot update")
	}

	fmt.Printf("Update Query: %s\n", updateDataQuery)
	fmt.Printf("Person ID: %d\n", *person.ID)

	result, err := db.Exec(updateDataQuery, person.Name, person.Email, person.Mobile, *person.ID)
	if err != nil {
		log.Printf("Error updating person: %v", err)
		return nil, fmt.Errorf("failed to update person with ID %d: %v", *person.ID, err)
	}

	return result, nil
}

// Function to delete a person from the 'persons' table
func DeletePerson(db *sql.DB, id int) (sql.Result, error) {
	deleteDataQuery := `
	DELETE FROM persons
	WHERE id=?;`

	result, err := db.Exec(deleteDataQuery, id)
	if err != nil {
		log.Printf("Error deleting person: %v", err)
	}
	return result, err
}

// Function to check the status of the application
func checkHandler(w http.ResponseWriter, r *http.Request) {
	// Respond with a success message
	status := map[string]string{"status": "OK"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)

	// Print a message indicating that the "/check" endpoint is handled
	fmt.Println("Check endpoint handled successfully.")
}

// Function to handle requests for getting a person by ID
func getPersonHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request URL to get the person ID
	personID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid person ID", http.StatusBadRequest)
		return
	}

	// Get the person by ID from the 'persons' table
	person, err := GetPersonByID(db, personID)
	if err != nil {
		http.Error(w, "Person not found", http.StatusNotFound)
		return
	}

	// Respond with the person details
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

// Function to handle requests for listing all persons
func listPersonsHandler(w http.ResponseWriter, r *http.Request) {
	// Get all persons from the 'persons' table
	persons, err := GetAllPersons(db)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with the list of persons as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
}

// Function to handle requests for creating a new person
func createPersonHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the new person data
	var newPerson Person
	err := json.NewDecoder(r.Body).Decode(&newPerson)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Insert the new person into the 'persons' table
	_, err = InsertPerson(db, newPerson)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "Person created successfully"}`)
}

// Function to handle requests for updating a person
func updatePersonHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request URL to get the person ID
	personID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid person ID", http.StatusBadRequest)
		return
	}

	fmt.Printf("Updating person with ID: %d\n", personID) // Add this print statement

	// Get the existing person from the 'persons' table
	existingPerson, err := GetPersonByID(db, personID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Existing person: %+v\n", existingPerson) // Add this print statement

	// Parse the request body to get the updated person data
	var updatedPerson Person
	err = json.NewDecoder(r.Body).Decode(&updatedPerson)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	fmt.Printf("Updated person data: %+v\n", updatedPerson) // Add this print statement

	// Update the person in the 'persons' table
	_, err = UpdatePerson(db, updatedPerson)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "Person updated successfully"}`)
}

// Function to handle requests for deleting a person
func deletePersonHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request URL to get the person ID
	personID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid person ID", http.StatusBadRequest)
		return
	}

	// Delete the person from the 'persons' table
	_, err = DeletePerson(db, personID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "Person deleted successfully"}`)
}

func main() {
	// Open the database connection
	var err error
	db, err = sql.Open("sqlite3", "myDataBase.db")
	if err != nil {
		log.Fatalf("Error opening the database: %v", err)
	}
	defer db.Close()

	// Create the 'persons' table
	err = CreatePersonsTable(db)
	if err != nil {
		log.Fatalf("Error creating 'persons' table: %v", err)
	}

	// Define HTTP endpoints and corresponding handlers
	http.HandleFunc("/check", checkHandler)
	http.HandleFunc("/list-persons", listPersonsHandler)
	http.HandleFunc("/get-person", getPersonHandler)
	http.HandleFunc("/create-person", createPersonHandler)
	http.HandleFunc("/update-person", updatePersonHandler)
	http.HandleFunc("/delete-person", deletePersonHandler)

	// Define default port
	defaultPort := "8081"

	// Check if a command-line argument is provided
	if len(os.Args) == 3 && os.Args[1] == "serve" {
		port := os.Args[2]
		fmt.Printf("Starting HTTP server on port %s...\n", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	} else {
		fmt.Printf("Starting HTTP server on default port %s...\n", defaultPort)
		log.Fatal(http.ListenAndServe(":"+defaultPort, nil))
	}
}
/*
To run this code you should run without debbiging to make the server open the port and then you could 
build an execusion file by go build in the terminal 
After we test our code by this cmd :
for Check Endpoint:
curl http://localhost:8081/check // you could change the port by 8080 for exemple

List Persons Endpoint:
curl http://localhost:8081/list-persons
Get Person by ID (Replace 1 with the desired ID):
curl http://localhost:8081/get-person?id=1
Create Person:
curl -X POST -H "Content-Type: application/json" -d '{"name": "Rayene Amina", "email": "Rayene@gmail.com", "mobile": "05842154"}' http://localhost:8081/create-person
Update Person (Replace 1 with the desired ID):
curl -X POST -H "Content-Type: application/json" -d '{"id": 1, "name": "Rayene Amina", "email": "Brayeneamina18@gmail.com", "mobile": "05842154"}' http://localhost:8081/update-person?id=1
Delete Person (Replace 1 with the desired ID):
curl -X POST http://localhost:8081/delete-person?id=1
*/