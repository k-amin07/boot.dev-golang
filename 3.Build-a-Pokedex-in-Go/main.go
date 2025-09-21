package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type config struct {
	Next     string
	Previous string
}

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *config) error
}

type PokeAPIResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func cleanInput(text string) []string {
	var returnValue []string
	substrings := strings.Split(strings.Trim(text, " "), " ")
	for _, substr := range substrings {
		returnValue = append(returnValue, strings.ToLower(substr))
	}

	return returnValue
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nexit: Exit the Pokedex\n")
	return nil
}

func commandMap(cfg *config) error {
	var locations PokeAPIResponse
	resp, err := http.Get(cfg.Next)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&locations); err != nil {
		return err
	}

	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}

	return nil
}

func main() {

	commandMap := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of 20 previous location areas",
			callback:    commandMap,
		},
	}

	offset := -20

	userInput := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		userInput.Scan()
		command := cleanInput(userInput.Text())[0]
		elem, ok := commandMap[command]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		switch command {
		case "map":
			offset += 20
		case "mapb":
			if offset-20 < 0 {
				fmt.Println("you're on the first page")
				continue
			}
			offset -= 20
		}
		elem.callback(&config{
			Previous: fmt.Sprintf("https://pokeapi.co/api/v2/location-area/?offset=%d&limit=20", offset),
			Next:     fmt.Sprintf("https://pokeapi.co/api/v2/location-area/?offset=%d&limit=20", offset),
		})
	}
}
