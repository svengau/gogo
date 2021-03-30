package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func ask(question string) string {
	fmt.Print(question)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("An error occured while reading input. Please try again", err)
	}
	return strings.TrimSuffix(input, "\n")
}

func rpad(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

func failIf(err error, msg string) {
	if err != nil {
		fmt.Printf("%v: %w", msg, err)
		os.Exit(3)
	}
}
