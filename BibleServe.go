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
	RequestString string
	Response      chan string
}

func loadVersebyStr(verse_specification string, mapOfVerses map[string]string) (string, error) {
	verse_specification = strings.Replace(verse_specification, "+", " ", 1)
	if !strings.Contains(verse_specification, ":") {
		err := fmt.Errorf("No colon found, so this isn't a verse.")
		return "", err
	}
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
	reverse_book_lookup := map[string]string{
		"Gen": "Genesis",
		"Exo": "Exodus",
		"Lev": "Leviticus",
		"Num": "Numbers",
		"Deu": "Deuteronomy",
		"Jos": "Joshua",
		//"Jud": "Judges",
		"Rut": "Ruth",
		"1Sa": "1 Samuel",
		"2Sa": "2 Samuel",
		"1Ki": "1 Kings",
		"2Ki": "2 Kings",
		"1Ch": "1 Chronicles",
		"2Ch": "2 Chronicles",
		"Ezr": "Ezra",
		"Neh": "Nehemiah",
		"Est": "Esther",
		"Job": "Job",
		"Psa": "Psalms",
		"Pro": "Proverbs",
		"Ecc": "Ecclesiastes",
		"Son": "Song of Solomon",
		"Isa": "Isaiah",
		"Jer": "Jeremiah",
		"Lam": "Lamentations",
		"Eze": "Ezekiel",
		"Dan": "Daniel",
		"Hos": "Hosea",
		"Joe": "Joel",
		"Amo": "Amos",
		"Oba": "Obadiah",
		"Jon": "Jonah",
		"Mic": "Micah",
		"Nah": "Nahum",
		"Hab": "Habakkuk",
		"Zep": "Zephaniah",
		"Hag": "Haggai",
		"Zec": "Zechariah",
		"Mal": "Malachi",
		"Mat": "Matthew",
		"Mar": "Mark",
		"Luk": "Luke",
		"Joh": "John",
		"Act": "Acts",
		"Rom": "Romans",
		"1Co": "1 Corinthians",
		"2Co": "2 Corinthians",
		"Gal": "Galatians",
		"Eph": "Ephesians",
		//"Phi": "Philippians",
		"Col": "Colossians",
		"1Th": "1 Thessalonians",
		"2Th": "2 Thessalonians",
		"1Ti": "1 Timothy",
		"2Ti": "2 Timothy",
		"Tit": "Titus",
		"Phi": "Philemon",
		"Heb": "Hebrews",
		"Jam": "James",
		"1Pe": "1 Peter",
		"2Pe": "2 Peter",
		"1Jo": "1 John",
		"2Jo": "2 John",
		"3Jo": "3 John",
		"Jud": "Jude",
		"Rev": "Revelation",
	}
	stringResponse := make(chan string)
	completed := make(chan bool)

	// Convert map to slice of keys for easier chunking
	keys := make([]string, 0, len(map_of_verses))
	for key := range map_of_verses {
		keys = append(keys, key)
	}

	chunkSize := (len(keys) + searchInstances - 1) / searchInstances

	// Launch multiple goroutines
	for i := 0; i < searchInstances; i++ {
		go func(start int) {
			end := start + chunkSize
			if end > len(keys) {
				end = len(keys)
			}
			for _, key := range keys[start:end] {
				value := map_of_verses[key]
				if strings.Contains(value, strings.ReplaceAll(searchString, "+", " ")) {
					stringResponse <- reverse_book_lookup[key[0:3]] + key[3:] + " " + value
				}
			}
			completed <- true
		}(i * chunkSize)
	}

	// Collect results
	var result string
	go func() {
		for i := 0; i < searchInstances; i++ {
			<-completed
		}
		close(stringResponse)
	}()

	for res := range stringResponse { //in Go, this is a way of "waiting" for a new item in the channel
		result += "\n\n" + res
	}

	return result
}

func verseHandler(mapOfVerses map[string]string, requests chan VerseRequest) {
	for req := range requests {
		log.Println("String received: ", req.RequestString)
		verse, err := loadVersebyStr(req.RequestString, mapOfVerses)
		if err != nil {
			req.Response <- searchBibleForStr(req.RequestString, mapOfVerses)
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

		RequestString := strings.TrimPrefix(r.URL.Path, "/api/")
		start_time := time.Now()
		responseChan := make(chan string)
		requests <- VerseRequest{RequestString: RequestString, Response: responseChan}
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
