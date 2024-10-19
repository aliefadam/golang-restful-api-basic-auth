package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	username = "admin"
	password = "admin"
)

type UlasanWhatsapp struct {
	ReviewID string `json:"reviewID"`
	Content  string `json:"content"`
	Score    string `json:"score"`
}

// Fungsi untuk Basic Authentication
func cekAutentikasi(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return false
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	authPair := strings.SplitN(string(payload), ":", 2)
	if len(authPair) != 2 {
		return false
	}

	expectedUsername := username
	expectedPassword := password

	return authPair[0] == expectedUsername && authPair[1] == expectedPassword
}

func koneksiDatabase() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/app-review")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func halamanUtama(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "<h3>Selamat Datang Di Ulasan Whatsapp API</h3>")
	fmt.Fprintln(w, "<ul>")
	fmt.Fprintln(w, "<li>(GET) /whatsapp : Ambil semua ulasan</li>")
	fmt.Fprintln(w, "<li>(POST) /whatsapp : Tambah ulasan</li>")
	fmt.Fprintln(w, "<li>(PUT) /whatsapp/{reviewID} : Update ulasan</li>")
	fmt.Fprintln(w, "<li>(DELETE) /whatsapp/{reviewID} : Hapus ulasan</li>")
	fmt.Fprintln(w, "</ul>")
}

// Ambil semua ulasan
func ambilSemuaUlasan(w http.ResponseWriter, r *http.Request) {
	db, err := koneksiDatabase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM whatsapp")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var daftarUlasan []UlasanWhatsapp
	for rows.Next() {
		var ulasan UlasanWhatsapp
		if err := rows.Scan(&ulasan.ReviewID, &ulasan.Content, &ulasan.Score); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		daftarUlasan = append(daftarUlasan, ulasan)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(daftarUlasan)
}

// Tambah ulasan baru
func tambahUlasan(w http.ResponseWriter, r *http.Request) {
	var ulasan UlasanWhatsapp
	err := json.NewDecoder(r.Body).Decode(&ulasan)
	if err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
	}

	db, err := koneksiDatabase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO whatsapp(reviewID, content, score) VALUES(?, ?, ?)", ulasan.ReviewID, ulasan.Content, ulasan.Score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"pesan": "Berhasil menambahkan ulasan"})
}

// Update ulasan berdasarkan ID
func updateUlasan(w http.ResponseWriter, r *http.Request, id string) {
	var ulasan UlasanWhatsapp
	err := json.NewDecoder(r.Body).Decode(&ulasan)
	if err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
	}

	db, err := koneksiDatabase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE whatsapp SET content=?, score=? WHERE reviewID=?", ulasan.Content, ulasan.Score, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"pesan": "Berhasil mengupdate ulasan"})
}

// Hapus ulasan berdasarkan ID
func hapusUlasan(w http.ResponseWriter, r *http.Request, id string) {
	db, err := koneksiDatabase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM whatsapp WHERE reviewID=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"pesan": "Berhasil menghapus ulasan"})
}

func handler(w http.ResponseWriter, r *http.Request) {
	if !cekAutentikasi(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"pesan": "Tidak ada autentikasi", "error": "Unauthorized"})
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if path == "" {
		halamanUtama(w, r)
		return
	}

	if len(parts) == 1 && parts[0] == "whatsapp" {
		switch r.Method {
		case "GET":
			ambilSemuaUlasan(w, r)
		case "POST":
			tambahUlasan(w, r)
		default:
			http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 2 && parts[0] == "whatsapp" {
		id := parts[1]
		switch r.Method {
		case "PUT":
			updateUlasan(w, r, id)
		case "DELETE":
			hapusUlasan(w, r, id)
		default:
			http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Endpoint tidak ditemukan", http.StatusNotFound)
	}
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Berhasil menjalankan server di http://localhost:5173")
	if err := http.ListenAndServe(":5173", nil); err != nil {
		log.Fatal(err)
	}
}
