package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Booking struct {
	ID      int       `json:"id"`
	CourtID int       `json:"court_id"`
	Mulai   time.Time `json:"mulai"`
	Selesai time.Time `json:"selesai"`
	Status  string    `json:"status"`
}

func (h *Handler) BuatBooking(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	var input Booking
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		kirimError(w, 400, "data tidak valid")
		return
	}

	if input.CourtID == 0 || !input.Selesai.After(input.Mulai) {
		kirimError(w, 400, "court_id wajib diisi dan jam selesai harus setelah jam mulai")
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		kirimError(w, 500, "gagal memulai transaksi")
		return
	}
	defer tx.Rollback()

	// kunci baris lapangan supaya dua booking bersamaan tidak lolos barengan
	var courtID int
	err = tx.QueryRow("SELECT id FROM courts WHERE id = $1 FOR UPDATE", input.CourtID).Scan(&courtID)
	if err == sql.ErrNoRows {
		kirimError(w, 404, "lapangan tidak ditemukan")
		return
	}
	if err != nil {
		kirimError(w, 500, "gagal membaca lapangan")
		return
	}

	// bentrok kalau jam mulai lama < selesai baru DAN selesai lama > mulai baru
	var jumlah int
	err = tx.QueryRow(
		`SELECT COUNT(*) FROM bookings
		 WHERE court_id = $1 AND status = 'aktif' AND mulai < $3 AND selesai > $2`,
		input.CourtID, input.Mulai, input.Selesai,
	).Scan(&jumlah)
	if err != nil {
		kirimError(w, 500, "gagal mengecek jadwal")
		return
	}
	if jumlah > 0 {
		kirimError(w, 409, "jadwal bentrok dengan booking lain")
		return
	}

	err = tx.QueryRow(
		`INSERT INTO bookings (user_id, court_id, mulai, selesai)
		 VALUES ($1, $2, $3, $4) RETURNING id, status`,
		userID, input.CourtID, input.Mulai, input.Selesai,
	).Scan(&input.ID, &input.Status)
	if err != nil {
		kirimError(w, 500, "gagal menyimpan booking")
		return
	}

	err = tx.Commit()
	if err != nil {
		kirimError(w, 500, "gagal menyimpan booking")
		return
	}

	kirimJSON(w, 201, input)
}

func (h *Handler) BookingSaya(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	rows, err := h.DB.Query(
		`SELECT id, court_id, mulai, selesai, status
		 FROM bookings WHERE user_id = $1 ORDER BY mulai`,
		userID,
	)
	if err != nil {
		kirimError(w, 500, "gagal mengambil data")
		return
	}
	defer rows.Close()

	bookings := []Booking{}
	for rows.Next() {
		var b Booking
		err = rows.Scan(&b.ID, &b.CourtID, &b.Mulai, &b.Selesai, &b.Status)
		if err != nil {
			kirimError(w, 500, "gagal membaca data")
			return
		}
		bookings = append(bookings, b)
	}

	kirimJSON(w, 200, bookings)
}

func (h *Handler) BatalkanBooking(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		kirimError(w, 400, "id tidak valid")
		return
	}

	hasil, err := h.DB.Exec(
		"UPDATE bookings SET status = 'batal' WHERE id = $1 AND user_id = $2 AND status = 'aktif'",
		id, userID,
	)
	if err != nil {
		kirimError(w, 500, "gagal membatalkan booking")
		return
	}

	jumlah, _ := hasil.RowsAffected()
	if jumlah == 0 {
		kirimError(w, 404, "booking tidak ditemukan")
		return
	}

	kirimJSON(w, 200, map[string]string{"pesan": "booking dibatalkan"})
}
