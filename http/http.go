package http

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/riraum/si-cheong/db"
	"github.com/riraum/si-cheong/security"
)

type Server struct {
	RootDir      string
	EmbedRootDir embed.FS
	DB           db.DB
	T            *template.Template
	Key          *[32]byte
}

func (s Server) getIndex(w http.ResponseWriter, r *http.Request) {
	par, err := parseRValuesMap(r)
	if err != nil {
		log.Fatalf("parse to map %v", err)
	}

	p, err := s.DB.ReadPosts(par)
	if err != nil {
		log.Fatalf("read posts: %v", err)
	}

	err = s.T.ExecuteTemplate(w, "index.html.tmpl", p)

	if err != nil {
		log.Fatalf("execute %v", err)
	}
}

func parseRValuesMap(r *http.Request) (map[string]string, error) {
	par := map[string]string{}

	if r.FormValue("sort") != "" {
		par["sort"] = r.FormValue("sort")
	}

	if r.FormValue("direction") != "" {
		par["direction"] = r.FormValue("direction")
	}

	if r.FormValue("author") != "" {
		par["author"] = r.FormValue("author")
	}

	return par, nil
}

func (s Server) getCSS(w http.ResponseWriter, _ *http.Request) {
	css, err := s.EmbedRootDir.ReadFile("static/pico.min.css")
	if err != nil {
		log.Fatalf("failed to read %v", err)
	}

	w.Header().Add("Content-Type", "text/css")
	fmt.Fprint(w, string(css))
}

func (s Server) getAPIPosts(w http.ResponseWriter, r *http.Request) {
	par, err := parseRValuesMap(r)
	if err != nil {
		log.Fatalf("parse to map %v", err)
	}

	p, err := s.DB.ReadPosts(par)
	if err != nil {
		log.Fatalf("read posts: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatalf("failed to encode %v", err)
	}
}

func parseDate(ti int64) time.Time {
	return time.Unix(ti, 0)
}

func parseRValues(r *http.Request) (db.Post, error) {
	var p db.Post

	// fmt.Println("id parse", r.PathValue("id"))

	if r.PathValue("id") != "" {
		ID, err := strconv.ParseFloat(r.PathValue("id"), 32)
		if err != nil {
			return p, fmt.Errorf("ID convert to float %w", err)
		}

		p.ID = float32(ID)
		// fmt.Println("ID", p.ID)
	}

	// fmt.Println("date parse", r.FormValue("date"))

	switch r.Method {
	case http.MethodPost:
		if r.FormValue("date") != "" {
			date := r.FormValue("date")

			time, err := time.Parse(time.DateOnly, date)
			if err != nil {
				return p, fmt.Errorf("date parse: %w", err)
			}

			// log.Println("time parse post:", time)
			p.Date = time.Unix()
		}
	case http.MethodGet:
		if r.FormValue("date") != "" {
			date := r.FormValue("date")

			time, err := time.Parse(time.DateOnly, date)
			if err != nil {
				return p, fmt.Errorf("date parse: %w", err)
			}

			// log.Println("time parse: get", time)
			p.ParsedDate = time
		}
	default:
	}

	// log.Println("author parse:", r.FormValue("author"))

	if r.FormValue("author") != "" {
		author, err := strconv.ParseFloat(r.FormValue("author"), 32)
		if err != nil {
			return p, fmt.Errorf("author convert to float: %w", err)
		}

		p.AuthorID = float32(author)
		// fmt.Println("AuthorID", p.AuthorID)
	}

	p.Title = r.FormValue("title")
	// log.Println("title parse:", p.Title)
	p.Link = r.FormValue("link")
	// log.Println("link parse:", p.Link)
	p.Content = r.FormValue("content")
	// log.Println("content parse:", p.Content)

	// fmt.Println("post parse", p)

	return p, nil
}

func (s Server) postAPIPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("authorName")
	if err != nil {
		log.Fatal("no author cookie", err)
	}

	if !s.authenticated(r, w) {
		return
	}

	p, err := parseRValues(r)
	if err != nil {
		log.Fatalf("failed to parse values: %v", err)
	}

	encryptedAuthorByte, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable)

		err = json.NewEncoder(w).Encode(cookie.Value)
		if err != nil {
			log.Fatalf("failed to encode %v", err)
		}

		return
	}

	decryptedAuthorByte, err := security.Decrypt(encryptedAuthorByte, s.Key)
	if err != nil {
		log.Fatalf("failed to decrypt: %v", err)
	}

	authorID, err := s.DB.AuthorNametoID(string(decryptedAuthorByte))
	if err != nil {
		http.Redirect(w, r, "/fail?reason=authorCookieError", http.StatusUnauthorized)
		log.Fatalf("failed to decode base64 string to byte: %v", err)

		return
	}

	p.AuthorID = authorID

	err = s.DB.NewPost(p)
	if err != nil {
		log.Fatalf("create new post in db: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatalf("failed to encode %v", err)
	}
}

