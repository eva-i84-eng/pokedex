package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Name struct {
	Name string `json:"name"`
}

type Type struct {
	Type Name `json:"type"`
}

type Pokemon struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Types []Type `json:"types"`
}

type ListePokemon struct {
	Results []Pokemon `json:"results"`
}

type Traduction struct {
	Name     string `json:"name"`
	Language Name   `json:"language"`
}
type Espece struct {
	Names []Traduction `json:"names"`
}

type PagePokemon struct {
	Pokemon Pokemon
	NomFr   string
}

type StatDetail struct {
	Name string `json:"name"`
}

type Stat struct {
	BaseStat int        `json:"BaseStat"` // Attention aux majuscules dans le JSON (base_stat)
	Stat     StatDetail `json:"stat"`
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

func nomFrancais(espece Espece) string {
	for _, traduction := range espece.Names {
		if traduction.Language.Name == "fr" {
			return traduction.Name
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
	page := PagePokemon{
		Pokemon: pokemon,
		NomFr: nomFrancais(getEspece(id)),
	}
	tmpl := template.Must(template.ParseFiles("templates/fiche.html"))
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
