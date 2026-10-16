package main

import (
	"fmt"
	"net/http"
	"strings"

	"vatsapartment-go/db"
)

const siteURL = "https://vatsapartment.com"

var staticSitemapPages = []string{
	"/",
	"/book-now",
	"/room-showcase",
	"/location",
	"/contact-us",
	"/blog",
	"/privacy",
	"/terms",
	"/support",
}

func handleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	for _, p := range staticSitemapPages {
		b.WriteString("  <url><loc>" + siteURL + p + "</loc></url>\n")
	}

	rows, err := db.DB.Query(
		`SELECT slug, TO_CHAR(updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')
		 FROM blog_posts WHERE status = 'published' ORDER BY created_at DESC`)
	if err != nil {
		// Static pages alone are still a valid sitemap; log and serve what we have.
		fmt.Printf("sitemap: blog query failed: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var slug, lastmod string
			if rows.Scan(&slug, &lastmod) == nil {
				b.WriteString("  <url><loc>" + siteURL + "/blog/" + slug + "</loc><lastmod>" + lastmod + "</lastmod></url>\n")
			}
		}
	}

	b.WriteString("</urlset>\n")
	w.Write([]byte(b.String()))
}

func handleRobots(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/robots.txt")
}
