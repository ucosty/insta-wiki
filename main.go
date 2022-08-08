//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=templates
package main

import (
	"bytes"
	"database/sql"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/ucosty/insta-wiki/models"
	"github.com/ucosty/insta-wiki/templates"
	"github.com/yuin/goldmark"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

const GetPage = "SELECT pages.id, pages.title, pages.revision_id, r.body FROM pages INNER JOIN revisions r on r.id = pages.revision_id WHERE path = ? LIMIT 1"
const CreatePage = "INSERT INTO pages (user_id, path, title, created_at, updated_at) VALUES (0, ?, ?, date('now'), date('now'))"
const CreateRevision = "INSERT INTO revisions (page_id, user_id, created_at, title, body) VALUES (?, 0, date('now'), ?, ?)"
const SearchQuery = "SELECT pages.id, pages.path, pages.title, pages.revision_id, r.body FROM pages INNER JOIN revisions r on r.id = pages.revision_id"
const UpdatePageRevision = "UPDATE pages SET revision_id = ?, title = ? WHERE id = ?"
const GetLastInsertedId = "select last_insert_rowid()"

type Database struct {
	DB                  *sql.DB
	GetPageQuery        *sql.Stmt
	CreateRevisionQuery *sql.Stmt
	CreatePageQuery     *sql.Stmt
	SearchQuery         *sql.Stmt
	UpdatePageRevision  *sql.Stmt
	GetLastInsertedId   *sql.Stmt
}

func readConfiguration() models.Configuration {
	var config models.Configuration
	port := flag.Int("port", 8080, "http server port")
	flag.Parse()

	databaseFile := flag.String("database", "wiki.db", "database file")

	config.ListenAddress = fmt.Sprintf(":%d", *port)
	config.DatabaseFile = *databaseFile

	return config
}

func getContentType(filename string, file []byte) string {
	contentType := http.DetectContentType(file)
	extension := filepath.Ext(filename)

	if strings.HasPrefix(contentType, "text/plain") && extension == ".css" {
		return "text/css"
	}

	return contentType
}

func staticFilesHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	filename := fmt.Sprintf("static/%s", params.ByName("file"))
	file, err := staticFiles.ReadFile(filename)

	contentType := getContentType(filename, file)
	w.Header().Add("Content-Type", contentType)

	if _, err = w.Write(file); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func getPageByPath(database *Database, path string, renderMarkdown bool) (models.Page, error) {
	result := database.GetPageQuery.QueryRow(path)

	page := models.Page{IsNewPage: false, Path: path}
	if err := result.Scan(&page.Id, &page.Title, &page.Revision, &page.BodyMarkdown); err != nil {
		return page, err
	}

	if renderMarkdown {
		//page.BodyFormatted = string(markdown.ToHTML([]byte(page.BodyMarkdown), nil, nil))
		var buffer bytes.Buffer
		if err := goldmark.Convert([]byte(page.BodyMarkdown), &buffer); err != nil {
			panic(err)
		}
		page.BodyFormatted = buffer.String()
	}

	return page, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "/wiki/home", http.StatusSeeOther)
}

func showPageHandler(w http.ResponseWriter, r *http.Request, page *models.Page) {
	contentType := r.Header.Get("accept")
	if contentType == "application/json" {
		output, err := json.Marshal(page)
		if err != nil {
			http.Error(w, "Failed to serialise page into JSON", 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(output)
		return
	}

	fmt.Fprintf(w, templates.Index(page))
}

func editPageHandler(database *Database) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		path := params.ByName("page")
		page, err := getPageByPath(database, path, true)
		if err != nil && err == sql.ErrNoRows {
			http.Error(w, "Page not found", 404)
			return
		}

		fmt.Fprintf(w, templates.Edit(&page))
	}
}

func newPageHandler(w http.ResponseWriter, params httprouter.Params) {
	path := params.ByName("page")
	page := models.Page{IsNewPage: true, Path: path}
	fmt.Fprintf(w, templates.Edit(&page))
}

func historyPageHandler(database *Database) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		path := params.ByName("page")
		page, err := getPageByPath(database, path, true)
		if err != nil && err == sql.ErrNoRows {
			http.Error(w, "Page not found", 404)
			return
		}

		fmt.Fprintf(w, templates.History(&page, []models.Page{}))
	}
}

func errorHandler(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), 500)
}

