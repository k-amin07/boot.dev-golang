package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/k-amin07/boot.dev-golang/3.Build-a-Pokedex-in-Go/internal/pokecache"
)

type config struct {
	Next     string
	Previous string
}

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *config) ([]byte, error)
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

func commandExit(cfg *config) ([]byte, error) {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil, nil
}

func commandHelp(cfg *config) ([]byte, error) {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nexit: Exit the Pokedex\n")
	return nil, nil
}

func commandMap(cfg *config) ([]byte, error) {
	resp, err := http.Get(cfg.Next)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {

	cache := pokecache.NewCache(5 * time.Second)

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
	var locations PokeAPIResponse
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
		default:
			elem.callback(nil)
			continue
		}
		pokeApiUrl := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/?offset=%d&limit=20", offset)
		if cacheRes, ok := cache.Get(pokeApiUrl); !ok {
			resp, err := elem.callback(&config{
				Previous: pokeApiUrl,
				Next:     pokeApiUrl,
			})
			if err != nil {
				continue
			}
			cache.Add(pokeApiUrl, resp)
			err = json.Unmarshal(resp, &locations)
			if err != nil {
				continue
			}

			for _, location := range locations.Results {
				fmt.Println(location.Name)
			}
		} else {
			err := json.Unmarshal(cacheRes, &locations)
			if err != nil {
				continue
			}

			for _, location := range locations.Results {
				fmt.Println(location.Name)
			}
		}

	}
}
