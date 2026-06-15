package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"vatsapartment-go/db"
)

type BlogPost struct {
	ID        string
	Title     string
	Slug      string
	Excerpt   string
	Content   string
	ImageURL  string
	Author    string
	Status    string
	CreatedAt string
	UpdatedAt string
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	var role string
	db.DB.QueryRow("SELECT role FROM users WHERE username = $1", cookie.Value).Scan(&role)
	if role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	return true
}

func handleBlog(w http.ResponseWriter, r *http.Request) {
	posts, err := getPublishedBlogs()
	if err != nil {
		http.Error(w, "Failed to load blog posts", 500)
		return
	}
	render(w, "blog.html", map[string]interface{}{"Posts": posts})
}

func handleBlogPost(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	var post BlogPost
	err := db.DB.QueryRow(
		`SELECT id, title, slug, COALESCE(excerpt,''), content, COALESCE(image_url,''), 
		 COALESCE(author,'Vats Apartment'), status, 
		 TO_CHAR(created_at, 'Mon DD, YYYY'), TO_CHAR(updated_at, 'Mon DD, YYYY')
		 FROM blog_posts WHERE slug = $1 AND status = 'published'`,
		slug,
	).Scan(&post.ID, &post.Title, &post.Slug, &post.Excerpt, &post.Content,
		&post.ImageURL, &post.Author, &post.Status, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	render(w, "blog_post.html", map[string]interface{}{"Post": post})
}

func handleAdminBlogList(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	rows, err := db.DB.Query(
		`SELECT id, title, slug, COALESCE(excerpt,''), COALESCE(image_url,''), status, 
		 TO_CHAR(created_at, 'Mon DD, YYYY')
		 FROM blog_posts ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, "Failed to load blog posts", 500)
		return
	}
	defer rows.Close()

	var posts []BlogPost
	for rows.Next() {
		var p BlogPost
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.ImageURL, &p.Status, &p.CreatedAt)
		posts = append(posts, p)
	}
	if posts == nil {
		posts = []BlogPost{}
	}

	renderPrivate(w, "admin_blog_list.html", map[string]interface{}{"Posts": posts, "Active": "blog", "Title": "Blog Posts"})
}

func handleAdminBlogForm(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	data := map[string]interface{}{"FormTitle": "New Blog Post"}

	editID := r.URL.Query().Get("id")
	if editID != "" {
		var post BlogPost
		err := db.DB.QueryRow(
			`SELECT id, title, slug, COALESCE(excerpt,''), content, COALESCE(image_url,''), status
			 FROM blog_posts WHERE id = $1`, editID,
		).Scan(&post.ID, &post.Title, &post.Slug, &post.Excerpt, &post.Content, &post.ImageURL, &post.Status)
		if err == nil {
			data["FormTitle"] = "Edit Blog Post"
			data["Post"] = post
		}
	}

	data["Active"] = "blog"
	data["Title"] = data["FormTitle"]
	renderPrivate(w, "admin_blog_form.html", data)
}

func handleAdminBlogSave(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	r.ParseForm()
	id := r.FormValue("id")
	title := r.FormValue("title")
	slug := r.FormValue("slug")
	content := r.FormValue("content")
	excerpt := r.FormValue("excerpt")
	imageURL := r.FormValue("image_url")
	status := r.FormValue("status")
	action := r.FormValue("action")

	if action == "delete" && id != "" {
		db.DB.Exec("DELETE FROM blog_posts WHERE id = $1", id)
		log.Printf("Blog post deleted: %s", id)
		http.Redirect(w, r, "/admin/blog", http.StatusSeeOther)
		return
	}

	if title == "" || content == "" {
		renderPrivate(w, "admin_blog_form.html", map[string]interface{}{
			"FormTitle": "New Blog Post",
			"Error":     "Title and content are required",
			"Active":    "blog",
			"Title":     "New Blog Post",
		})
		return
	}

	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	if id != "" {
		_, err := db.DB.Exec(
			`UPDATE blog_posts SET title=$1, slug=$2, excerpt=$3, content=$4, image_url=$5, status=$6, updated_at=$7 WHERE id=$8`,
			title, slug, excerpt, content, imageURL, status, now, id,
		)
		if err != nil {
			log.Printf("Blog update error: %v", err)
			renderPrivate(w, "admin_blog_form.html", map[string]interface{}{
				"FormTitle":    "Edit Blog Post",
				"Error":        "Failed to save: " + err.Error(),
				"Active":       "blog",
				"Title":        "Edit Blog Post",
			})
			return
		}
	} else {
		id = fmt.Sprintf("blog_%d", time.Now().UnixNano())
		_, err := db.DB.Exec(
			`INSERT INTO blog_posts (id, title, slug, excerpt, content, image_url, status, created_at, updated_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`,
			id, title, slug, excerpt, content, imageURL, status, now,
		)
		if err != nil {
			log.Printf("Blog insert error: %v", err)
			renderPrivate(w, "admin_blog_form.html", map[string]interface{}{
				"FormTitle":    "New Blog Post",
				"Error":        "Failed to create: " + err.Error(),
				"Active":       "blog",
				"Title":        "New Blog Post",
			})
			return
		}
	}

	log.Printf("Blog post saved: %s (%s)", title, id)
	http.Redirect(w, r, "/admin/blog", http.StatusSeeOther)
}

func getPublishedBlogs() ([]BlogPost, error) {
	rows, err := db.DB.Query(
		`SELECT id, title, slug, COALESCE(excerpt,''), COALESCE(image_url,''), 
		 COALESCE(author,'Vats Apartment'), TO_CHAR(created_at, 'Mon DD, YYYY')
		 FROM blog_posts WHERE status = 'published' ORDER BY created_at DESC LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []BlogPost
	for rows.Next() {
		var p BlogPost
		rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.ImageURL, &p.Author, &p.CreatedAt)
		posts = append(posts, p)
	}
	if posts == nil {
		posts = []BlogPost{}
	}
	return posts, nil
}

func blogHTML(content string) template.HTML {
	return template.HTML(content)
}

func handleBlogPreview(w http.ResponseWriter, r *http.Request) {
	jsonContent(w)
	posts, err := getPublishedBlogs()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"posts": []BlogPost{}})
		return
	}
	if len(posts) > 3 {
		posts = posts[:3]
	}
	type Preview struct {
		Title     string `json:"title"`
		Slug      string `json:"slug"`
		Excerpt   string `json:"excerpt"`
		CreatedAt string `json:"created_at"`
	}
	var previews []Preview
	for _, p := range posts {
		previews = append(previews, Preview{
			Title:     p.Title,
			Slug:      p.Slug,
			Excerpt:   p.Excerpt,
			CreatedAt: p.CreatedAt})
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"posts": previews})
}

