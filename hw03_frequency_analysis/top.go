package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

type WordCount struct {
	word  string
	count int
}

var re = regexp.MustCompile(`(\p{L}|-{2,})(?:\S*(\p{L}|-))?`)

func Top10(text string) []string {
	mySlice := WordsCount(text)
	result := make([]string, 0, 10)
	for i := 0; i < 10 && i < len(mySlice); i++ {
		result = append(result, mySlice[i].word)
	}
	return result
}

func WordsCount(text string) []WordCount {
	word2count := make(map[string]int)
	for _, word := range ToWords(text) {
		word2count[word]++
	}
	mySlice := make([]WordCount, 0, len(word2count))
	for word, count := range word2count {
		mySlice = append(mySlice, WordCount{word, count})
	}
	sort.Slice(mySlice, func(i, j int) bool {
		if mySlice[i].count == mySlice[j].count {
			return mySlice[i].word < mySlice[j].word
		}
		return mySlice[i].count > mySlice[j].count
	})
	return mySlice
}

func ToWords(text string) []string {
	result := make([]string, 0, utf8.RuneCountInString(text)/5)
	matches := re.FindAllString(text, -1)
	for _, word := range matches {
		result = append(result, strings.ToLower(word))
	}
	return result
}
