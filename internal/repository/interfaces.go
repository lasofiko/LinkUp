package repository
import "github.com/lasofiko/LinkUp/internal/domain"

type UserRepository interface {
	Create(user domain.User) error
	GetByID( id int) (domain.User, error)
	GetAll() ([]domain.User, error)
}