package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/context"
	"gopkg.in/mgo.v2"
)

func main() {

	// connect to the database
	db, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal("cannot dial mongo", err)
	}
	defer db.Close() // clean up when we're done

	// Adapt our handle function using withDB
	h := Adapt(http.HandlerFunc(handle), withDB(db))

	// add the handler
	http.Handle("/comments", context.ClearHandler(h))

	// start the server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleRead(w, r)
	case "POST":
		handleInsert(w, r)
	default:
		http.Error(w, "Not supported", http.StatusMethodNotAllowed)
	}
}

type comment struct {
	ID     bson.ObjectId `json:"id" bson:"_id"`
	Author string        `json:"author" bson:"author"`
	Text   string        `json:"text" bson:"text"`
	When   time.Time     `json:"when"  bson:"when"`
}


type SiteInfo struct {
    ID              bson.ObjectId   `json:"id" bson:"_id"`
    When            time.Time       `json:"when" bson:"when"`
    SiteName        string          `json:"siteName" bson:"siteName"`
    SiteSubTitle    string          `json:"siteSubTitle" bson:"siteSubTitle"`
    HeaderUrl       string          `json:"headerUrl" bson:"headerUrl"`
    SiteDesc        string          `json:"siteDesc" bson:"siteDesc"`
    SiteCreator     string          `json:"siteCreator" bson:"siteCreator"`           

}

type FrontPageProfile struct {
    ID              bson.ObjectId       `json:"id" bson:"_id"`
    When            time.Time           `json:"when" bson:"when"`
    Name            string              `json:"name" bson:"name"`
    PictureUrl      string              `json:"pictureUrl" bson:"pictureUrl"`
    SkillsShort     string              `json:"skillsShort" bson:"skillsShort"`
    ShortBio        string              `json:"shortBio" bson:"shortBio"`
}

type RealProfile struct {
    ID              bson.ObjectId       `json:"id" bson:"_id"`
    When            time.Time           `json:"when" bson:"when"`
	Name            string              `json:"name" bson:"name"`      
    LongBio         string              `json:"pictureUrl" bson:"pictureUrl"`
    AlmaMater       string              `json:"almaMater" bson:"almaMater"`
	Pictures        []string            `json:"pictures" bson:"pictures"`
	Skills          []PersonSkills      `json:"skills" bson:"skills"`
	Contact         ContactInfo         `json:"contact" bson:"contact"`

}

type PersonSkills struct {
	SkillName, SkillPhoto, SkillUse string
}

type ContactInfo struct {
	Email, Discord, Phone, Telegram string
}

func handleInsert(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "database").(*mgo.Session)

	// decode the request body
	var c comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// give the comment a unique ID
	c.ID = bson.NewObjectId()
	c.When = time.Now()

	// insert it into the database
	if err := db.DB("commentsapp").C("comments").Insert(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// redirect to it
	http.Redirect(w, r, "/comments/"+c.ID.Hex(), http.StatusTemporaryRedirect)
}
func handleRead(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "database").(*mgo.Session)

	// load the comments
	var comments []*comment
	if err := db.DB("commentsapp").C("comments").
		Find(nil).Sort("-when").Limit(100).All(&comments); err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write it out
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func withDB(db *mgo.Session) Adapter {

	// return the Adapter
	return func(h http.Handler) http.Handler {

		// the adapter (when called) should return a new handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// copy the database session
			dbsession := db.Copy()
			defer dbsession.Close() // clean up

			// save it in the mux context
			context.Set(r, "database", dbsession)

			// pass execution to the original handler
			h.ServeHTTP(w, r)

		})
	}
}


/*

jData, err := json.Marshal(Data)
if err != nil {
    // handle error
}
w.Header().Set("Content-Type", "application/json")
w.Write(jData)

*/