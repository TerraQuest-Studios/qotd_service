package main

import (
	"log"
	"net/http"
	"time"
	"encoding/json"
	"fmt"
	"os"
	"database/sql"
	_ "embed"

	"github.com/joho/godotenv"
	"github.com/go-co-op/gocron"
	"github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
	"github.com/go-sql-driver/mysql"
	"github.com/TerraQuest-Studios/qotd_service/quotes"
	"github.com/TerraQuest-Studios/qotd_service/response"
	"github.com/TerraQuest-Studios/qotd_service/webhook"
)

//go:embed logo.png
var logoPng []byte

func main() {
	fmt.Println("QOTD Service Starting...")

	fmt.Println("Loading .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fmt.Println(".env file loaded")

	fmt.Println("Connecting to database...")
	var db *sql.DB
	cfg := mysql.NewConfig()
    cfg.User = os.Getenv("DBUSER")
    cfg.Passwd = os.Getenv("DBPASS")
    cfg.Net = "tcp"
    cfg.Addr = "127.0.0.1:3306"
    cfg.DBName = os.Getenv("DBNAME")
    cfg.ParseTime = true
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
        log.Fatal(err)
    }

	pingErr := db.Ping()
    if pingErr != nil {
        log.Fatal(pingErr)
    }
    fmt.Println("Connected to database!")

	q := quotes.New(db)

	/* _ = func(ctx context.Context, Type string) (quotes.GetLatestQuoteByTypeRow, error) {
		q.ActivateOldestQuote(ctx, Type)
		quote, err := q.GetLatestQuoteByType(ctx, Type)
		return quote, err
	} */

	fmt.Println("Starting scheduler...")
	// Create a new scheduler in UTC
	s := gocron.NewScheduler(time.UTC)

	// Define a job to run every 2 seconds
	/* s.Every(2).Seconds().Do(func() {
		fmt.Println("Scheduled task running:", time.Now())
	}) */
	//set a job to run every day at 11:30pm est (translated to utc)
	/* s.Every(1).Day().At("04:30").Do(func() {
		fmt.Println("Scheduled task running at 11:30pm EST (04:30 UTC):", time.Now())
	}) */
	
	// Start the scheduler in a separate goroutine
	s.StartAsync()
	fmt.Println("Scheduler started")

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.DefaultResponse())
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.DefaultResponse())
	})
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		//redirect any requests to /?route={route}&type={type} to /api/v2/quote/{route}/{type}
		//for legacy compatibility
		route := r.URL.Query().Get("route")
		typeParam := r.URL.Query().Get("type")
		if route != "" && typeParam != "" {
			http.Redirect(w, r, fmt.Sprintf("/api/v2/quote/%s/%s", route, typeParam), http.StatusMovedPermanently)
			return
		}

		//add json header
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.DefaultResponse())
	})
	r.Get("/api/v2/quote/{route}/{type}", func(w http.ResponseWriter, r *http.Request) {
		routeParam := chi.URLParam(r, "route")
		typeParam := chi.URLParam(r, "type")

		//check that type exists
		exists, err := q.TypeExists(r.Context(), typeParam)
		if err != nil {
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/json")

			if os.Getenv("DEBUG") == "true" {
				json.NewEncoder(w).Encode(response.ServerErrorResponse(err.Error()))
			} else {
				json.NewEncoder(w).Encode(response.ServerErrorResponse("error checking type existence"))
			}

			return
		}
		if !exists {
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response.Response{
				Success: false,
				Message: "type does not exist",
				Data:    map[string]interface{}{},
			})
			return
		}

		if routeParam == "random" {
			quote, err := q.GetRandomQuoteByType(r.Context(), typeParam)
			if err != nil {
				w.WriteHeader(500)
				w.Header().Set("Content-Type", "application/json")

				if os.Getenv("DEBUG") == "true" {
					json.NewEncoder(w).Encode(response.ServerErrorResponse(err.Error()))
				} else {
					json.NewEncoder(w).Encode(response.ServerErrorResponse("error fetching quote"))
				}
				return
			}

			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			enc.Encode(response.Response{
				Success: true,
				Message: "have a random quote",
				Data:    map[string]interface{}{"quote": quote.Quote},
			})
			return
		} else if routeParam == "latest" {
			quote, err := q.GetLatestQuoteByType(r.Context(), typeParam)
			if err != nil {
				w.WriteHeader(500)
				w.Header().Set("Content-Type", "application/json")

				if os.Getenv("DEBUG") == "true" {
					json.NewEncoder(w).Encode(response.ServerErrorResponse(err.Error()))
				} else {
					json.NewEncoder(w).Encode(response.ServerErrorResponse("error fetching quote"))
				}
				return
			}

			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			enc.Encode(response.Response{
				Success: true,
				Message: "have the latest quote",
				Data:    map[string]interface{}{"quote": quote.Quote},
			})
			return
		}

		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.Response{
			Success: false,
			Message: "invalid route parameter",
			Data:    map[string]interface{}{},
		})
		return
	})
	r.Get("/assets/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(logoPng)
		return
	})
	/* r.Get("/api/v2/dev/sendtestwebhook", func(w http.ResponseWriter, r *http.Request) {
		webhookURL := os.Getenv("WEBHOOK_URL")
		payload := webhook.Payload{
			Content: "This is a test webhook message from QOTD Service!",
			UserName: "QOTD Bot",
			AvatarURL: "https://qotd.terraqueststudios.net/assets/logo.png",
		}
		err := webhook.Exec(webhookURL, payload)
		if err != nil {
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/json")

			if os.Getenv("DEBUG") == "true" {
				json.NewEncoder(w).Encode(response.ServerErrorResponse(err.Error()))
			} else {
				json.NewEncoder(w).Encode(response.ServerErrorResponse("error sending test webhook"))
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.Response{
			Success: true,
			Message: "test webhook sent successfully",
			Data:    map[string]interface{}{},
		})
		return
	}) */
	log.Println("Starting server on :1780")
    http.ListenAndServe(":1780", r)
}