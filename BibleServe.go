// What's left? For learning, include concurrent access to DB
// Spawn off a goroutine to handle web requests and respond accordingly
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

type BibleVerse struct {
	Book        string
	Chapter     uint8
	VerseNumber uint8
	VerseStr    string
}

func loadVersebyStr(verse_specification string, mapOfVerses *map[string]string) (string, error) { //intelligently parse our string input and call loadVersebyBook accordingly
	//lookup verse from database
	//strings.Replace()
	fmt.Println("Verse string received: ", verse_specification)
	book := strings.Split(verse_specification, " ")[0]
	book = book[0:3]
	chapter, verse := func(input string) (string, string) {
		right_side := strings.Split(input, " ")[1]
		return strings.Split(right_side, ":")[0], strings.Split(right_side, ":")[1]
	}(verse_specification)
	chapter_i, _ := strconv.Atoi(chapter)
	verse_i, _ := strconv.Atoi(verse)
	return loadVersebyBook(book, uint8(chapter_i), uint8(verse_i), mapOfVerses)
}

func loadVersebyBook(book string, chapter, VerseNumber uint8, mapOfVerses *map[string]string) (string, error) { //for now this accesses our ballooned struct list in memory (inefficiently). This will EVENTUALLY access a database with o(1) time complexity
	access_string := fmt.Sprint(book, " ", chapter, ":", VerseNumber)
	//fmt.Println("Accessing map by access_string: ", access_string)
	return (*mapOfVerses)[access_string], nil
}

func createMapOfVerses(bible_slice *[]BibleVerse) map[string]string { //this is our access method for now, even though it balloons memory even further. It DOES make search must faster.
	mapOfVerses := make(map[string]string)
	for _, verse := range *bible_slice {
		access_string := fmt.Sprint(verse.Book, " ", verse.Chapter, ":", verse.VerseNumber)
		mapOfVerses[access_string] = verse.VerseStr
	}
	return mapOfVerses
}

func scanBibleFromTxtFile(file_name string) map[string]string { //this is currently big, fat and unabstracted
	var bible []BibleVerse
	//for now this is done in memory, eventually it'll be a db
	bibleTxtFile, err := os.Open(file_name)
	if err != nil {
		log.Fatal(err)
	}
	defer bibleTxtFile.Close()

	scanner := bufio.NewScanner(bibleTxtFile) //using bufio.Scanner because it takes data in auto-sized chunks
	for scanner.Scan() {                      //runtime largely doesn't matter on this, so this is somewhat inefficient
		fmt.Println(scanner.Text())
		colonIndex := strings.Index(scanner.Text(), ":")
		switch { //just including a switch in case I want to expand this logic in the future
		case colonIndex < 8 && colonIndex != -1 && unicode.IsNumber(rune(scanner.Text()[colonIndex-1])) && strings.Count(scanner.Text(), " ") > 1: //this is a new verse; it also only works because of short-circuiting (look up)
			bible = append(bible,
				func(line string) BibleVerse {
					//parse line
					//return BibleVerse
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
			fmt.Println("Verses added: ", len(bible))
		default:
			lastVerse := bible[len(bible)-1]
			lastVerse.VerseStr += " " + scanner.Text()
			bible[len(bible)-1] = lastVerse
		}
	}
	fmt.Println("Total verses: ", len(bible))
	return createMapOfVerses(&bible)
}

// This is a "wrapper" around the handler function which secretly just calls the proper handler function
// Except doing it this way allows for mapOfVerses to be accessed by passing in the reference
// This prevents global variable shenanigens
func handler(mapOfVerses *map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received a new request")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Extract the verse specification from the URL path
		verseSpec := strings.TrimPrefix(r.URL.Path, "/api/")
		//log.Println("Received GET Request: ", body)
		start_time := time.Now()
		verse, err := loadVersebyStr(verseSpec, mapOfVerses)
		if err != nil {
			log.Println("Error loading verse:", err)
			http.Error(w, "Verse not found", http.StatusNotFound)
			return
		}
		log.Printf("Total execution time for this lookup: %d us\n", time.Since(start_time).Microseconds())
		fmt.Fprintf(w, verse)
		log.Println("Handler completed for request.")
	}
}

func main() {
	//GET https.../api/John 3:16 should get the verse
	//just going to https://....com/ should get a basic HTML page describing the API
	mapOfVerses := scanBibleFromTxtFile("ESVBible.txt")
	v, _ := loadVersebyStr("John 3:16", &mapOfVerses)
	fmt.Println("John 3:16 is: ", v)

	http.HandleFunc("/api/", handler(&mapOfVerses)) //handler here is a FUNCTION which'll be used
	log.Println("Starting server on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
