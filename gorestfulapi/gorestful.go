package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
)

type PostData struct {
	PostId      int    `json:"postid"`
	PostYazarId int    `json:"postyazarid"`
	PostTitle   string `json:"posttitle"`
}

var PostDatas []PostData
var PostDatasJSON []PostData

/*
func PostRequest() []PostData {
	response, err := http.Get("localhost:6060/GetById")
	if err != nil {
		fmt.Println("get errorr")
	}
	body, err := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &PostDatasJSON)
	return PostDatasJSON
	//var FetchedJson []PostData
	//json.Unmarshal(response, &FetchedJson)
	http.NewRequest()
}
*/
func GetById(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	PostDatas = PostDatas[:0]
	db, err := sql.Open("mysql", "root:1234@/db_blog")
	if err != nil {
		fmt.Println("Database acilirken bir sikinti olustu")
	}
	defer db.Close()
	MinId := r.Form["MinId"][0]
	MaxId := r.Form["MaxId"][0]
	fmt.Println("MınId:", MinId, "MaxId:", MaxId)
	ResultRows, err := db.Query("SELECT postID,yazarID,postTitle FROM postlar WHERE postID BETWEEN ? AND ?", MinId, MaxId)
	//ResultRows, err := db.Query("SELECT postID,yazarID,postTitle FROM postlar WHERE postID BETWEEN 20 AND 50")
	if err != nil {
		fmt.Println("DB'den istenen aralıkta değer çekilemedi")
	}
	var PostIdToken int
	var PostYazarIdToken int
	var PostTitleToken string
	for ResultRows.Next() {
		ResultRows.Scan(&PostIdToken, &PostYazarIdToken, &PostTitleToken)
		PostAdd := PostData{
			PostId:      PostIdToken,
			PostYazarId: PostYazarIdToken,
			PostTitle:   PostTitleToken,
		}
		fmt.Println(PostIdToken, PostYazarIdToken, PostTitleToken)
		PostDatas = append(PostDatas, PostAdd)
	}
	/*
		jsonbytes, err := json.Marshal(PostDatas)
		if err != nil {
			fmt.Println("Marshalda sıkıntı var")
		}
		w.Write(jsonbytes)
		http.Redirect(w, r, "http://localhost:6060/restapi", 301)
		//
	*/
	json.NewEncoder(w).Encode(PostDatas)
}

func anasayfa(w http.ResponseWriter, r *http.Request) {
	PostDatas = PostDatas[:0]
	template, err := template.ParseFiles("restfulhome.html")
	if err != nil {
		fmt.Println("home template hata")
	}
	template.Execute(w, nil)
}

func restapi(w http.ResponseWriter, r *http.Request) {
	PostDatasJSON = PostDatasJSON[:0]
	//urlData:=url.Values{}
	//urlData.Set("20","50")
	response, err := http.PostForm("http://localhost:6060/GetById", url.Values{"MinId": {"20"}, "MaxId": {"50"}})
	if err != nil {
		fmt.Println("request body okunamadı")
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Body okunamadı")
	}
	json.Unmarshal(body, &PostDatasJSON)
	for index, value := range PostDatasJSON {
		fmt.Println(index, ". Post Degeri:", value)
	}
	template, err := template.ParseFiles("printjsons.html")
	if err != nil {
		fmt.Println("template bulunamadı")
	}
	template.Execute(w, PostDatasJSON)
}

func main() {
	http.HandleFunc("/", anasayfa)
	http.HandleFunc("/GetById", GetById)
	http.HandleFunc("/restapi", restapi)
	err := http.ListenAndServe(":6060", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
}
