package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment, Port, FrontendOrigin, FirebaseProjectID       string
	CloudinaryCloudName, CloudinaryAPIKey, CloudinaryAPISecret string
	UseMemoryStore                                             bool
	ReadTimeout, WriteTimeout, IdleTimeout                     time.Duration
}

func Load() (Config, error) {
	c := Config{
		Environment: env("APP_ENV", "development"), Port: env("PORT", "8080"),
		FrontendOrigin:      env("FRONTEND_ORIGIN", "http://localhost:5173"),
		FirebaseProjectID:   os.Getenv("FIREBASE_PROJECT_ID"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"), CloudinaryAPIKey: os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
		UseMemoryStore:      envBool("USE_MEMORY_STORE", false), ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}
	if strings.TrimSpace(c.FrontendOrigin) == "*" && c.Environment == "production" {
		return c, fmt.Errorf("wildcard CORS is forbidden in production")
	}
	if !c.UseMemoryStore && c.FirebaseProjectID == "" {
		return c, fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}
	if c.CloudinaryCloudName == "" || c.CloudinaryAPIKey == "" || c.CloudinaryAPISecret == "" {
		return c, fmt.Errorf("CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY and CLOUDINARY_API_SECRET are required")
	}
	return c, nil
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func envBool(k string, d bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	b, e := strconv.ParseBool(v)
	if e != nil {
		return d
	}
	return b
}
