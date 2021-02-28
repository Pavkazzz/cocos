package engine

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"

	log "github.com/go-pkgz/lgr"
	"github.com/pavkazzz/cocos/backend/app/store"
	"github.com/pkg/errors"
)

// BoltDB implements store.Interface, represents multiple sites with multiplexing to different bolt dbs. Thread safe.
// there are 6 types of top-level buckets:
//  - comments for post in "posts" top-level bucket. Each url (post) makes its own bucket and each k:v pair is commentID:comment
//  - history of all comments. They all in a single "last" bucket (per site) and key is defined by ref struct as ts+commentID
//    value is not full comment but a reference combined from post-url+commentID
//  - user to comment references in "users" bucket. It used to get comments for user. Key is userID and value
//    is a nested bucket named userID with kv as ts:reference
//  - users details in "user_details" bucket. Key is userID, value - UserDetailEntry
//  - blocking info sits in "block" bucket. Key is userID, value - ts
//  - counts per post to keep number of comments. Key is post url, value - count
//  - readonly per post to keep status of manually set RO posts. Key is post url, value - ts
type BoltDB struct {
	db *bolt.DB
}

const (
	// top level buckets
	ingredientBucketName = "posts"
	cocktailBucketName   = "last"
	userBucketName       = "users"

	tsNano = "2006-01-02T15:04:05.000000000Z07:00"
)

// BoltSite defines single site param
type BoltSite struct {
	FileName string // full path to boltdb
}

// NewBoltDB makes persistent boltdb-based store. For each site new boltdb file created
func NewBoltDB(options bolt.Options, site BoltSite) (*BoltDB, error) {
	log.Printf("[INFO] bolt store for sites %+v, options %+v", site, options)
	result := BoltDB{db: &bolt.DB{}}

	db, err := bolt.Open(site.FileName, 0600, &options) //nolint:gocritic //octalLiteral is OK as FileMode
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make boltdb for %s", site.FileName)
	}

	// make top-level buckets
	topBuckets := []string{ingredientBucketName, cocktailBucketName, userBucketName}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bktName := range topBuckets {
			if _, e := tx.CreateBucketIfNotExists([]byte(bktName)); e != nil {
				return errors.Wrapf(e, "failed to create top level bucket %s", bktName)
			}
		}
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create top level bucket)")
	}

	result.db = db
	log.Printf("[DEBUG] bolt store created")
	return &result, nil
}

func (b *BoltDB) CreateIngredients(ingredient store.Ingredient) (ingredientID string, err error) {

	err = b.db.Update(func(tx *bolt.Tx) (err error) {
		// TODO: Update cockatail too
		ingredientsBkt := tx.Bucket([]byte(ingredientBucketName))

		// check if key already in store, reject doubles
		if ingredientsBkt.Get([]byte(ingredient.ID)) != nil {
			return errors.Errorf("key %s already in store", ingredient.ID)
		}

		// serialize comment to json []byte for bolt and save
		if err = b.save(ingredientsBkt, ingredient.ID, ingredient); err != nil {
			return errors.Wrapf(err, "failed to put key %s", ingredient.ID)
		}

		return nil
	})

	return ingredient.ID, err
}
func (b *BoltDB) UpdateIngredients(ingredient store.Ingredient) error {
	// b.GetIngredients(GetIngredientRequest{IngredientID: ingredient.ID})

	return b.db.Update(func(tx *bolt.Tx) error {
		ingredientsBkt := tx.Bucket([]byte(ingredientBucketName))
		return b.save(ingredientsBkt, ingredient.ID, ingredient)
	})

}

func (b *BoltDB) GetIngredient(ingredientID string) (store.Ingredient, error) {
	ingredient := store.Ingredient{}
	err := b.db.View(func(tx *bolt.Tx) error {
		ingredientsBkt := tx.Bucket([]byte(ingredientBucketName))
		return b.load(ingredientsBkt, ingredientID, &ingredient)
	})
	return ingredient, err
}

