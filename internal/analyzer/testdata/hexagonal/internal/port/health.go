package port

type HealthPort interface {
	Health() string
}
