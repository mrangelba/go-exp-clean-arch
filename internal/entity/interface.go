package entity

type OrderRepositoryInterface interface {
	Save(order *Order) (int, error)
	List() ([]Order, error)
}
