package api

import (
	"net/http"

	"github.com/go-chi/render"
	log "github.com/go-pkgz/lgr"
	"github.com/pavkazzz/cocos/backend/app/rest"
	"github.com/pavkazzz/cocos/backend/app/store"
)

type public struct {
	dataService pubStore
}

type pubStore interface {
	CreateIngredients(ingredient store.Ingredient) (ingredientID string, err error)
	// UpdateIngredients(ingredient store.Ingredient) error
	GetIngredient(ingredientID string) (store.Ingredient, error)
	GetIngredients() ([]store.Ingredient, error)
	// FindIngredients(search string) ([]store.Ingredient, error)

	// CreateCocktails(comment store.Cocktail) (cocktailID string, err error)
	// UpdateCocktails(comment store.Cocktail) error
	// GetCocktail(cocktailID string) (store.Cocktail, error)
	// GetCocktails() ([]store.Cocktail, error)
	// FindCocktails(search string) ([]store.Cocktail, error)

	ValidateIngredient(*store.Ingredient) error
	ValidateCocktail(*store.Cocktail) error
}

// POST /ingredients - adds ingredient
func (p *public) createIngredientsCtrl(w http.ResponseWriter, r *http.Request) {

	ingredient := store.Ingredient{}
	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &ingredient); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't create ingredient", rest.ErrDecode)
		return
	}

	if err := p.dataService.ValidateIngredient(&ingredient); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "invalid ingredient", rest.ErrValidation)
		return
	}

	id, err := p.dataService.CreateIngredients(ingredient)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't save ingredient", rest.ErrInternal)
		return
	}

	// dataService modifies ingredient
	finalIngredient, err := p.dataService.GetIngredient(id)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't load created ingredient", rest.ErrInternal)
		return
	}
	log.Printf("[DEBUG] created commend %+v", finalIngredient)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &finalIngredient)
}

// GET /ingredient - get ingredient list
func (p *public) getIngredientListCtrl(w http.ResponseWriter, r *http.Request) {
	ingredients, err := p.dataService.GetIngredients()
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "invalid ingredient", rest.ErrValidation)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, ingredients)
}
