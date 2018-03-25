package db

import (
	"PumpSrv/db"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Sites struct {
	Id     string `bson:"_id"`
	SiteId string
	Msgs   []Messages
}

type Messages struct {
	PumpId  string
	FlowGPM string
	Voltage string
	Status  string
}

//Log any errors
func LogError(data string, mgoDb *mgo.Session) {

	c := mgoDb.DB("SiteData").C("Errors")

	_, err := c.Upsert(bson.M{"_id": ""}, bson.M{"msgs": data})
	if err != nil {
		log.Printf("AddSiteMsg error = %s\n", err.Error())
	}

	mgoDb.Close()
}

//Get pump data for a site
func GetPumpData(siteId string, mgoDb *mgo.Session) Sites, error {
	//func GetPumpData(siteId string) {

	c := mgoDb.DB("SiteData").C("PumpInfo")
	msgs := Sites{}

	err := c.Find(bson.M{"_id": siteId}).One(&msgs)
	if err != nil {
		log.Printf("GetPumpData error = %s\n", err.Error())
	}

	mgoDb.Close()
	log.Println("In GetPumpData")

	return msgs, err
}

//Update pump values sent from controller
func UpdatePumpData(siteId string, pumpId string, flowGPM string, voltage string,
	status string, mgoDb *mgo.Session) error {

	c := mgoDb.DB("SiteData").C("PumpInfo")
	sites := Sites{}
	var id = siteId + pumpId

	sites.SiteId = siteId
	newMsg := Messages{}
	newMsg.PumpId = pumpId
	newMsg.Status = status
	newMsg.FlowGPM = flowGPM
	newMsg.Voltage = voltage
	sites.Msgs = append(sites.Msgs, newMsg)

	log.Printf("Msg = %v", sites)
	_, err := c.Upsert(bson.M{"_id": id}, bson.M{"PumpInfo": sites.Msgs})
	if err != nil {
		log.Printf("UpdatePumpData error = %s\n", err.Error())
		db.LogError(err.Error(), mgo.Session.Copy())
	}

	return err
}

//Add a new pump for a site
func AddPump(siteId string, pumpId string, mgoDb *mgo.Session) error {
	c := mgoDb.DB("SiteData").C("PumpInfo")
	sites := Sites{}
	var id = siteId + pumpId

	sites.SiteId = siteId
	newMsg := Messages{}
	newMsg.PumpId = pumpId
	newMsg.Status = "0"
	newMsg.FlowGPM = "0"
	newMsg.Voltage = "0"
	sites.Msgs = append(sites.Msgs, newMsg)

	_, err := c.Upsert(bson.M{"_id": id}, bson.M{"PumpInfo": sites.Msgs})
	if err != nil {
		log.Printf("AddSiteMsg error = %s\n", err.Error())
		db.LogError(err.Error(), mgo.Session.Copy())
	}

	mgoDb.Close()

	return err
}
