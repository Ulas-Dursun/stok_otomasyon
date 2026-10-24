package models

type Info struct {
	ID            string `bson:"id" json:"id"`
	Adres         string `bson:"adres" json:"adres"`
	Telefon       string `bson:"telefon" json:"telefon"`
	CalisanSayisi uint   `bson:"calisan_sayisi" json:"calisan_sayisi"`
}
