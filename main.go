package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type Name struct {
	Name string `json:"name"`
}

type Type struct {
	Type Name `json:"type"`
}

type Pokemon struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Types     []Type    `json:"types"`
	Height    int       `json:"height"`
	Weight    int       `json:"weight"`
	Stats     []Stat    `json:"stats"`
	Abilities []Ability `json:"abilities"`
}

type ListePokemon struct {
	Results []Pokemon `json:"results"`
}

type Traduction struct {
	Name     string `json:"name"`
	Language Name   `json:"language"`
}
type Espece struct {
	Names             []Traduction  `json:"names"`
	Genera            []Genere      `json:"genera"`
	FlavorTextEntries []Description `json:"flavor_text_entries"`
	IsLegendary       bool          `json:"is_legendary"`
	IsMythical        bool          `json:"is_mythical"`
	EvolutionChain    EvolutionRef  `json:"evolution_chain"`
}

type PagePokemon struct {
	Pokemon        Pokemon
	NomFr          string
	CategorieFr    string
	DescriptionFr  string
	IsLegendary    bool
	IsMythical     bool
	EvolutionChain ChainLink
}

type StatDetail struct {
	Name string `json:"name"`
}

type Stat struct {
	BaseStat int        `json:"base_stat"`
	Stat     StatDetail `json:"stat"`
}

type AbilityDetail struct {
	Name string `json:"name"`
}

type Ability struct {
	Ability  AbilityDetail `json:"ability"`
	IsHidden bool          `json:"is_hidden"`
}

type Genere struct {
	Genus    string `json:"genus"`
	Language Name   `json:"language"`
}

type Description struct {
	FlavorText string `json:"flavor_text"`
	Language   Name   `json:"language"`
}

type ChainLink struct {
	Species   Name        `json:"species"`
	EvolvesTo []ChainLink `json:"evolves_to"`
}

type EvolutionChain struct {
	Chain ChainLink `json:"chain"`
}

type EvolutionRef struct {
	URL string `json:"url"`
}

func getPokemons() []Pokemon {
	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon?limit=151")
	if err != nil {
		fmt.Println("Erreur :", err)
		return nil
	}
	defer resp.Body.Close()

	var liste ListePokemon
	json.NewDecoder(resp.Body).Decode(&liste)

	for i := range liste.Results {
		liste.Results[i].ID = i + 1
	}

	return liste.Results
}

func getPokemon(id string) Pokemon {
	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + id)
	if err != nil {
		fmt.Println("Erreur :", err)
		return Pokemon{}
	}

	defer resp.Body.Close()
	var pokemon Pokemon
	json.NewDecoder(resp.Body).Decode(&pokemon)

	return pokemon
}

func getEspece(id string) Espece {
	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon-species/" + id)
	if err != nil {
		fmt.Println("Erreur :", err)
		return Espece{}
	}
	defer resp.Body.Close()
	var espece Espece
	json.NewDecoder(resp.Body).Decode(&espece)
	return espece
}

func getEvolutionChain(url string) EvolutionChain {
	trimmedURL := strings.TrimPrefix(url, "https://pokeapi.co/api/v2/")
	resp, err := http.Get("https://pokeapi.co/api/v2/" + trimmedURL)
	if err != nil {
		fmt.Println("Erreur :", err)
		return EvolutionChain{}
	}
	defer resp.Body.Close()

	var chain EvolutionChain
	json.NewDecoder(resp.Body).Decode(&chain)
	return chain
}

func nomFrancais(espece Espece) string {
	for _, traduction := range espece.Names {
		if traduction.Language.Name == "fr" {
			return traduction.Name
		}
	}
	return ""
}

func categorieFr(espece Espece) string {
	for _, g := range espece.Genera {
		if g.Language.Name == "fr" {
			return g.Genus
		}
	}
	return ""
}

func descriptionFr(espece Espece) string {
	for _, d := range espece.FlavorTextEntries {
		if d.Language.Name == "fr" {
			return d.FlavorText
		}
	}
	return ""
}

func erreur(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	tmpl := template.Must(template.ParseFiles("templates/erreur.html"))
	tmpl.Execute(w, message)
}

func accueil(w http.ResponseWriter, r *http.Request) {
	pokemons := getPokemons()

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, pokemons)
}

func fiche(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pokemon := getPokemon(id)
	if pokemon.ID == 0 {
		erreur(w, 404, "Ce Pokémon n'existe pas.")
		return
	}

	espece := getEspece(id)
	evolutionChain := getEvolutionChain(espece.EvolutionChain.URL)

	page := PagePokemon{
		Pokemon:        pokemon,
		NomFr:          nomFrancais(espece),
		CategorieFr:    categorieFr(espece),
		DescriptionFr:  descriptionFr(espece),
		IsLegendary:    espece.IsLegendary,
		IsMythical:     espece.IsMythical,
		EvolutionChain: evolutionChain.Chain,
	}

	funcMap := template.FuncMap{
		"div": func(a, b int) float64 {
			return float64(a) / float64(b)
		},
	}

	tmpl := template.Must(template.New("fiche.html").Funcs(funcMap).ParseFiles("templates/fiche.html"))
	tmpl.Execute(w, page)
}

func introuvable(w http.ResponseWriter, r *http.Request) {
	erreur(w, 404, "Cette page n'existe pas.")
}

func main() {
	http.HandleFunc("/{$}", accueil)
	http.HandleFunc("/pokemon/{id}", fiche)
	http.HandleFunc("/", introuvable)
	http.Handle("/static/", http.FileServer(http.Dir(".")))

	fmt.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
