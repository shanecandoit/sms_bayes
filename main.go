package main

import (
	"encoding/csv"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jbrukh/bayesian"
)

const (
	Good bayesian.Class = "Good"
	Bad  bayesian.Class = "Bad"
)

var globalClassifier *bayesian.Classifier

// data from https://www.kaggle.com/uciml/sms-spam-collection-dataset

func train(strings_bools map[string]bool) *bayesian.Classifier {

	classifier := bayesian.NewClassifier(Good, Bad)
	//goodStuff := []string{"tall", "rich", "handsome"}
	//goodStrings(strings_bools map[string]bool)
	goodStuff := goodStrings(strings_bools)
	for _, goodSample := range goodStuff {
		fmt.Println("goodSample", goodSample)
		goodWords := strings.Split(goodSample, " ")
		classifier.Learn(goodWords, Good)
	}

	//badStuff := []string{"poor", "smelly", "ugly"}
	//
	badStuff := badStrings(strings_bools)
	for _, badSample := range badStuff {
		fmt.Println("badSample", badSample)
		badWords := strings.Split(badSample, " ")
		classifier.Learn(badWords, Bad)
	}
	classifier.Learn(goodStuff, Good)

	return classifier
}

func goodStrings(strings_bools map[string]bool) []string {
	goodStrs := make([]string, len(strings_bools))
	for str, b := range strings_bools {
		if b {
			goodStrs = append(goodStrs, str)
			//goodStrs.append(str)
		}
	}
	return goodStrs
}

func badStrings(strings_bools map[string]bool) []string {
	badStrs := make([]string, len(strings_bools))
	for str, b := range strings_bools {
		if b == false {
			badStrs = append(badStrs, str)
		}
	}
	return badStrs
}

func loadFile(path string) (map[string]bool, error) {

	recordsOK := make(map[string]bool)

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error loading file:", path)
		log.Fatal(err)
		return nil, err
	}

	r := csv.NewReader(file)

	for {
		record, err := r.Read()

		if err == io.EOF {
			fmt.Println("EOF")
			break
		}
		if err != nil {
			fmt.Println("err", err)
			log.Fatal(err)
		}

		fmt.Println(len(record))
		fmt.Println(record)

		// 0 class
		// 1 message
		if len(record[0]) > 1 && len(record[1]) > 4 {
			class, mesg := record[0], record[1]
			fmt.Println("class", class, "mesg", mesg)
			//recordsOK[record] = true

			var isok = false
			if class == "ham" {
				isok = true
			}
			// else if class == "spam" {
			//	isok = false
			//}

			recordsOK[mesg] = isok
			//return nil, nil
		} else if len(record) == 0 {
			// skip empty lines
			continue
		}
	}

	return recordsOK, nil
}

func probBad(classifier *bayesian.Classifier, input string) float64 {
	sampleWords := strings.Split(input, " ")
	scores, count, str := classifier.ProbScores(sampleWords)
	fmt.Println("scores", scores)
	fmt.Println("count", count)
	fmt.Println("strict", str)

	// scores[0] is good
	// scores[1] is bad
	probIsBad := scores[1] - scores[0]

	return probIsBad
}

func main() {
	f := "spam.csv"
	sms_ok, err := loadFile(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("loaded ok")

	fmt.Println("sms counts", len(sms_ok))

	//*bayesian.Classifier
	classifier := train(sms_ok)
	globalClassifier = classifier
	howManyTrained := classifier.Learned()
	fmt.Println("training eamples", howManyTrained)
	// training eamples 15481

	newSample := "hey, how are you doing?"
	sampleWords := strings.Split(newSample, " ")

	scores, count, str := classifier.ProbScores(sampleWords)
	fmt.Println("scores", scores)
	fmt.Println("count", count)
	fmt.Println("strict", str)
	/*
		scores [0.9999999999999996 3.7512045385811273e-16]
		count 0
		strict true
	*/

	badSample := "Hi babe its Princess Twilight, how r u? Im home from abroad and lonely, text me back if u wanna chat xxSP expensivesms.com Text stop to stop"
	badWords := strings.Split(badSample, " ")
	scores, count, str = classifier.ProbScores(badWords)
	fmt.Println("scores", scores)
	fmt.Println("count", count)
	fmt.Println("strict", str)
	/*
		scores [1.0734089649579967e-25 1]
		count 1
		strict true
	*/

	fmt.Println("listenting on localhost:9090")
	http.HandleFunc("/", defaultHandler)
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])

	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		fmt.Fprintf(w, homePage)
	} else if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		name := r.FormValue("name")
		message := r.FormValue("message")

		probIsBad := probBad(globalClassifier, message)

		fmt.Fprintf(w, "<html><pre> \n")
		fmt.Fprintf(w, "Name \t %s\n", html.EscapeString(name))
		fmt.Fprintf(w, "message \t %s\n", html.EscapeString(message))
		fmt.Fprintf(w, "probabilityIsBad \t %f\n", probIsBad)
		fmt.Fprintln(w, `<br><a href="/">back</a>`)
	} else {
		http.Error(w, "404 not found.", http.StatusNotFound)
	}
}

const homePage = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
</head>
<body>
<div>
  	<form method="POST" action="/">     
		<label>Name<br>
			<input name="name" type="text" value="" />
			</label>
		<br>
		<label>Message<br>
			<textarea name="message"  cols="40" rows="9"></textarea>
		</label>
		<br>
		<input type="submit" value="submit" />
  	</form>
</div>
</body>
</html>
`