func (s Server) postPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("authorName")
	if err != nil {
		log.Fatal("no author cookie", err)
	}

	if !s.authenticated(r, w) {
		return
	}

	p, err := parseRValues(r)
	if err != nil {
		log.Fatalf("failed to parse values: %v", err)
	}

	encryptedAuthorByte, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		log.Fatalf("failed to decode base64 string to byte: %v", err)
	}

	decryptedAuthorByte, err := security.Decrypt(encryptedAuthorByte, s.Key)
	if err != nil {
		log.Fatalf("failed to decrypt: %v", err)
	}

	authorID, err := s.DB.AuthorNametoID(string(decryptedAuthorByte))
	if err != nil {
		http.Redirect(w, r, "/fail?reason=authorCookieError", http.StatusUnauthorized)
		log.Fatalf("failed string to float conversion: %v", err)

		return
	}

	p.AuthorID = authorID
	// fmt.Println("postAPIPost AuthorID", p.AuthorID)

	err = s.DB.NewPost(p)
	if err != nil {
		log.Fatalf("create new post in db: %v", err)
	}

	http.Redirect(w, r, "/done", http.StatusSeeOther)
}

func (s Server) deleteAPIPost(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r, w) {
		fmt.Fprintln(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	p, err := parseRValues(r)
	if err != nil {
		log.Fatalf("failed to parse values: %v", err)
	}

	err = s.DB.DeletePost(p.ID)
	if err != nil {
		log.Fatalf("delete post in db: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatalf("failed to encode %v", err)
	}
}

func (s Server) deletePost(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r, w) {
		return
	}

	p, err := parseRValues(r)
	if err != nil {
		log.Fatalf("failed to parse values: %v", err)
	}

	err = s.DB.DeletePost(p.ID)
	if err != nil {
		log.Fatalf("delete post in db: %v", err)
	}

	w.WriteHeader(http.StatusGone)

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatalf("failed to encode %v", err)
	}
}

func (s Server) viewPost(w http.ResponseWriter, r *http.Request) {
	p, err := parseRValues(r)
	if err != nil {
		log.Fatalf("failed to parse values: %v", err)
	}

	p, err = s.DB.ReadPost(int(p.ID))
	if err != nil {
		log.Fatalf("read posts: %v", err)
	}

	p.ParsedDate = parseDate(p.Date)

	err = s.T.ExecuteTemplate(w, "post.html.tmpl", p)

	if err != nil {
		log.Fatalf("execute %v", err)
	}
}

func (s Server) authenticated(r *http.Request, w http.ResponseWriter) bool {
	cookie, err := r.Cookie("authorName")
	if err != nil {
		http.Redirect(w, r, "/fail?reason=cookieDoesntExist", http.StatusSeeOther)
		return false
	}

	encryptedAuthorByte, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		log.Fatalf("failed to decode base64 string to byte: %v", err)
	}

	decryptedAuthorByte, err := security.Decrypt(encryptedAuthorByte, s.Key)
	if err != nil {
		log.Fatalf("failed to decrypt: %v", err)
	}

	authorExists, err := s.DB.AuthorExists(string(decryptedAuthorByte))
	if err != nil {
		log.Fatalf("failed sql author exist check: %v", err)
	}

	if !authorExists {
		http.Redirect(w, r, "/fail?reason=authorDoesntExist", http.StatusUnauthorized)

		return false
	}

	return true
}

