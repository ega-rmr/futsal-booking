package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"futsal-booking/db"
	"futsal-booking/handlers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("gagal baca .env: ", err)
	}

	conn, err := db.Connect()
	if err != nil {
		log.Fatal("gagal konek database: ", err)
	}
	defer conn.Close()

	err = db.CreateTables(conn)
	if err != nil {
		log.Fatal("gagal bikin tabel: ", err)
	}

	h := &handlers.Handler{DB: conn}

	http.HandleFunc("POST /register", h.Register)
	http.HandleFunc("POST /login", h.Login)

	http.HandleFunc("GET /courts", h.DaftarLapangan)
	http.HandleFunc("POST /courts", h.CekToken(h.TambahLapangan))

	http.HandleFunc("POST /bookings", h.CekToken(h.BuatBooking))
	http.HandleFunc("GET /bookings", h.CekToken(h.BookingSaya))
	http.HandleFunc("DELETE /bookings/{id}", h.CekToken(h.BatalkanBooking))

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	port := os.Getenv("PORT")
	log.Println("server jalan di port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
