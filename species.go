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

type Location struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type Species struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SpeciesName string             `json:"species_name" bson:"species_name"`
	Category    primitive.ObjectID `json:"category" bson:"category"`
	Image       string             `json:"image" bson:"image"`
	Location    Location           `json:"location" bson:"location"`
}

var speciesCollection *mongo.Collection

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

	speciesCollection = client.Database("palvelinohjelmointi").Collection("species")
}

func getAllSpecies(w http.ResponseWriter, r *http.Request) {
	var species []Species
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := speciesCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var specie Species
		cursor.Decode(&specie)
		species = append(species, specie)
	}
	json.NewEncoder(w).Encode(species)
}

func getSpecies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var specie Species
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = speciesCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&specie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(specie)
}

func createSpecies(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	var specie Species
	json.Unmarshal(reqBody, &specie)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := speciesCollection.InsertOne(ctx, specie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func updateSpecies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	reqBody, _ := io.ReadAll(r.Body)
	var updatedSpecie Species
	json.Unmarshal(reqBody, &updatedSpecie)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = speciesCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updatedSpecie})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updatedSpecie)
}

func deleteSpecies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = speciesCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSpeciesRequests(router *mux.Router) {
	router.HandleFunc("/species", getAllSpecies).Methods("GET")
	router.HandleFunc("/species/{id}", getSpecies).Methods("GET")
	router.HandleFunc("/species", createSpecies).Methods("POST")
	router.HandleFunc("/species/{id}", updateSpecies).Methods("PUT")
	router.HandleFunc("/species/{id}", deleteSpecies).Methods("DELETE")
}
