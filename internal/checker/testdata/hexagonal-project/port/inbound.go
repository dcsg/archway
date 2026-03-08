package port

type OrderService interface {
	CreateOrder(name string) error
	GetOrder(id string) (interface{}, error)
}
