package password

import "golang.org/x/crypto/bcrypt"

type Hasher struct {
	cost int
}

func NewHasher(cost int) *Hasher {
	return &Hasher{cost: cost}
}

func (h *Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *Hasher) Compare(hash, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}
	return true
}
