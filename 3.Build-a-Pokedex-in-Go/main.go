package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
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
	callback    func(string, *config) ([]byte, error)
}

type PokeAPIResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

func cleanInput(text string) []string {
	var returnValue []string
	substrings := strings.Split(strings.Trim(text, " "), " ")
	for _, substr := range substrings {
		returnValue = append(returnValue, strings.ToLower(substr))
	}

	return returnValue
}

func commandExit(location_name string, cfg *config) ([]byte, error) {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil, nil
}

func commandHelp(location_name string, cfg *config) ([]byte, error) {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nexit: Exit the Pokedex\n")
	return nil, nil
}

func commandMap(location_name string, cfg *config) ([]byte, error) {
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

func commandExplore(location_url string, cfg *config) ([]byte, error) {
	resp, err := http.Get(location_url)
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

func getPokemonDataFromApi(pokemonName string) ([]byte, error) {
	pokemonUrl := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	resp, err := http.Get(pokemonUrl)
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

func getPokemonData(pokemonName string, cache *pokecache.Cache, pokemon *Pokemon) error {
	cacheRes, ok := cache.Get(pokemonName)
	if !ok {
		pokemonData, err := getPokemonDataFromApi(pokemonName)
		if err != nil {
			return err
		}
		err = json.Unmarshal(pokemonData, &pokemon)
		if err != nil {
			return err
		}
		cache.Add(pokemonName, pokemonData)
	} else {
		err := json.Unmarshal(cacheRes, &pokemon)
		if err != nil {
			pokemonData, err := getPokemonDataFromApi(pokemonName)
			if err != nil {
				return err
			}
			err = json.Unmarshal(pokemonData, &pokemon)
			if err != nil {
				return err
			}
			cache.Add(pokemonName, pokemonData)
		}
	}
	return nil
}

func commandCatch(pokemonName string, cache *pokecache.Cache) (*Pokemon, error) {
	var pokemon Pokemon
	err := getPokemonData(pokemonName, cache, &pokemon)
	if err != nil {
		return nil, err
	}

	var catchProbability int
	if pokemon.BaseExperience < 100 {
		catchProbability = 95
	} else if pokemon.BaseExperience < 200 {
		catchProbability = 80
	} else if pokemon.BaseExperience < 300 {
		catchProbability = 60
	} else if pokemon.BaseExperience < 400 {
		catchProbability = 35
	} else {
		catchProbability = 20
	}

	randomInteger := rand.IntN(100)
	if catchProbability > randomInteger {
		return &pokemon, nil
	}

	return nil, nil
}

func printResults(command string, resp []byte) {
	var locations PokeAPIResponse
	err := json.Unmarshal(resp, &locations)
	if err != nil {
		return
	}
	if command == "explore" {
		fmt.Println("Found Pokemon:")
		for _, encounters := range locations.PokemonEncounters {
			fmt.Printf(" - %s\n", encounters.Pokemon.Name)
		}
	} else {
		for _, location := range locations.Results {
			fmt.Println(location.Name)
		}
	}
}

func executeCommand(cache *pokecache.Cache, command *cliCommand, pokeApiUrl string) {
	cacheRes, ok := cache.Get(pokeApiUrl)
	if !ok {
		resp, err := command.callback(pokeApiUrl, &config{
			Previous: pokeApiUrl,
			Next:     pokeApiUrl,
		})
		if err != nil {
			return
		}
		cache.Add(pokeApiUrl, resp)
		cacheRes = resp
	}
	printResults(command.name, cacheRes)
}

func getCommandMap() map[string]cliCommand {
	return map[string]cliCommand{
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
		"explore": {
			name:        "explore",
			description: "Displays a list of all the Pokémon located in an area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Tries to catch a pokemon",
			callback:    nil,
		},
		"inspect": {
			name:        "inspect",
			description: "Displays the data of a caught pokemon",
			callback:    nil,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Displays the data of all caught pokemon",
			callback:    nil,
		},
	}
}

func main() {
	cache := pokecache.NewCache(5 * time.Second)
	commandMap := getCommandMap()

	offset := -20
	userInput := bufio.NewScanner(os.Stdin)

	var params string
	var args string

	pokedex := make(map[string]*Pokemon)

	for {
		params = ""
		args = ""
		path := "location-area"
		fmt.Print("Pokedex > ")
		userInput.Scan()
		userInputArray := cleanInput(userInput.Text())
		userCommand := userInputArray[0]
		if len(userInputArray) > 1 {
			args = userInputArray[1]
		}
		command, ok := commandMap[userCommand]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		switch command.name {
		case "map":
			offset += 20
			params = fmt.Sprintf("?offset=%d&limit=20", offset)
		case "mapb":
			if offset-20 < 0 {
				fmt.Println("you're on the first page")
				continue
			}
			offset -= 20
			params = fmt.Sprintf("?offset=%d&limit=20", offset)
		case "explore":
			params = args
			fmt.Printf("Exploring %s...\n", params)
		case "catch":
			path = "pokemon"
			fmt.Printf("Throwing a Pokeball at %s...\n", args)
			pokemon, err := commandCatch(args, cache)
			if err != nil || pokemon == nil {
				fmt.Printf("%s escaped!\n", args)
				continue
			}
			fmt.Printf("%s was caught!\nYou may now inspect it with the inspect command\n", args)
			pokedex[args] = pokemon
			continue
		case "inspect":
			pokemon, ok := pokedex[args]
			if !ok {
				fmt.Println("you have not caught that pokemon")
				continue
			}
			fmt.Printf("%+v\n", pokemon)
			continue
		case "pokedex":
			fmt.Println("Your Pokedex:")
			for _, val := range pokedex {
				fmt.Printf(" - %s\n", val.Name)
			}
			continue
		default:
			params = ""
			command.callback("", nil)
			continue
		}
		pokeApiUrl := fmt.Sprintf("https://pokeapi.co/api/v2/%s/%s", path, params)
		executeCommand(cache, &command, pokeApiUrl)
	}
}
