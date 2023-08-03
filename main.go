package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	b, err := base64.StdEncoding.DecodeString(os.Getenv("FIREBASE_ADMIN_SDK_BASE64"))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	clientOption := option.WithCredentialsJSON(b)
	app, err := firebase.NewApp(ctx, nil, clientOption)
	if err != nil {
		log.Fatalln(err)
	}

	fs, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	defer fs.Close()

	router := gin.Default()
	albums := router.Group("/albums")
	{
		albums.POST("/", func(ctx *gin.Context) {
			var album Album
			if err := ctx.BindJSON(&album); err != nil {
				ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "bad request"})
				return
			}
			docRef := fs.Collection("albums").NewDoc()
			album.ID = docRef.ID
			_, err := docRef.Set(ctx, album)
			if err != nil {
				return
			}
		})

		albums.GET("/", func(ctx *gin.Context) {
			albums := make([]Album, 0)
			iter := fs.Collection("albums").Documents(ctx)
			for {
				doc, err := iter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return
				}
				var album Album
				err = doc.DataTo(&album)
				if err != nil {
					return
				}
				albums = append(albums, album)
			}
			ctx.IndentedJSON(http.StatusOK, albums)
		})

		albums.GET("/:id", func(ctx *gin.Context) {
			id := ctx.Param("id")
			ds, err := fs.Collection("albums").Doc(id).Get(ctx)
			if err != nil {
				ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
				return
			}
			var album Album
			ds.DataTo(&album)
			ctx.IndentedJSON(http.StatusOK, album)

		})
	}

	router.Run("localhost:8080")
}
