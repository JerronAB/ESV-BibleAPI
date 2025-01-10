package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	searchInstances = 10
)

type BibleVerse struct {
	Book        string
	Chapter     uint8
	VerseNumber uint8
	VerseStr    string
}

type VerseRequest struct {
	VerseSpec string
	Response  chan string
}

func loadVersebyStr(verse_specification string, mapOfVerses map[string]string) (string, error) {
	log.Println("Verse string received: ", verse_specification)
	verse_specification = strings.Replace(verse_specification, "+", " ", 1)
	verse_w_book_truncd := (strings.Split(verse_specification, " ")[0])[0:3] + " " + strings.Split(verse_specification, " ")[1]
	if verse_result := mapOfVerses[verse_w_book_truncd]; verse_result != "" {
		return verse_result, nil
	}
	log.Println("Initial string lookup failed for: ", verse_w_book_truncd)
	book := strings.Split(verse_w_book_truncd, " ")[0]
	chapter, verse := func(input string) (string, string) {
		right_side := strings.Split(input, " ")[1]
		return strings.Split(right_side, ":")[0], strings.Split(right_side, ":")[1]
	}(verse_specification)
	chapter_i, _ := strconv.Atoi(chapter)
	verse_i, err := strconv.Atoi(verse)
	if err != nil {
		first_verse, _ := strconv.Atoi(strings.Split(verse, "-")[0])
		last_verse, _ := strconv.Atoi(strings.Split(verse, "-")[1])
		composite_verse := ""
		for i := first_verse; i <= last_verse; i++ {
			v, _ := loadVersebyBook(book, uint8(chapter_i), uint8(i), mapOfVerses)
			composite_verse += v + " "
		}
		return composite_verse, nil
	}
	return loadVersebyBook(book, uint8(chapter_i), uint8(verse_i), mapOfVerses)
}

func loadVersebyBook(book string, chapter, VerseNumber uint8, mapOfVerses map[string]string) (string, error) {
	access_string := fmt.Sprint(book, " ", chapter, ":", VerseNumber)
	return mapOfVerses[access_string], nil
}

func createMapOfVerses(bible_slice *[]BibleVerse) map[string]string {
	mapOfVerses := make(map[string]string)
	for _, verse := range *bible_slice {
		access_string := fmt.Sprint(verse.Book, " ", verse.Chapter, ":", verse.VerseNumber)
		mapOfVerses[access_string] = verse.VerseStr
	}
	log.Println("Verse map has been processed.")
	return mapOfVerses
}

func scanBibleFromTxtFile(file_name string) map[string]string {
	var bible []BibleVerse
	bibleTxtFile, err := os.Open(file_name)
	if err != nil {
		log.Fatal(err)
	}
	defer bibleTxtFile.Close()

	scanner := bufio.NewScanner(bibleTxtFile)
	for scanner.Scan() {
		colonIndex := strings.Index(scanner.Text(), ":")
		switch {
		case colonIndex < 8 && colonIndex != -1 && unicode.IsNumber(rune(scanner.Text()[colonIndex-1])) && strings.Count(scanner.Text(), " ") > 1:
			bible = append(bible,
				func(line string) BibleVerse {
					book_chp_verse_txt := strings.SplitN(line, ":", 2)
					book_chp := book_chp_verse_txt[0]
					verse_txt := book_chp_verse_txt[1]
					book_chptr := strings.Split(book_chp, " ")
					verse_verseLine := strings.SplitN(verse_txt, " ", 2)
					chptr, err := strconv.ParseUint(book_chptr[1], 10, 8)
					if err != nil {
						fmt.Println("Error:", err)
					}
					verse, err := strconv.ParseUint(verse_verseLine[0], 10, 8)
					if err != nil {
						fmt.Println("Error:", err)
					}
					return BibleVerse{Book: book_chptr[0], Chapter: uint8(chptr), VerseNumber: uint8(verse), VerseStr: verse_verseLine[1]}
				}(scanner.Text()))
		default:
			lastVerse := bible[len(bible)-1]
			lastVerse.VerseStr += " " + scanner.Text()
			bible[len(bible)-1] = lastVerse
		}
	}
	fmt.Println("Total verses: ", len(bible))
	return createMapOfVerses(&bible)
}

func searchBibleForStr(searchString string, map_of_verses map[string]string) string {
	stringResponse := make(chan string)
	for i := 0; i < searchInstances; i++ {
		go func() {

		}()
	}
}

func verseHandler(mapOfVerses map[string]string, requests chan VerseRequest) {
	for req := range requests {
		verse, err := loadVersebyStr(req.VerseSpec, mapOfVerses)
		if err != nil {
			req.Response <- searchBibleForStr(req.VerseSpec, mapOfVerses)
		} else {
			req.Response <- verse
		}
	}
}

func handler(requests chan VerseRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received a new request")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		verseSpec := strings.TrimPrefix(r.URL.Path, "/api/")
		start_time := time.Now()
		responseChan := make(chan string)
		requests <- VerseRequest{VerseSpec: verseSpec, Response: responseChan}
		verse := <-responseChan
		elapsed := time.Since(start_time).Milliseconds()
		log.Println(verse)
		log.Printf("Total execution time for this lookup: %d ms\n", elapsed)
		fmt.Fprintf(w, verse)
		log.Println("Handler completed for request.")
	}
}

func main() {
	mapOfVerses := scanBibleFromTxtFile("ESVBible.txt")
	requests := make(chan VerseRequest)
	go verseHandler(mapOfVerses, requests)

	http.HandleFunc("/api/", handler(requests))
	log.Println("Starting server on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
