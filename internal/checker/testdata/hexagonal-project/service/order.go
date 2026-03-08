package service

type OrderService struct{}

func (s *OrderService) CreateOrder(name string) error {
	// Swallowed error — should trigger swallowed_error.
	var err error
	if err != nil {
	}
	return nil
}

func (s *OrderService) GetOrder(id string) (interface{}, error) {
	return nil, nil
}
