package main

import (
	"PumpSrv/db"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"

	"time"
)

var mgoSession *mgo.Session

// app settings
type Settings struct {
	MongoUrl string `json:"mongoUrl"`
	CertName string `json:"certname"`
	KeyName  string `json:"keyname"`
	Port     int    `json:"port"`
}

var appSettings = Settings{}

//Read settings file and populate structure
func getSettings() (err error) {

	settingsStr, err := ioutil.ReadFile("./settings.json")

	err = json.Unmarshal(settingsStr, &appSettings)

	if err != nil {
		log.Printf("Error = %s\n", err)
		return err
	}

	return err

}

//Get pump data for a specific site
func getPumpInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	siteId := p.ByName("SiteId")
	log.Printf(siteId)
	log.Printf("MGO = %v\n", mgoSession)
	m, err := db.GetPumpData(siteId, mgoSession.Copy())

	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("Error marshalling site data. %s", err.Error())
		io.WriteString(w, err.Error())
	}
	else {

	  io.WriteString(w, string(b))
	  log.Printf("Data = %v", m)
	}

}

//Add new pump at a site
func addPumpInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	siteId := p.ByName("SiteId")
	log.Printf(siteId)

	pumpId := p.ByName("PumpId")
	log.Printf(pumpId)

	err := db.AddPump(siteId, pumpId, mgoSession.Copy())
	if err != nil {
		log.Printf("Error addPumpInfo %s", err.Error())
		io.WriteString(w, err.Error())
	} else {
		io.WriteString(w, "OK")
	}
}

//Update pump data from local controller
func updatePumpData(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	siteId := p.ByName("SiteId")
	pumpId := p.ByName("PumpId")
	status := p.ByName("Status")
	flowGPM := p.ByName("FlowGPM")
	voltage := p.ByName("Voltage")

	err := db.UpdatePumpData(siteId, pumpId, status, flowGPM, voltage, mgoSession.Copy())

	if err != nil {
		log.Printf("Error updatePumpData %s", err.Error())
		io.WriteString(w, err.Error())
	} else {
		io.WriteString(w, "OK")
	}

}

//refresh mongo session
func storageDbKeepAlive() {

	mgoSession.Refresh()

}

//refresh mongo session every 2 minutes
func dbKeepAlives() {

	t := time.NewTicker(time.Minute * 2)
	for _ = range t.C {

		storageDbKeepAlive()
	}
}

func main() {

	err := getSettings()
	if err != nil {
		log.Fatal("Error reading settings file: " + err.Error())
	}

	//setup mongoDb
	mgoSession, err = mgo.Dial(appSettings.MongoUrl)
	if err != nil {
		log.Panicf("Error Connection to Db %s\n", err.Error())
	}
	mgoSession.SetMode(mgo.Monotonic, true)

	defer mgoSession.Close()

	go dbKeepAlives()

	r := httprouter.New()

	r.GET("/GetPumpData/:SiteId", getPumpInfo)
	r.POST("/AddPump/:SiteId/:PumpId", addPumpInfo)
	r.PUT("/UpdatePumpData/:SiteId/:PumpId/:Status/:FlowGPM/:Voltage", updatePumpData)

	log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(appSettings.Port),
		appSettings.CertName, appSettings.KeyName, r))

	//Non SSL used for testing
	//log.Fatal(http.ListenAndServe(":"+strconv.Itoa(appSettings.Port), r))
}