func (s Server) editPost(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r, w) {
		return
	}

	p, err := parseRValues(r)
	if err != nil {
		log.Fatalf("failed to parse values: %v", err)
	}

	fmt.Println("editPost print date:", p.Date)

	err = s.DB.UpdatePost(p)
	if err != nil {
		log.Fatalf("edit post in db: %v", err)
	}

	http.Redirect(w, r, "/done", http.StatusSeeOther)
}

func (s Server) editAPIPost(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r, w) {
		return
	}

	p, err := parseRValues(r)
	if err != nil {
		log.Fatalf("failed to parse values: %v", err)
	}

	err = s.DB.UpdatePost(p)
	if err != nil {
		log.Fatalf("edit post in db: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		log.Fatalf("failed to encode %v", err)
	}
}

func (s Server) getLogin(w http.ResponseWriter, _ *http.Request) {
	err := s.T.ExecuteTemplate(w, "login.html.tmpl", nil)
	if err != nil {
		log.Fatalf("execute %v", err)
	}
}

func (s Server) postLogin(w http.ResponseWriter, r *http.Request) {
	authorInput := r.FormValue("author")

	encryptedAuthorByte, err := security.Encrypt([]byte(authorInput), s.Key)
	if err != nil {
		log.Fatal(err)
	}

	cookie := http.Cookie{
		Name:   "authorName",
		Value:  base64.StdEncoding.EncodeToString(encryptedAuthorByte),
		Path:   "/",
		Secure: true,
	}

	authorExists, err := s.DB.AuthorExists(authorInput)
	if err != nil {
		http.Redirect(w, r, "/fail?reason=authorDoesntExist", http.StatusSeeOther)
	}

	if authorExists {
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	if !authorExists {
		http.Redirect(w, r, "/fail?reason=authorDoesntExist", http.StatusSeeOther)
		return
	}
}

func (s Server) getDone(w http.ResponseWriter, _ *http.Request) {
	err := s.T.ExecuteTemplate(w, "done.html.tmpl", nil)
	if err != nil {
		log.Fatalf("execute %v", err)
	}
}

func (s Server) getFail(w http.ResponseWriter, r *http.Request) {
	reason := r.URL.Query().Get("reason")

	err := s.T.ExecuteTemplate(w, "fail.html.tmpl", reason)
	if err != nil {
		log.Fatalf("execute %v", err)
	}
}

func (s Server) SetupMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.getIndex)
	mux.HandleFunc("GET /static/pico.min.css", s.getCSS)
	mux.HandleFunc("GET /api/v0/posts", s.getAPIPosts)
	mux.HandleFunc("POST /api/v0/post", s.postAPIPost)
	mux.HandleFunc("POST /post", s.postPost)
	mux.HandleFunc("DELETE /api/v0/post/{id}", s.deleteAPIPost)
	mux.HandleFunc("DELETE /post/{id}", s.deletePost)
	mux.HandleFunc("GET /post/{id}", s.viewPost)
	mux.HandleFunc("POST /api/v0/post/{id}", s.editAPIPost)
	mux.HandleFunc("POST /post/{id}", s.editPost)
	mux.HandleFunc("GET /login", s.getLogin)
	mux.HandleFunc("POST /api/v0/login", s.postLogin)
	mux.HandleFunc("GET /done", s.getDone)
	mux.HandleFunc("GET /fail", s.getFail)

	return mux
}

func Run(mux *http.ServeMux) {
	log.Fatal(http.ListenAndServe(":8080", mux))
}
