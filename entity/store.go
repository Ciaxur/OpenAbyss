package entity

type EntityStore struct {
	Keys   map[string]Entity
	Length int64
}

/**
 * Adds a key to the entity store
 */
func (entityStore *EntityStore) Add(elt Entity) {
	entityStore.Keys[elt.Name] = elt
	entityStore.Length++
}

/**
 * Return entity given entity's name
 */
func (entityStore *EntityStore) Get(keyName string) *Entity {
	entity := entityStore.Keys[keyName]
	return &entity
}

/**
 * Checks if the given keyname exists in store
 */
func (entityStore *EntityStore) Has(keyName string) bool {
	_, ok := entityStore.Keys[keyName]
	return ok
}
