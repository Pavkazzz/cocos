package service

import (
	"github.com/pavkazzz/cocos/backend/app/store"
	"github.com/pavkazzz/cocos/backend/app/store/engine"
)

// DataStore wraps store.Interface with additional methods
type DataStore struct {
	Engine engine.Interface
}

//
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
