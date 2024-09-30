package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	ctx := context.Background()

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		fmt.Println("Failed to create storage client:", err)
		return
	}

	http.HandleFunc("/upload", handleUpload(client, ctx))
	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func handleUpload(client *storage.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to get file from request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		objectName := uuid.New().String() + ".pdf"

		wc := client.Bucket(os.Getenv("BUCKETNAME")).Object(objectName).NewWriter(ctx)

		_, err = io.Copy(wc, file)
		if err != nil {
			http.Error(w, "Failed to copy file to GCS", http.StatusInternalServerError)
			return
		}

		err = wc.Close()
		if err != nil {
			http.Error(w, "Failed to close writer", http.StatusInternalServerError)
			return
		}

		fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", os.Getenv("BUCKETNAME"), objectName)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"file": "%s"}`, fileURL)
	}
}
