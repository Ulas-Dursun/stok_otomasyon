package models

type Product struct {
	ID          string `bson:"id" json:"id"`
	Adi         string `bson:"adi" json:"adi"`
	Miktar      uint   `bson:"miktar" json:"miktar"`
	StokDurumu  bool   `bson:"stok_durumu" json:"stok_durumu"`
	SatisDurumu bool   `bson:"satis_durumu" json:"satis_durumu"`
}
