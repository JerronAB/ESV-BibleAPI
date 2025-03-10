package main

import (
	"bufio"
	"fmt"
	"log"
	"maps"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

var searchInstances = runtime.NumCPU()

type BibleVerse struct {
	Book        string
	Chapter     uint8
	VerseNumber uint8
	VerseStr    string
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

func searchBibleForStr(searchString string, map_of_verses map[string]string, delimiter string, caseSensitive bool) string {
	reverse_book_lookup := map[string]string{
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

	//NOTE: Evaluate buffer size later
	stringResponses := make(chan string, 10)
	var wg sync.WaitGroup
	// This function "populates" verseSearch Channel with verses to search
	verses := maps.Keys(map_of_verses)

	for i := 0; i < searchInstances; i++ {
		wg.Add(1)
		if caseSensitive {
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				for key := range verses {
					value := map_of_verses[key]
					if strings.Contains(value, searchString) {
						key = strings.Replace(key, "Judges", "Jdg", 1)
						key = strings.Replace(key, "Philemon", "Phm", 1)
						key = strings.Replace(key, "Son", "Sol", 1)
						stringResponses <- reverse_book_lookup[key[0:3]] + key[3:] + " - " + value
					}
				}
			}(&wg) //pass in REFERENCE to our WaitGroup
		} else {
			searchString = strings.ToLower(searchString)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				for key := range verses {
					value := strings.ToLower(map_of_verses[key])
					if strings.Contains(value, searchString) {
						key = strings.Replace(key, "Judges", "Jdg", 1)
						key = strings.Replace(key, "Philemon", "Phm", 1)
						key = strings.Replace(key, "Son", "Sol", 1)
						stringResponses <- reverse_book_lookup[key[0:3]] + key[3:] + " - " + value
					}
				}
			}(&wg) //pass in REFERENCE to our WaitGroup
		}
	}

	//close the string channel once waitgroup is done
	go func() {
		wg.Wait()
		close(stringResponses)
	}()

	// Collect results
	var result strings.Builder
	for res := range stringResponses {
		result.WriteString(delimiter)
		result.WriteString(res)
	}

	return result.String()
}

func loadVersebyStr(verse_specification string, mapOfVerses map[string]string) (string, error) {
	//verse_specification = strings.Replace(verse_specification, "+", " ", 1)
	if !strings.Contains(verse_specification, ":") {
		err := fmt.Errorf("no colon found, input string probably isn't a verse")
		return "", err
	}
	//try to use the raw specification first
	if verse_result := mapOfVerses[verse_specification]; verse_result != "" {
		return verse_result, nil
	}

	//if that fails, assume the request has a full book name and try to find info that way
	verse_specification = strings.Replace(verse_specification, "Judges", "Jdg", 1)
	verse_specification = strings.Replace(verse_specification, "Philemon", "Phm", 1)
	verse_specification = strings.Replace(verse_specification, "Son", "Sol", 1)
	verse_w_book_truncd := (strings.Split(verse_specification, " ")[0])[0:3] + " " + strings.Split(verse_specification, " ")[1]
	if verse_result := mapOfVerses[verse_w_book_truncd]; verse_result != "" {
		return verse_result, nil
	}

	//and if THAT fails, assume the request has a full book name and a range of verses
	//the following messy stanza just splits up a string with a verse, into its components (book, chapter, verse)
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

func requestHandler(map_of_verses *map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received a new request: ")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		start_time := time.Now()
		fullRequestEncoded := strings.TrimPrefix(r.URL.String(), "/api")
		fullRequestDecoded, err := url.QueryUnescape(fullRequestEncoded)
		if err != nil {
			fmt.Println("Error decoding:", err)
			return
		}
		log.Println(fullRequestEncoded)
		// Getting lazy with the parsing here. Review this later.
		var searchMode, response, searchString string
		var radius int
		delimiter := "\n\n"
		caseSensitive := true
		options := strings.Split(fullRequestDecoded, "?")
		for _, option := range options {
			opt, value, _ := strings.Cut(option, "=")
			if opt == "searchString" {
				searchString = value
			} else if opt == "searchMode" {
				searchMode = value
			} else if opt == "delimiter" {
				delimiter = value
			} else if opt == "caseSensitive" {
				caseSensitive = value == "true" //cool way to turn this string into a bool
			} else if opt == "radius" {
				radius, _ = strconv.Atoi(value)
			}
		}

		switch searchMode {
		case "versesearch":
			//options include: delimiter="string" and radius=int
			response, err = loadVersebyStr(searchString, *map_of_verses)
			if err != nil {
				log.Println(err)
				fmt.Fprintf(w, "Error retrieving verse: %s", searchString)
			}
		case "stringsearch":
			//options include: delimiter="string" and caseSensitive=true
			response = searchBibleForStr(searchString, *map_of_verses, delimiter, caseSensitive)
		case "neartermsearch":
			fmt.Print(radius)
		}

		elapsed := time.Since(start_time).Milliseconds()
		//log.Println(verse)
		log.Printf("Total execution time for lookup: %d ms - Completed.\n", elapsed)
		fmt.Fprint(w, response)
	}
}

func main() {
	mapOfVerses := scanBibleFromTxtFile("bible/ESVBible.txt")

	//http://localhost/api?searchMode=stringsearch?searchString=Shem?caseSensitive=true
	http.HandleFunc("/api", requestHandler(&mapOfVerses))

	log.Println("Starting server on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