func getPageHandler(database *Database) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		path := params.ByName("page")
		page, err := getPageByPath(database, path, true)

		// If the page does not exist, we're going to always show the new page form
		if err != nil && err == sql.ErrNoRows {
			newPageHandler(w, params)
			return
		}

		// Any other error
		if err != nil {
			errorHandler(w, err)
			return
		}

		showPageHandler(w, r, &page)
	}
}

func createPageHandler(database *Database, w http.ResponseWriter, r *http.Request) {
	title := r.Form.Get("title")
	body := r.Form.Get("body")

	path := r.Form.Get("path")
	if path == "" {
		path = slug.Make(title)
	}

	redirectUrl := fmt.Sprintf("/wiki/%s", path)

	database.CreatePageQuery.Exec(path, title)
	result := database.GetLastInsertedId.QueryRow()
	var pageId int
	result.Scan(&pageId)

	database.CreateRevisionQuery.Exec(pageId, title, body)
	result = database.GetLastInsertedId.QueryRow()
	var revisionId int
	result.Scan(&revisionId)

	database.UpdatePageRevision.Exec(revisionId, title, pageId)

	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func updatePageHandler(database *Database, w http.ResponseWriter, r *http.Request, page *models.Page) {
	title := r.Form.Get("title")
	body := r.Form.Get("body")
	path := r.Form.Get("path")
	if path == "" {
		path = slug.Make(title)
	}

	fmt.Print("Updating page with path ", path, " and title ", title)

	redirectUrl := fmt.Sprintf("/wiki/%s", path)

	database.CreateRevisionQuery.Exec(page.Id, title, body)
	result := database.GetLastInsertedId.QueryRow()
	var revisionId int
	result.Scan(&revisionId)

	database.UpdatePageRevision.Exec(revisionId, title, page.Id)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func postPageHandler(database *Database) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if err := r.ParseForm(); err != nil {
			return
		}

		path := params.ByName("page")
		page, err := getPageByPath(database, path, false)

		// If the page does not exist, we're going to always show the new page form
		if err != nil && err == sql.ErrNoRows {
			createPageHandler(database, w, r)
			return
		}

		updatePageHandler(database, w, r, &page)
	}
}

func searchHandler(database *Database) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		result, err := database.SearchQuery.Query()
		if err != nil {
			return
		}

		defer result.Close()

		var pages []models.Page

		for result.Next() {
			var page models.Page
			if err := result.Scan(&page.Id, &page.Path, &page.Title, &page.Revision, &page.BodyMarkdown); err != nil {
				return
			}

			pages = append(pages, page)
		}

		fmt.Fprintf(w, templates.Search("", 0, pages))
	}
}

func setupPreparedStatements(db *sql.DB) (database *Database, err error) {
	database = &Database{}
	if database.CreateRevisionQuery, err = db.Prepare(CreateRevision); err != nil {
		return nil, err
	}
	if database.CreatePageQuery, err = db.Prepare(CreatePage); err != nil {
		return nil, err
	}
	if database.UpdatePageRevision, err = db.Prepare(UpdatePageRevision); err != nil {
		return nil, err
	}
	if database.SearchQuery, err = db.Prepare(SearchQuery); err != nil {
		return nil, err
	}
	if database.GetPageQuery, err = db.Prepare(GetPage); err != nil {
		return nil, err
	}
	if database.GetLastInsertedId, err = db.Prepare(GetLastInsertedId); err != nil {
		return nil, err
	}
	return database, nil
}

func main() {
	config := readConfiguration()

	goose.SetBaseFS(embeddedMigrations)

	router := httprouter.New()

	connectionString := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL", config.DatabaseFile)
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}

	database, err := setupPreparedStatements(db)
	if err != nil {
		log.Fatal(err)
	}

	router.GET("/", indexHandler)
	router.GET("/search", searchHandler(database))
	router.GET("/static/:file", staticFilesHandler)
	router.GET("/wiki/:page", getPageHandler(database))
	router.GET("/wiki/:page/edit", editPageHandler(database))
	router.GET("/wiki/:page/history", historyPageHandler(database))
	router.POST("/wiki/:page", postPageHandler(database))

	log.Printf("Insta-Wiki listening on %s\n", config.ListenAddress)
	if err := http.ListenAndServe(config.ListenAddress, router); err != nil {
		log.Fatal(err)
	}
}
