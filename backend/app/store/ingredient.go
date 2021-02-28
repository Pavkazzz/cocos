package store

type Ingredient struct {
	ID     string `json:"id" bson:"_id"`
	Name   string `json:"name"`
	EnName string `json:"en_name"`
	Abv    int    `json:"abv"`
}
