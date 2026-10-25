package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type registerInput struct {
	Nama     string `json:"nama"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input registerInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		kirimError(w, 400, "data tidak valid")
		return
	}

	if input.Nama == "" || input.Email == "" || input.Password == "" {
		kirimError(w, 400, "nama, email, dan password wajib diisi")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		kirimError(w, 500, "gagal memproses password")
		return
	}

	_, err = h.DB.Exec(
		"INSERT INTO users (nama, email, password) VALUES ($1, $2, $3)",
		input.Nama, input.Email, string(hash),
	)
	if err != nil {
		kirimError(w, 400, "email sudah terdaftar")
		return
	}

	kirimJSON(w, 201, map[string]string{"pesan": "register berhasil"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		kirimError(w, 400, "data tidak valid")
		return
	}

	var id int
	var hash string
	err = h.DB.QueryRow(
		"SELECT id, password FROM users WHERE email = $1", input.Email,
	).Scan(&id, &hash)
	if err != nil {
		kirimError(w, 401, "email atau password salah")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password))
	if err != nil {
		kirimError(w, 401, "email atau password salah")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		kirimError(w, 500, "gagal membuat token")
		return
	}

	kirimJSON(w, 200, map[string]string{"token": tokenString})
}
