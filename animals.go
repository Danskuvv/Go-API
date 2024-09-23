package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LocationAnimals struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type Animal struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	AnimalName string             `json:"animal_name" bson:"animal_name"`
	Species    primitive.ObjectID `json:"species" bson:"species"`
	Image      string             `json:"image" bson:"image"`
	Birthdate  time.Time          `json:"birthdate" bson:"birthdate"`
	Location   LocationAnimals    `json:"location" bson:"location"`
	Owner      primitive.ObjectID `json:"owner" bson:"owner"`
}

var animalCollection *mongo.Collection

func init() {
	clientOptions := options.Client().ApplyURI("mongodb+srv://danielvv:koira7mongo@cluster0.lhg1h.mongodb.net/palvelinohjelmointi?retryWrites=true&w=majority&appName=Cluster0")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	animalCollection = client.Database("palvelinohjelmointi").Collection("animals")
}

func getAllAnimals(w http.ResponseWriter, r *http.Request) {
	var animals []Animal
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := animalCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var animal Animal
		cursor.Decode(&animal)
		animals = append(animals, animal)
	}
	json.NewEncoder(w).Encode(animals)
}

func getAnimal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var animal Animal
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = animalCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&animal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(animal)
}

func createAnimal(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	var animal Animal
	json.Unmarshal(reqBody, &animal)

	// Convert species and owner to ObjectID
	speciesID, err := primitive.ObjectIDFromHex(animal.Species.Hex())
	if err != nil {
		http.Error(w, "Invalid species ID", http.StatusBadRequest)
		return
	}
	animal.Species = speciesID

	ownerID, err := primitive.ObjectIDFromHex(animal.Owner.Hex())
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	}
	animal.Owner = ownerID

	// Ensure location type and coordinates are set correctly
	if animal.Location.Type == "" || animal.Location.Coordinates == nil {
		http.Error(w, "Invalid location data", http.StatusBadRequest)
		return
	}

	// Parse birthdate
	animal.Birthdate, err = time.Parse("2006-01-02", animal.Birthdate.Format("2006-01-02"))
	if err != nil {
		http.Error(w, "Invalid birthdate format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := animalCollection.InsertOne(ctx, animal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func updateAnimal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	reqBody, _ := io.ReadAll(r.Body)
	var updatedAnimal Animal
	json.Unmarshal(reqBody, &updatedAnimal)

	// Convert species and owner to ObjectID
	speciesID, err := primitive.ObjectIDFromHex(updatedAnimal.Species.Hex())
	if err != nil {
		http.Error(w, "Invalid species ID", http.StatusBadRequest)
		return
	}
	updatedAnimal.Species = speciesID

	ownerID, err := primitive.ObjectIDFromHex(updatedAnimal.Owner.Hex())
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	}
	updatedAnimal.Owner = ownerID

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = animalCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updatedAnimal})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updatedAnimal)
}

func deleteAnimal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = animalCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleAnimalRequests(router *mux.Router) {
	router.HandleFunc("/animals", getAllAnimals).Methods("GET")
	router.HandleFunc("/animal/{id}", getAnimal).Methods("GET")
	router.HandleFunc("/animal", createAnimal).Methods("POST")
	router.HandleFunc("/animal/{id}", updateAnimal).Methods("PUT")
	router.HandleFunc("/animal/{id}", deleteAnimal).Methods("DELETE")
}
