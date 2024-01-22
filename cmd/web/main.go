package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// 🔥 FILL THIS OUT FIRST! 🔥
// 🔥 GET YOUR GEMINI API KEY AT 🔥
// 🔥 https://makersuite.google.com/app/apikey 🔥
// This can also be provided as the API_KEY environment variable.
var apiKey = "TODO"

func usage() {
	fmt.Fprintf(os.Stderr, "usage: web [options]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	addr = flag.String("addr", "localhost:8080", "address to serve")
)

func generateHandler(w http.ResponseWriter, r *http.Request, model *genai.GenerativeModel) {
	if apiKey == "TODO" {
		http.Error(w, "Error: To get started, get an API key at https://makersuite.google.com/app/apikey and enter it in main.go", http.StatusInternalServerError)
		return
	}

	image, prompt := r.FormValue("chosen-image"), r.FormValue("prompt")
	contents, err := os.ReadFile(filepath.Join("static", "images", filepath.Base(image)))
	if err != nil {
		log.Printf("Unable to read image %s: %v\n", image, err)
		http.Error(w, "Error: unable to generate content", http.StatusInternalServerError)
		return
	}

	// Generate the response and aggregate the streamed response.
	iter := model.GenerateContentStream(r.Context(), genai.Text(prompt), genai.ImageData("jpeg", contents))
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error generating content: %v\n", err)
			http.Error(w, "Error: unable to generate content", http.StatusInternalServerError)
			return
		}
		if resp == nil {
			continue
		}
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					fmt.Fprint(w, part)
				}
			}
		}
	}
}

type Page struct {
	Images []string
}

var tmpl = template.Must(template.ParseFiles("static/index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Load all baked goods images from the static/images directory.
	matches, err := filepath.Glob(filepath.Join("static", "images", "baked_goods_*.jpeg"))
	if err != nil {
		log.Printf("Error loading baked goods images: %v", err)
	}
	var page = &Page{Images: make([]string, len(matches))}
	for i, match := range matches {
		page.Images[i] = filepath.Base(match)
	}
	switch r.URL.Path {
	case "/":
		err = tmpl.Execute(w, page)
		if err != nil {
			log.Printf("Template execution error: %v", err)
		}
	}
}

func main() {
	// Parse flags.
	flag.Usage = usage
	flag.Parse()

	// Parse and validate arguments (none).
	args := flag.Args()
	if len(args) != 0 {
		usage()
	}

	// Get the Gemini API key from the environment.
	if key := os.Getenv("API_KEY"); key != "" {
		apiKey = key
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		log.Println(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-pro-vision") // use 'gemini-pro' for text -> text
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}

	// Serve static files and handle API requests.
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) { generateHandler(w, r, model) })
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
