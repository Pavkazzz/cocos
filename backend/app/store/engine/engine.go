package engine

import "github.com/pavkazzz/cocos/backend/app/store"

// Interface defines methods provided by low-level storage engine
type Interface interface {
	CreateIngredients(ingredient store.Ingredient) (ingredientID string, err error)
	UpdateIngredients(ingredient store.Ingredient) error
	GetIngredient(ingredientID string) (store.Ingredient, error)
	GetIngredients() ([]store.Ingredient, error)
	FindIngredients(search string) ([]store.Ingredient, error)

	CreateCocktails(comment store.Cocktail) (cocktailID string, err error)
	UpdateCocktails(comment store.Cocktail) error
	GetCocktail(cocktailID string) (store.Cocktail, error)
	GetCocktails() ([]store.Cocktail, error)
	FindCocktails(search string) ([]store.Cocktail, error)

	Close() error // close storage engine
}
