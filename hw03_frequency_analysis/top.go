package hw03frequencyanalysis

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type EntryMap struct {
	word  string
	count int
}

func Top10(text string) []string {
	word2count := make(map[string]int, 10)
	var currentWord strings.Builder
	for _, r := range text {
		if unicode.IsSpace(r) {
			if currentWord.String() != "" {
				word2count[currentWord.String()]++
				currentWord.Reset()
			}
		} else {
			currentWord.WriteRune(r)
		}
	}
	mySlice := make([]EntryMap, 0, len(word2count))

	for word, count := range word2count {
		mySlice = append(mySlice, EntryMap{word, count})
	}
	sort.Slice(mySlice, func(i, j int) bool {
		if mySlice[i].count == mySlice[j].count {
			return mySlice[i].word < mySlice[j].word
		}
		return mySlice[i].count > mySlice[j].count
	})
	fmt.Println(mySlice)
	result := make([]string, 0, 10)
	for i := 0; i < len(mySlice) && i < 10; i++ {
		result = append(result, mySlice[i].word)
	}
	return result
}
