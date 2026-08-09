package main

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go/v4"
	"flag"
	"fmt"
	"github.com/lunjae/jasen-jela-web/backend/internal/models"
	"github.com/lunjae/jasen-jela-web/backend/internal/services"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

func main() {
	allowNonEmpty := flag.Bool("allow-non-empty", false, "allow adding seed data when products or categories already exist")
	allowProduction := flag.Bool("allow-production", false, "allow running against APP_ENV=production")
	flag.Parse()
	ctx := context.Background()
	project := os.Getenv("FIREBASE_PROJECT_ID")
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if project == "" {
		log.Fatal("FIREBASE_PROJECT_ID is required")
	}
	if os.Getenv("APP_ENV") == "production" && !*allowProduction {
		log.Fatal("refusing to seed production; pass --allow-production explicitly")
	}
	var options []option.ClientOption
	if credentials != "" {
		options = append(options, option.WithCredentialsFile(credentials))
	}
	app, e := firebase.NewApp(ctx, &firebase.Config{ProjectID: project}, options...)
	if e != nil {
		log.Fatal(e)
	}
	client, e := app.Firestore(ctx)
	if e != nil {
		log.Fatal(e)
	}
	defer client.Close()
	if !*allowNonEmpty && (collectionHasDocuments(ctx, client, "categories") || collectionHasDocuments(ctx, client, "products")) {
		log.Fatal("refusing to seed a non-empty database; pass --allow-non-empty explicitly")
	}
	now := time.Now().UTC()
	cats := []models.Category{{Name: "Drveni sanduci", Description: "Klasični drveni modeli.", Published: true}, {Name: "Klasični modeli", Description: "Tradicionalni modeli od pažljivo odabranog drveta.", Published: true}, {Name: "Premium modeli", Description: "Reprezentativni modeli sa unapređenom završnom obradom.", Published: true}}
	ids := []string{}
	for _, c := range cats {
		ref := client.Collection("categories").NewDoc()
		c.ID = ref.ID
		c.Slug = services.Slug(c.Name)
		c.CreatedAt = now
		c.UpdatedAt = now
		if _, e = ref.Set(ctx, c); e != nil {
			log.Fatal(e)
		}
		ids = append(ids, ref.ID)
	}
	prices := []float64{52000, 78000, 96000, 64000}
	products := []models.Product{{Name: "Mirni Jasen", ShortDescription: "Klasičan model toplog tona i odmerene završne obrade.", Description: "Model izrađen od jasenovog drveta, sa pažljivo obrađenim detaljima i postojanom završnom zaštitom.", CategoryID: ids[1], Material: "Jasen", Color: "Prirodni orah", Price: &prices[0], Currency: "RSD", Featured: true, Published: true}, {Name: "Tiha Jela", ShortDescription: "Model čistih linija od kvalitetnog drveta jele.", Description: "Pouzdan i odmeren model, namenjen porodicama koje biraju jednostavnost i dostojanstven izgled.", CategoryID: ids[0], Material: "Jela", Color: "Svetli hrast", Price: &prices[1], Currency: "RSD", Published: true}, {Name: "Večni Hrast", ShortDescription: "Premium model od hrasta sa bogatom završnom obradom.", Description: "Reprezentativan model izrađen od odabranog hrasta, sa preciznim spojevima i diskretnim ukrasnim detaljima.", CategoryID: ids[2], Material: "Hrast", Color: "Tamni orah", Price: &prices[2], Currency: "RSD", Featured: true, Published: true}, {Name: "Dostojanstvo", ShortDescription: "Sveden model uravnoteženih proporcija i pouzdane izrade.", Description: "Model namenjen tradicionalnom izboru, izrađen uz strogu kontrolu materijala i završne obrade.", CategoryID: ids[1], Material: "Bukva", Color: "Mahagoni", Price: &prices[3], Currency: "RSD", Published: true}}
	for _, p := range products {
		ref := client.Collection("products").NewDoc()
		p.ID = ref.ID
		p.Slug = services.Slug(p.Name)
		p.Images = []models.ProductImage{}
		p.CreatedAt = now
		p.UpdatedAt = now
		if _, e = ref.Set(ctx, p); e != nil {
			log.Fatal(e)
		}
	}
	fmt.Printf("Seeded %d categories and %d products.\n", len(cats), len(products))
}

func collectionHasDocuments(ctx context.Context, client *firestore.Client, collection string) bool {
	it := client.Collection(collection).Limit(1).Documents(ctx)
	defer it.Stop()
	_, err := it.Next()
	if err == iterator.Done {
		return false
	}
	if err != nil {
		log.Fatal(err)
	}
	return true
}
