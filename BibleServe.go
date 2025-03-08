package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

type VerseRequest struct {
	RequestString string
	Response      chan string
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

func loadVersebyStr(verse_specification string, mapOfVerses *map[string]string) ([]Verse, error) {
	verse_specification = strings.Replace(verse_specification, "+", " ", 1)
	if !strings.Contains(verse_specification, ":") {
		err := fmt.Errorf("no colon found, input string isn't a verse")
		return []Verse{{VerseName: "", VerseContent: ""}}, err
	}

	verse_specification = strings.Replace(verse_specification, "Judges", "Jdg", 1)
	verse_specification = strings.Replace(verse_specification, "Philemon", "Phm", 1)
	verse_specification = strings.Replace(verse_specification, "Son", "Sol", 1)
	verse_w_book_truncd := (strings.Split(verse_specification, " ")[0])[0:3] + " " + strings.Split(verse_specification, " ")[1]
	if verse_result := mapOfVerses[verse_w_book_truncd]; verse_result != "" {
		return []Verse{verse_result}, nil
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
		var composite_verse strings.Builder
		for i := first_verse; i <= last_verse; i++ {
			v, _ := loadVersebyBook(book, uint8(chapter_i), uint8(i), mapOfVerses)
			composite_verse.WriteString(v)
			composite_verse.WriteByte(' ')
		}
		return composite_verse.String(), nil
	}

	return loadVersebyBook(book, uint8(chapter_i), uint8(verse_i), mapOfVerses)
}

func loadVersebyBook(book string, chapter, VerseNumber uint8, mapOfVerses map[string]string) (string, error) {
	access_string := fmt.Sprint(book, " ", chapter, ":", VerseNumber)
	return mapOfVerses[access_string], nil
}

func searchBibleForStr(searchString string, case_sensitive bool, map_of_verses *map[string]string) []Verse {
	reverse_book_lookup := map[string]string{ //should this be global?
		"Gen": "Genesis",
		"Exo": "Exodus",
		"Lev": "Leviticus",
		"Num": "Numbers",
		"Deu": "Deuteronomy",
		"Jos": "Joshua",
		"Jdg": "Judges",
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
		"Sol": "Song of Solomon",
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
		"Phi": "Philippians",
		"Col": "Colossians",
		"1Th": "1 Thessalonians",
		"2Th": "2 Thessalonians",
		"1Ti": "1 Timothy",
		"2Ti": "2 Timothy",
		"Tit": "Titus",
		"Phm": "Philemon",
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
	verseResponses := make(chan Verse, 100)
	switch {
	case case_sensitive:
		for key, value := range *map_of_verses {
			key = strings.Replace(key, "Judges", "Jdg", 1)
			key = strings.Replace(key, "Philemon", "Phm", 1)
			key = strings.Replace(key, "Son", "Sol", 1)
			if strings.Contains(value, searchString) {
				verseResponses <- Verse{
					VerseName:    reverse_book_lookup[key[0:3]] + key[3:],
					VerseContent: value,
				}
			}
		}
		break
	case !case_sensitive:
		for key, value := range *map_of_verses {
			key = strings.Replace(key, "Judges", "Jdg", 1)
			key = strings.Replace(key, "Philemon", "Phm", 1)
			key = strings.Replace(key, "Son", "Sol", 1)
			if strings.Contains(strings.ToLower(value), strings.ToLower(searchString)) {
				verseResponses <- Verse{
					VerseName:    reverse_book_lookup[key[0:3]] + key[3:],
					VerseContent: value,
				}
			}
		}
	}

	// Collect results
	var result []Verse

	for res := range verseResponses { //in Go, this is a way of "waiting" for a new item in the channel
		result = append(result, res)
	}

	return result
}

type SearchRequest struct {
	SearchMode     string   `json:"search_mode"`
	SearchTerms    []string `json:"search_terms"`
	GeneralOptions struct {
		CaseSensitive bool `json:"case_sensitive"`
		TrimSpaces    bool `json:"trim_spaces"`
		//ordered?
		//old/new testament?
	} `json:"general_options"`
	ModeOptions struct {
		PlainSearchOptions struct {
			CaseSensitive bool `json:"case_sensitive"`
		} `json:"plain_search_options"`
		AssocSearchOptions struct {
			Radius int `json:"radius"`
		} `json:"assoc_options"`
	} `json:"mode_options"`
}

type Verse struct {
	VerseName    string `json:"verse_name"`
	VerseContent string `json:"verse_content"`
}

func encodeToJSON(response []Verse) string { //returns JSON to be sent to the user
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Error encoding to JSON:", err)
		return ""
	}

	return string(jsonResponse)
}

func decodeFromJSON(data []byte) (SearchRequest, error) {
	var request SearchRequest
	err := json.Unmarshal(data, &request)
	if err != nil {
		return SearchRequest{}, err
	}
	return request, nil
}

func handler(map_of_verses *map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received a new request")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		start_time := time.Now()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}

		fullRequest, err := decodeFromJSON(body)
		if err != nil {
			http.Error(w, "Unable to decode JSON", http.StatusBadRequest)
			return
		}

		requestType := fullRequest.SearchMode
		var response []Verse
		//I could use a channel here, but don't see any explicit benefit for that
		switch {
		case requestType == "plain_search": //simple string searching
			response = searchBibleForStr(fullRequest.SearchTerms[0], fullRequest.GeneralOptions.CaseSensitive, map_of_verses)
			return
		case requestType == "specified_search": //specifying verse to search for
			response, err = loadVersebyStr(fullRequest.SearchTerms[0], map_of_verses)
			return
		case requestType == "assoc_search":
			//this is where we access the radius and multiple search terms
			return
		default:
			// do nothing
		}

		json_response := encodeToJSON(response)
		elapsed := time.Since(start_time).Milliseconds()
		//log.Println(verse)
		log.Printf("Total execution time for lookup: %d ms - Completed.\n", elapsed)
		fmt.Fprint(w, json_response)
	}
}

func main() {
	mapOfVerses := scanBibleFromTxtFile("bible/ESVBible.txt")

	http.HandleFunc("/api/", handler(&mapOfVerses))
	http.HandleFunc()

	log.Println("Starting server on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
