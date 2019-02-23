package identity

type Identifier interface{ Identifier() string }

func Has(k string, ids []Identifier) bool {
	_, has := Find(k, ids)
	return has
}

func Find(k string, ids []Identifier) (interface{}, bool) {
	for _, id := range ids {
		if ident := id.(Identifier); ident.Identifier() == k {
			return id, true
		}
	}
	return nil, false
}

func Remove(k string, ids []Identifier) bool {
	for i, id := range ids {
		if ident := id.(Identifier); ident.Identifier() == k {
			ids = append(ids[:i], ids[i+1:]...)
			return true
		}
	}
	return false
}

func Add(new Identifier, ids []Identifier) bool {
	if Has(new.Identifier(), ids) {
		return false
	}
	Set(new, ids)
	return true
}

func Set(new Identifier, ids []Identifier) {
	ids = append(ids, new)
}
