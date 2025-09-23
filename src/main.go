package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Translation struct {
	Identifier   string `json:"identifier"`
	Name         string `json:"name"`
	Language     string `json:"language"`
	LanguageCode string `json:"langauge_code"`
	License      string `json:"license"`
}

type Book struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type BookInfo struct {
	Translation Translation `json:"translation"`
	Books       []Book      `json:"books"`
}

type Chapter struct {
	BookID  string `json:"book_id"`
	Book    string `json:"book"`
	Chapter int    `json:"chapter"`
	URL     string `json:"url"`
}

type ChapterInfo struct {
	Translation Translation `json:"translation"`
	Chapters    []Chapter   `json:"chapters"`
}

type Verse struct {
	BookID   string `json:"book_id"`
	BookName string `json:"book_name"`
	Chapter  int8   `json:"chapter"`
	Verse    int8   `json:"verse"`
	Text     string `json:"text"`
}

type VerseInfo struct {
	Translation Translation `json:"translation"`
	Verses      []Verse     `json:"verses"`
}

type Response struct {
	Reference       string  `json:"reference"`
	Verses          []Verse `json:"verses"`
	Text            string  `json:"text"`
	TranslationID   string  `json:"translation_id"`
	TranslationName string  `json:"translation_name"`
	TranslationNote string  `json:"translation_note"`
}

func APIResponse(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid request")
	}
	return resp, err
}

func GetBookInfo(book_info *BookInfo) error {
	resp, err := APIResponse("https://bible-api.com/data/web")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&book_info)
	if err != nil {
		return err
	}
	return nil
}

func GetChapterInfo(book string, chapter_info *ChapterInfo) error {
	url := fmt.Sprintf("https://bible-api.com/data/web/%s", book)
	resp, err := APIResponse(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&chapter_info)
	if err != nil {
		return err
	}
	return nil
}

func GetVerseInfo(book string, chapter string, verse_info *VerseInfo) error {
	url := fmt.Sprintf("https://bible-api.com/data/asv/%s/%v", book, chapter)
	resp, err := APIResponse(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&verse_info)
	if err != nil {
		return err
	}
	return nil
}

func HtmlStart(w http.ResponseWriter, title string) {
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>%s</title>
	</head>
	<body>
	`, title)
}

func HtmlEnd(w http.ResponseWriter) {
	fmt.Fprint(w, `
	</body>
	</html>
	`)
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	var book_info BookInfo
	err := GetBookInfo(&book_info)
	if err != nil {
		http.NotFound(w, r)
		fmt.Println(err)
		return
	}
	HtmlStart(w, "ASV Bible")
	for _, book := range book_info.Books {
		io.WriteString(w, fmt.Sprintf("<a href=\"%s/%s\">%s</a> <br>", r.URL.Host, strings.ReplaceAll(strings.ToLower(book.Name), " ", ""), book.Name))
	}
	HtmlEnd(w)
	// show all books
}

func getChapters(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	book_name := vars["book"]
	var book_info BookInfo
	err := GetBookInfo(&book_info)
	if err != nil {
		http.NotFound(w, r)
		fmt.Println(err)
		return
	}

	var book_code string
	for _, book := range book_info.Books {
		//fmt.Println(strings.ReplaceAll(strings.ToLower(book.Name), " ", ""))
		if strings.Compare(strings.ReplaceAll(strings.ToLower(book.Name), " ", ""), book_name) == 0 {
			book_code = book.ID
		}
	}

	var chapter_info ChapterInfo
	err = GetChapterInfo(book_code, &chapter_info)
	if err != nil {
		http.NotFound(w, r)
		fmt.Println(err)
		return
	}
	HtmlStart(w, chapter_info.Chapters[0].Book)
	for _, chapter := range chapter_info.Chapters {
		io.WriteString(w, fmt.Sprintf("<a href=\"%s/%v\">%v</a> <br>", r.URL.Path, chapter.Chapter, chapter.Chapter))
	}
	HtmlEnd(w)
	// only show chapters
}

func getVerses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	book_name := vars["book"]
	chapter := vars["chapter"]

	var book_info BookInfo
	err := GetBookInfo(&book_info)
	if err != nil {
		http.NotFound(w, r)
		fmt.Println(err)
		return
	}

	var book_code string
	for _, book := range book_info.Books {

		if strings.Compare(strings.ReplaceAll(strings.ToLower(book.Name), " ", ""), book_name) == 0 {

			book_code = book.ID
		}
	}

	var verse_info VerseInfo
	err = GetVerseInfo(book_code, chapter, &verse_info)
	if err != nil {
		http.NotFound(w, r)
		fmt.Println(err)
		return
	}
	HtmlStart(w, chapter)
	for _, verse := range verse_info.Verses {
		io.WriteString(w, fmt.Sprintf("%v%s%s%s", verse.Verse, " : ", verse.Text, "<br>"))
	}
	HtmlEnd(w)
	// show verses and values
}

func getPassage(w http.ResponseWriter, r *http.Request) {
	// later, plan if i need this rn
}

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/", getBooks)
	m.HandleFunc("/{book}", getChapters)
	m.HandleFunc("/{book}/{chapter}", getVerses)
	// later
	m.HandleFunc("/{book}{chapter}/{verses}", getPassage)

	err := http.ListenAndServe(":3000", m)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else {
		log.Fatal(err)
	}
}
