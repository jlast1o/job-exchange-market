package auth

import "golang.org/x/crypto/bcrypt"

// PasswordHasher - контракт для операции с паролями. Сервис знает только этот интерфейс, без деталей его реализации.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) error
}

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{
		cost: cost,
	}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

func (h *BcryptHasher) Check(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
