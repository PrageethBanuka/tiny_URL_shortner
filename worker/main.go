package worker

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool" // Upgraded to v5 to match your core-api
)

func main() {
	// 1. Pull the exact same environment variables OpenChoreo injects into your API
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Fallbacks just in case you run this locally outside the cluster
	if host == "" { host = "core-db-postgresql" }
	if port == "" { port = "5432" }
	if user == "" { user = "postgres" }
	if dbname == "" { dbname = "coreapi" }
	if sslmode == "" { sslmode = "disable" }

	// 2. Construct the exact DSN string your core-api uses
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)

	ctx := context.Background()

	// 3. Connect to the pool using pgx/v5
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	log.Println("CronJob started: Running database optimization...")
	
	// 4. Execute the exact cleanup query you wrote in store.go
	query := `DELETE FROM links WHERE expires_at < NOW()`
	commandTag, err := pool.Exec(ctx, query)
	if err != nil {
		log.Fatalf("Error cleaning up expired links: %v\n", err)
	}
	
	fmt.Printf("Successfully purged %d expired links.\n", commandTag.RowsAffected())
	log.Println("CronJob complete. Shutting down gracefully.")
}