package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dslipak/pdf"

	openai "github.com/sashabaranov/go-openai"
)

func readPdf(path string) (string, error) {
	r, err := pdf.Open(path)
	// remember close file

	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
    b, err := r.GetPlainText()
    if err != nil {
        return "", err
    }
    buf.ReadFrom(b)
	return buf.String(), nil
}

func generateImage(){
	c := openai.NewClient("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")
	ctx := context.Background()

	// Sample image by link
	reqUrl := openai.ImageRequest{
		Prompt:         "Street Style fashion 2023",
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := c.CreateImage(ctx, reqUrl)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}
	fmt.Println(respUrl.Data[0].URL)

	// Example image as base64
	reqBase64 := openai.ImageRequest{
		Prompt:         "Street Wear fashion 2013",
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              3,
	}

	respBase64, err := c.CreateImage(ctx, reqBase64)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		fmt.Printf("PNG decode error: %v\n", err)
		return
	}

	file, err := os.Create("example.png")
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, imgData); err != nil {
		fmt.Printf("PNG encode error: %v\n", err)
		return
	}

	fmt.Println("The image was saved as example.png")
}

func uploadHandler(w http.ResponseWriter, r *http.Request){
	err := r.ParseMultipartForm(52 << 20) // maxMemory 52MB
	if err != nil {
		http.Error(w, "Max size is 52MB", http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the uploads folder if it doesn't
	// already exist
	err = os.MkdirAll("./pdf", os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new file in the pdf directory
	dst, err := os.Create(fmt.Sprintf("./pdf/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer dst.Close()


	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// filetype := http.DetectContentType(buff)
	// if filetype != "pdf" {
	// 	http.Error(w, "The provided file format is not allowed. Please upload a pdf file", http.StatusBadRequest)
	// 	return
	// }

	fmt.Fprintf(w, "File uploaded successfully: ")
	fmt.Fprintf(w, fileHeader.Filename)
	fmt.Println(fileHeader.Filename)

	// truncated for brevity
}

type requestBody struct {
	Prompt string `json:"prompt"`
	Size string `json:"size"`
	ResponseFormat string `json:"response_format"`
	N int32 `json:"n"`
	Url string `json:"url"`
}

func generateSpecImage(w http.ResponseWriter, r *http.Request){
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
	}

	var rb requestBody
	err = json.Unmarshal(body, rb)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
	}


}

func main() {
	fmt.Println("Start Running")
	fmt.Println("Listening in PORT 4500")

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)


	//pdf.DebugOn = true
	content, err := readPdf("./pdf/1692122891294243000.pdf") // Read local pdf file
	if err != nil {
		panic(err)
	}
	fmt.Println(content)

	concated_string := fmt.Sprintf("%s\n\n%s", "Can you take the resume about the text below?", content)
	client := openai.NewClient("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")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: concated_string,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
	
	if err := http.ListenAndServe(":4500", mux); err != nil {
		log.Fatal(err)
	}

}