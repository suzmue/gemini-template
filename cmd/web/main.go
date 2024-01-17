package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var t = template.Must(template.ParseFiles("./index.html"))

func generateHandler(w http.ResponseWriter, r *http.Request) {
	image, prompt := r.FormValue("chosen-image"), r.FormValue("prompt")
	// 🔥 FILL THIS OUT FIRST! 🔥
	// 🔥 GET YOUR GEMINI API KEY AT 🔥
	// 🔥 https://makersuite.google.com/app/apikey 🔥
	API_KEY := "TODO"

	if API_KEY == "TODO" {
		fmt.Fprint(w, "Error: To get started, get an API key at https://makersuite.google.com/app/apikey and enter it in main.go")
		return
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(API_KEY))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro-vision") // use 'gemini-pro' for text -> text
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}

	// Read the file using the contents reader.
	contents, err := os.ReadFile("../../" + image)
	if err != nil {
		log.Println(err)
	}

	resp, err := model.GenerateContent(r.Context(), genai.Text(prompt), genai.ImageData("jpeg", contents))
	if err != nil {
		log.Println(err)
	}

	fmt.Fprint(w, sprintResponse(resp))

}

func sprintResponse(resp *genai.GenerateContentResponse) string {
	if resp == nil {
		return ""
	}

	var s string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				s += fmt.Sprintln(part)
			}
		}
	}
	return s
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: replace this with dynamic reading of fileserver?
	data := map[string]interface{}{"images": []string{
		"baked_goods_1.jpeg",
		"baked_goods_2.jpeg",
		"baked_goods_3.jpeg",
	},
		"prompt":   "Provide a recipe for the baked goods in the image",
		"response": "(Results will appear here)",
	}

	var err error
	switch r.URL.Path {
	case "/":
		err = t.Execute(w, data)
	}
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func main() {
	fs := http.FileServer(http.Dir("content"))
	http.Handle("/content/", http.StripPrefix("/content/", fs))
	http.HandleFunc("/api/generate", generateHandler)
	http.HandleFunc("/", indexHandler)
	panic(http.ListenAndServe(":8080", nil))
}