func (b *BoltDB) GetIngredients() ([]store.Ingredient, error) {
	var ingredients []store.Ingredient

	err := b.db.View(func(tx *bolt.Tx) error {
		var err error
		ingredientsBkt := tx.Bucket([]byte(ingredientBucketName))
		ingredients, err = b.loadAllIngredients(ingredientsBkt)
		return err
	})
	return ingredients, err
}

func (b *BoltDB) FindIngredients(ingredientID string) ([]store.Ingredient, error) {
	ingredients := []store.Ingredient{}

	return ingredients, nil
}

func (b *BoltDB) CreateCocktails(cocktail store.Cocktail) (cocktailID string, err error) {
	// TODO: fetch cocktails
	err = b.db.Update(func(tx *bolt.Tx) (err error) {
		var cocktailsBkt *bolt.Bucket

		// check if key already in store, reject doubles
		if cocktailsBkt.Get([]byte(cocktail.ID)) != nil {
			return errors.Errorf("key %s already in store", cocktail.ID)
		}

		// serialize cocktail to json []byte for bolt and save
		if err = b.save(cocktailsBkt, cocktail.ID, cocktail); err != nil {
			return errors.Wrapf(err, "failed to put key %s", cocktail.ID)
		}

		return nil
	})

	return cocktail.ID, err

}
func (b *BoltDB) UpdateCocktails(cocktail store.Cocktail) error {
	// b.GetCocktails(GetCocktailRequest{CocktailID: cocktail.ID})

	return b.db.Update(func(tx *bolt.Tx) error {
		ingredientsBkt := tx.Bucket([]byte(cocktailBucketName))
		return b.save(ingredientsBkt, cocktail.ID, cocktail)
	})
}

func (b *BoltDB) GetCocktails() ([]store.Cocktail, error) {
	return []store.Cocktail{}, nil
}

func (b *BoltDB) GetCocktail(cocktailID string) (store.Cocktail, error) {
	cocktail := store.Cocktail{}
	err := b.db.View(func(tx *bolt.Tx) error {
		cocktailsBkt := tx.Bucket([]byte(cocktailBucketName))
		return b.load(cocktailsBkt, cocktailID, &cocktail)
	})
	return cocktail, err
}

func (b *BoltDB) FindCocktails(search string) ([]store.Cocktail, error) {
	cocktails := []store.Cocktail{}

	return cocktails, nil
}

// Close boltdb store
func (b *BoltDB) Close() error {
	return errors.Wrapf(b.db.Close(), "can't close db")
}

// save marshaled value to key for bucket. Should run in update tx
func (b *BoltDB) save(bkt *bolt.Bucket, key string, value interface{}) (err error) {
	if value == nil {
		return errors.Errorf("can't save nil value for %s", key)
	}
	jdata, jerr := json.Marshal(value)
	if jerr != nil {
		return errors.Wrap(jerr, "can't marshal data")
	}
	if err = bkt.Put([]byte(key), jdata); err != nil {
		return errors.Wrapf(err, "failed to save key %s", key)
	}
	return nil
}

// load and unmarshal json value by key from bucket. Should run in view tx
func (b *BoltDB) load(bkt *bolt.Bucket, key string, res interface{}) error {
	value := bkt.Get([]byte(key))
	if value == nil {
		return errors.Errorf("no value for %s", key)
	}

	if err := json.Unmarshal(value, &res); err != nil {
		return errors.Wrap(err, "failed to unmarshal")
	}
	return nil
}

// load and unmarshal all json value from bucket. Should run in view tx
func (b *BoltDB) loadAllIngredients(bkt *bolt.Bucket) ([]store.Ingredient, error) {
	ingredients := []store.Ingredient{}
	err := bkt.ForEach(func(k, v []byte) error {
		ingredient := store.Ingredient{}

		if err := json.Unmarshal(v, &ingredient); err != nil {
			return errors.Wrap(err, "failed to unmarshal")
		}

		ingredients = append(ingredients, ingredient)
		return nil
	})

	return ingredients, err
}
