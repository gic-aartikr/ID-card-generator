package main

import (
	"encoding/json"
	"fmt"
	"idCardDemo/idcardService"
	"idCardDemo/pojo"
	"log"
	"strings"

	"net/http"
)

var con = idcardService.IDCardGenerator{}

const tokenIdForAdmin = "tokenAdmin123456783"

func init() {
	con.Server = "mongodb://localhost:27017"
	//  cla.Server = "mongodb+srv://m001-student:m001-mongodb-basics@sandbox.7zffz3a.mongodb.net/?retryWrites=true&w=majority"
	con.Database = "IDCardGenerator"
	con.Collection = "IDCard"

	con.Connect()
}

func main() {
	http.HandleFunc("/create-idCard/", createIdCardGenerator)
	http.HandleFunc("/write-to-pdf-idCard/", writeToPDF)
	http.HandleFunc("/update-idCard/", updateIDCard)
	http.HandleFunc("/search-idCard/", searchIDCard)
	http.HandleFunc("/delete-idCard/", deleteIDCard)
	fmt.Println("Excecuted Main Method")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createIdCardGenerator(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token := r.Header.Get("tokenid")

	admin := token == tokenIdForAdmin

	if !admin {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "POST" {

		respondWithError(w, http.StatusBadRequest, "Invalid Method")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
		return

	}

	files := r.MultipartForm.File["file"]
	if len(files) != 1 {
		respondWithError(w, http.StatusBadRequest, "Please provide only one excel file")
		return
	}
	requestBody := r.MultipartForm.Value["request"][0]

	var idCard pojo.IDCardGenerator

	err := json.Unmarshal([]byte(requestBody), &idCard)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	if result, err := con.InsertIDCardData(idCard, files); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": result,
		})
	}
	// var idCard pojo.IDCardGenerator

	// if err := json.NewDecoder(r.Body).Decode(&idCard); err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Invalid request")
	// }

	// if err := con.Insert(idCard); err != nil {

	// 	respondWithError(w, http.StatusBadRequest, "Invalid request")
	// } else {
	// 	respondWithJson(w, http.StatusAccepted, map[string]string{
	// 		"message": "Record inserted successfully",
	// 	})
	// }
}

func writeToPDF(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method != "GET" {
		respondWithError(w, http.StatusBadRequest, "Method not allowed")
	}
	id := strings.Split(r.URL.Path, "/")[2]

	fmt.Println("ID:", id)
	if err := con.WriteIDCardDataInPDF(id); err != nil {

		respondWithError(w, http.StatusBadRequest, err.Error())
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": "write pdf data successfully",
		})
	}
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {

	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func deleteIDCard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token := r.Header.Get("tokenid")

	id := strings.Split(r.URL.Path, "/")[2]

	if token != tokenIdForAdmin {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "DELETE" {
		respondWithError(w, http.StatusBadRequest, "Invalid Method")
		return
	}

	if deleted, err := con.DeleteIDCard(id); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": deleted,
		})
	}
}

func updateIDCard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	token := r.Header.Get("tokenid")

	admin := token == tokenIdForAdmin

	id := strings.Split(r.URL.Path, "/")[2]

	if !(admin) {
		respondWithError(w, http.StatusBadRequest, "Unauthorized")
		return
	}

	if r.Method != "PUT" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var idCard pojo.IDCardGenerator

	if err := json.NewDecoder(r.Body).Decode(&idCard); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	if updated, err := con.UpdateDataInIDcard(idCard, id); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, map[string]string{
			"message": updated,
		})
	}
}

func searchIDCard(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if r.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, "Invalid method")
		return
	}

	var searchData pojo.Search

	if err := json.NewDecoder(r.Body).Decode(&searchData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// if cl.City == "" {
	// 	respondWithError(w, http.StatusBadRequest, "Please provide city for search")
	// 	return
	// }

	fmt.Println(searchData)
	if searchdocs, err := con.SearchByNameEmployeeAndJoiningDate(searchData); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	} else {
		respondWithJson(w, http.StatusAccepted, searchdocs)

	}
}
