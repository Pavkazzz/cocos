package store

type Cocktail struct {
	ID         string `json:"id" bson:"_id"`
	Name       string
	Ingredient []Ingredient
}
