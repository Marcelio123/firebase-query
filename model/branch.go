package model

type DataBranch struct {
	UUID     string       `firestore:"uuid"`
	Name     string       `firestore:"name"`
	Whatsapp DataWhatsapp `firestore:"whatsapp"`
}

type DataWhatsapp struct {
	CountryNumber interface{} `firestore:"country_number"`
	Number        interface{} `firestore:"number"`
}
