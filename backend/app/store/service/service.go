package service

import (
	"strings"

	"github.com/pavkazzz/cocos/backend/app/store"
	"github.com/pavkazzz/cocos/backend/app/store/engine"
)

// DataStore wraps store.Interface with additional methods
type DataStore struct {
	Engine engine.Interface
}

// ValidateIngredient
func (s *DataStore) ValidateIngredient(i *store.Ingredient) error {
	return nil
}

//
func (s *DataStore) ValidateCocktail(c *store.Cocktail) error {
	return nil
}

//
func (s *DataStore) CreateIngredients(ingredient store.Ingredient) (ingredientID string, err error) {
	return s.Engine.CreateIngredients(ingredient)
}

func (s *DataStore) GetIngredient(ingredientID string) (store.Ingredient, error) {
	return s.Engine.GetIngredient(ingredientID)
}

func (s *DataStore) GetIngredients() ([]store.Ingredient, error) {
	return s.Engine.GetIngredients()
}


func (s *DataStore) FindIngredients(search string) ([]store.Ingredient, error) {
	ingredients, err := s.GetIngredients()
	if err != nil {
		return ingredients, err
	}
	res := []store.Ingredient{}
	for _, ingredient := range ingredients {
		if strings.Contains(ingredient.Name, search) || strings.Contains(ingredient.EnName, search) {
			res = append(res, ingredient)
		}
	}
	return res, nil
}

func (s *DataStore) UpdateIngredients(ingredient store.Ingredient) error {
	return s.Engine.UpdateIngredients(ingredient)
}

func (s *DataStore) Search(search string) ([]store.Ingredient, []store.Cocktail, error) {
	ingredients, err := s.FindIngredients(search)
	return ingredients, nil, err
}