var validAPIKey string

func setAPIKey(key string) {
	validAPIKey = key
}

func checkAPIKey(r *http.Request) bool {
	key := r.Header.Get("X-API-Key")
	if key == "" {
		key = r.URL.Query().Get("api_key")
	}
	if key == "" {
		return false
	}
	var exists int
	db.DB.QueryRow("SELECT 1 FROM api_keys WHERE key = $1", key).Scan(&exists)
	if exists == 1 {
		db.DB.Exec("UPDATE api_keys SET last_used_at = NOW() WHERE key = $1", key)
	}
	return exists == 1
}

func handleAPIBlogCreate(w http.ResponseWriter, r *http.Request) {
	jsonContent(w)

	if !checkAPIKey(r) {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid or missing API key"})
		return
	}

	var body struct {
		Title    string `json:"title"`
		Slug     string `json:"slug"`
		Content  string `json:"content"`
		Excerpt  string `json:"excerpt"`
		ImageURL string `json:"image_url"`
		Status   string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON body"})
		return
	}

	if body.Title == "" || body.Content == "" {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "title and content are required"})
		return
	}

	if body.Slug == "" {
		body.Slug = strings.ToLower(strings.ReplaceAll(body.Title, " ", "-"))
	}
	if body.Status == "" {
		body.Status = "draft"
	}

	id := fmt.Sprintf("blog_%d", time.Now().UnixNano())
	now := time.Now().Format("2006-01-02 15:04:05")

	_, err := db.DB.Exec(
		`INSERT INTO blog_posts (id, title, slug, excerpt, content, image_url, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`,
		id, body.Title, body.Slug, body.Excerpt, body.Content, body.ImageURL, body.Status, now,
	)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to create post: " + err.Error()})
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
		"slug":    body.Slug})
}

func handleAdminAPIKeys(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		action := r.FormValue("action")
		keyID := r.FormValue("id")

		if action == "delete" && keyID != "" {
			db.DB.Exec("DELETE FROM api_keys WHERE id = $1", keyID)
			http.Redirect(w, r, "/admin/api-keys", http.StatusSeeOther)
			return
		}

		if action == "create" {
			name := r.FormValue("name")
			if name == "" {
				name = "API Key"
			}
			newKey := generateAPIKey()
			id := fmt.Sprintf("apikey_%d", time.Now().UnixNano())
			db.DB.Exec("INSERT INTO api_keys (id, key, name) VALUES ($1, $2, $3)", id, newKey, name)
			http.Redirect(w, r, "/admin/api-keys?new="+newKey, http.StatusSeeOther)
			return
		}
	}

	rows, err := db.DB.Query(
		`SELECT id, key, name, COALESCE(permissions,'blog'), 
		 TO_CHAR(created_at, 'Mon DD, YYYY'), COALESCE(TO_CHAR(last_used_at, 'Mon DD, YYYY HH24:MI'), 'Never')
		 FROM api_keys ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, "Failed to load API keys", 500)
		return
	}
	defer rows.Close()

	type APIKey struct {
		ID, Key, Name, Permissions, CreatedAt, LastUsed, KeyPreview string
	}
	var keys []APIKey
	for rows.Next() {
		var k APIKey
		rows.Scan(&k.ID, &k.Key, &k.Name, &k.Permissions, &k.CreatedAt, &k.LastUsed)
		if len(k.Key) > 12 {
			k.KeyPreview = k.Key[:8] + "..." + k.Key[len(k.Key)-4:]
		} else {
			k.KeyPreview = k.Key
		}
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []APIKey{}
	}

	renderPrivate(w, "admin_api_keys.html", map[string]interface{}{
		"Keys":   keys,
		"Active": "api_keys",
		"Title":  "API Keys",
		"NewKey": r.URL.Query().Get("new"),
	})
}

func generateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "vats_" + hex.EncodeToString(b)
}
