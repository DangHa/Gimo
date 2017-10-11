package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/pat"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Bao loi tra ve
func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}

//Phan hoi tra ve
func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(json)
}

//Document - Member
type ListMember struct { //Struct cua JSON input
	Id      bson.ObjectId `bson:"_id,omitempty"`
	Name    string        `bson:"Name"`
	Scores  int64         `bson:"Scores"`
	Email   string        `bson:"Email"`
	Country string        `bson:"Country"`
}

//Gui JSON khong chua email cho web
type NotEmailListMember struct {
	Position int16  `bson:"Position"`
	Name     string `bson:"Name"`
	Scores   int64  `bson:"Scores"`
	Country  string `bson:"Country"`
}

//Position - tra ve json chua position cho app IOS
type yourPosition struct {
	Position int64
}

//module topCountry
type topCountrys struct {
	Position int16
	Country  string
	Scores   int64
}

func main() {

	//***** Ket noi voi co so du lieu mongoDB *****
	fmt.Println("Connecting MongoDB ... ")

	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal("Connect error:", err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// ******top 10 nuoc dung dau (thinh thoang moi cap nhat 1 lan) *******
	var topCoun []topCountrys //Chua 10 nuoc co diem cao nhat
	var checkUpdate = true    //true: chua cap nhat, flase: da cap nhat

	//***** Khoi dong Server *****
	fmt.Println("Server running ... ")

	k := pat.New()
	k.Get("/country", topCountrySend(session, &topCoun, &checkUpdate))

	m := pat.New()
	m.Get("/member", topMember(session))
	m.Post("/member", InsertOrUpdateMember(session))
	m.Delete("/member", deleteMember(session))

	http.Handle("/member", m)
	http.Handle("/country", k)

	http.HandleFunc("/", Home)
	http.HandleFunc("/countryLeaders.html", Country)
	http.HandleFunc("/about.html", About)

	//Xoa sau khi day vao server
	http.HandleFunc("/hide.html", Hide)

	http.Handle("/view/", http.StripPrefix("/view", http.FileServer(http.Dir("view"))))
	port := "8080"
	http.ListenAndServe(":"+port, nil)
}

//Hide
func Hide(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("view/hide.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tpl.ExecuteTemplate(w, "hide.html", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//Home - leaderBoard
func Home(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("view/leaders.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tpl.ExecuteTemplate(w, "leaders.html", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//Country Board
func Country(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("view/countryLeaders.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tpl.ExecuteTemplate(w, "countryLeaders.html", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//About
func About(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("view/about.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tpl.ExecuteTemplate(w, "about.html", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//Update khi Hour%3==0
func topCountryUpdate(s *mgo.Session, topCoun *[]topCountrys) {
	session := s.Copy()
	defer session.Close()

	c := session.DB("Timo").C("ListMember")

	var members []ListMember
	err := c.Find(bson.M{}).Sort("-Scores").All(&members) //Chuyen BSON thanh struct (theo thu tu tu lon den be)
	if err != nil {
		log.Println("Failed get all members: ", err)
		return
	}

	//Set lai top Coun = null
	*topCoun = []topCountrys{}

	//Doi sang Country - Scores (2 mang)
	allCountrySearch := make(map[string]int16) //Name Country: Vi tri cua no trong array nameCountry
	var nameCountry []string
	var scoresCountry []int64
	var position int16 = 1 //Vi tri them vao

	for i := 0; i < len(members); i++ {
		if allCountrySearch[members[i].Country] == 0 {
			allCountrySearch[members[i].Country] = position
			nameCountry = append(nameCountry, members[i].Country)
			scoresCountry = append(scoresCountry, members[i].Scores)
			position++
		} else {
			scoresCountry[allCountrySearch[members[i].Country]-1] += members[i].Scores // allCountrySearch[members[i].Country] la vi tri cua nameCountry trong array nameCountry
		}
	}

	//Sap xep
	for i := 0; i < len(scoresCountry)-1; i++ {
		for j := i + 1; j < len(scoresCountry); j++ {
			if scoresCountry[i] < scoresCountry[j] {
				tmpScores := scoresCountry[i]
				scoresCountry[i] = scoresCountry[j]
				scoresCountry[j] = tmpScores

				tmpName := nameCountry[i]
				nameCountry[i] = nameCountry[j]
				nameCountry[j] = tmpName
			}
		}
	}

	//Cho 10 nuoc dung dau vao mang
	length := len(nameCountry)
	if length > 10 { //Tra ve 100 nguoi dung dau
		length = 10
	}

	for i := 0; i < length; i++ {
		*topCoun = append(*topCoun, topCountrys{int16(i + 1), nameCountry[i], scoresCountry[i]})
	}

}

//Top Country gui
func topCountrySend(s *mgo.Session, topCoun *[]topCountrys, checkUpdate *bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if *checkUpdate == true {
			*checkUpdate = false
			topCountryUpdate(s, topCoun) //Cap nhat
		}
		if time.Now().Hour()%3 == 0 && *checkUpdate == false {
			*checkUpdate = true
		}

		resBody, err := json.MarshalIndent(*topCoun, "", "  ") //Get 200
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, resBody, http.StatusOK)

	}
}

//lay du lieu trong server
func topMember(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//Tao luong moi
		session := s.Copy()
		defer session.Clone()

		c := session.DB("Timo").C("ListMember")

		var members []ListMember
		err := c.Find(bson.M{}).Sort("-Scores").All(&members) //Chuyen BSON thanh struct (theo thu tu tu lon den be)

		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all members: ", err)
			return
		}

		//Bo Email
		var membersNotEmail []NotEmailListMember
		length := len(members)
		if length > 100 { //Tra ve 100 nguoi dung dau
			length = 100
		}
		for i := 0; i < length; i++ {
			membersNotEmail = append(membersNotEmail, NotEmailListMember{int16(i + 1), members[i].Name, members[i].Scores, members[i].Country})
		}

		//Chuyen lai thanh JSON
		resBody, err := json.MarshalIndent(membersNotEmail, "", "  ") //Get 200
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, resBody, http.StatusOK)
	}
}

//add them 1 member neu trung email thi doi thanh update
func InsertOrUpdateMember(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//Tao luong moi
		session := s.Copy()
		defer session.Close()

		//Giai ma body(JSON) cua goi HTTP vao member
		var member ListMember
		err := json.NewDecoder(r.Body).Decode(&member)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB("Timo").C("ListMember")

		//*** Kiem tra co trung email ko *** Phai viet lai cai nay bang cach tim kiem cua mongoDB
		var members []ListMember
		err = c.Find(bson.M{}).All(&members)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all members: ", err)
			return
		}

		//Kiem tra xem update(1) hay insert(0)
		checkInsertOrUpdate := 0
		for i := 0; i < len(members); i++ {
			if members[i].Email == member.Email {
				checkInsertOrUpdate = 1
			}
		}

		//Xu li Update or insert
		if checkInsertOrUpdate == 0 {
			err = c.Insert(member) //Insert 201

			if err != nil {
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed insert book: ", err)
				return
			}
		}

		if checkInsertOrUpdate == 1 {
			errUpdate := c.Update(bson.M{"Email": member.Email}, &member)

			if errUpdate != nil {
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed update book: ", err)
				return
			}
		}

		//*** Xac dinh vi tri cua member vua input ***
		members = []ListMember{}
		err = c.Find(bson.M{}).Sort("-Scores").All(&members)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all members: ", err)
			return
		}

		var pos yourPosition //Vi tri trong bang
		for i := 0; i < len(members); i++ {
			if members[i].Email == member.Email {
				pos.Position = int64(i) + 1
			}
		}

		//Chuyen position lai thanh JSON
		yourPos, er := json.Marshal(pos)
		if er != nil {
			log.Fatal(er)
		}

		//Tra ve ket qua
		if checkInsertOrUpdate == 0 {
			ResponseWithJSON(w, yourPos, http.StatusCreated)
		}
		if checkInsertOrUpdate == 1 {
			ResponseWithJSON(w, yourPos, http.StatusAccepted)
		}

	}
}

//Xoa 1 member (hide)
func deleteMember(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB("Timo").C("ListMember")
		c.EnsureIndexKey("Email")

		//Giai ma body(JSON) cua goi HTTP vao member
		var member ListMember
		err := json.NewDecoder(r.Body).Decode(&member)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		err = c.Remove(bson.M{"Email": member.Email})
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed delete book: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "Book not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)

	}
}
