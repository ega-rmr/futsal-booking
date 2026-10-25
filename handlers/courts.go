package handlers

import (
	"encoding/json"
	"net/http"
)

type Court struct {
	ID          int    `json:"id"`
	Nama        string `json:"nama"`
	HargaPerJam int    `json:"harga_per_jam"`
}

func (h *Handler) TambahLapangan(w http.ResponseWriter, r *http.Request) {
	var input Court
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		kirimError(w, 400, "data tidak valid")
		return
	}

	if input.Nama == "" || input.HargaPerJam <= 0 {
		kirimError(w, 400, "nama dan harga per jam wajib diisi")
		return
	}

	err = h.DB.QueryRow(
		"INSERT INTO courts (nama, harga_per_jam) VALUES ($1, $2) RETURNING id",
		input.Nama, input.HargaPerJam,
	).Scan(&input.ID)
	if err != nil {
		kirimError(w, 500, "gagal menyimpan lapangan")
		return
	}

	kirimJSON(w, 201, input)
}

func (h *Handler) DaftarLapangan(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, nama, harga_per_jam FROM courts ORDER BY id")
	if err != nil {
		kirimError(w, 500, "gagal mengambil data")
		return
	}
	defer rows.Close()

	courts := []Court{}
	for rows.Next() {
		var c Court
		err = rows.Scan(&c.ID, &c.Nama, &c.HargaPerJam)
		if err != nil {
			kirimError(w, 500, "gagal membaca data")
			return
		}
		courts = append(courts, c)
	}

	kirimJSON(w, 200, courts)
}
