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

type Category struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CategoryName string             `json:"category_name" bson:"category_name"`
}

var categoryCollection *mongo.Collection

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

	categoryCollection = client.Database("palvelinohjelmointi").Collection("categories")
}

func getAllCategories(w http.ResponseWriter, r *http.Request) {
	var categories []Category
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := categoryCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var category Category
		cursor.Decode(&category)
		categories = append(categories, category)
	}
	json.NewEncoder(w).Encode(categories)
}

func getCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var category Category
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = categoryCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(category)
}

func createCategory(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := io.ReadAll(r.Body)
	var category Category
	json.Unmarshal(reqBody, &category)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := categoryCollection.InsertOne(ctx, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	reqBody, _ := io.ReadAll(r.Body)
	var updatedCategory Category
	json.Unmarshal(reqBody, &updatedCategory)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = categoryCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updatedCategory})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updatedCategory)
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = categoryCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleCategoryRequests(router *mux.Router) {
	router.HandleFunc("/categories", getAllCategories).Methods("GET")
	router.HandleFunc("/category/{id}", getCategory).Methods("GET")
	router.HandleFunc("/category", createCategory).Methods("POST")
	router.HandleFunc("/category/{id}", updateCategory).Methods("PUT")
	router.HandleFunc("/category/{id}", deleteCategory).Methods("DELETE")
}
