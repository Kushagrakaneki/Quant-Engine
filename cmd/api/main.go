package main
import(
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

)

func main(){

	r:=chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60*time.Second))

	r.Get("/health",func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"Quant-Lite Matcher is ALIVE"}`))
	})

	port:=":8080"
	fmt.Printf("Starting Quant-Lite Trading Engine on port %s...\n", port)

	err:=http.ListenAndServe(port,r)
	if err!=nil{
		fmt.Printf("CRITICAL: Server crashed: %v\n", err)
	}


}