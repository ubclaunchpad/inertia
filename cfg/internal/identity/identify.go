package identity

// Identifier wraps classes with Identifier()
type Identifier interface{ Identifier() string }

// Has returns true if k exists in given ids
func Has(k string, ids []Identifier) bool {
	_, has := Get(k, ids)
	return has
}

// Get finds and returns the value and true if k exists in given ids
func Get(k string, ids []Identifier) (interface{}, bool) {
	for _, id := range ids {
		if ident := id.(Identifier); ident.Identifier() == k {
			return id, true
		}
	}
	return nil, false
}

// Remove deletes identifier with name k in given ids
func Remove(k string, ids *[]Identifier) bool {
	idv := *ids
	for i, id := range idv {
		if id.Identifier() == k {
			idv = append(idv[:i], idv[i+1:]...)
			*ids = idv
			return true
		}
	}
	return false
}

// Add inserts new into given ids
func Add(new Identifier, ids *[]Identifier) bool {
	if Has(new.Identifier(), *ids) {
		return false
	}
	idv := *ids
	idv = append(idv, new)
	*ids = idv
	return true
}

// Set updates an existing entry with the same name as k, or just adds it
func Set(new Identifier, ids *[]Identifier) {
	Remove(new.Identifier(), ids)
	idv := *ids
	idv = append(idv, new)
	*ids = idv
}
