package identity

type Identifier interface{ Identifier() string }

func Has(k string, ids []Identifier) bool {
	_, has := Get(k, ids)
	return has
}

func Get(k string, ids []Identifier) (interface{}, bool) {
	for _, id := range ids {
		if ident := id.(Identifier); ident.Identifier() == k {
			return id, true
		}
	}
	return nil, false
}

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

func Add(new Identifier, ids *[]Identifier) bool {
	if Has(new.Identifier(), *ids) {
		return false
	}
	idv := *ids
	idv = append(idv, new)
	*ids = idv
	return true
}

func Set(new Identifier, ids *[]Identifier) {
	Remove(new.Identifier(), ids)
	idv := *ids
	idv = append(idv, new)
	*ids = idv
}
