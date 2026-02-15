package domain

type Service interface {
	Health() string
}
