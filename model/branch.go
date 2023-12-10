package model

type DataBranch struct {
	UUID     string                 `firestore:"uuid"`
	Name     string                 `firestore:"name"`
	Whatsapp map[string]interface{} `firestore:"whatsapp"`
}
