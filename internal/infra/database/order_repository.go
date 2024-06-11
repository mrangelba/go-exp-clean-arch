package database

import (
	"database/sql"

	"github.com/mrangelba/go-exp-clean-arch/internal/entity"
)

type OrderRepository struct {
	Db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{Db: db}
}

// inserr order and return id
func (r *OrderRepository) Save(order *entity.Order) (int, error) {
	stmt, err := r.Db.Prepare("INSERT INTO orders (price, tax, final_price) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(order.Price, order.Tax, order.FinalPrice)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (r *OrderRepository) GetTotal() (int, error) {
	var total int
	err := r.Db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *OrderRepository) List() ([]entity.Order, error) {
	var orders []entity.Order

	stmt, err := r.Db.Prepare("SELECT * FROM orders")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order entity.Order
		err := rows.Scan(
			&order.ID,
			&order.Price,
			&order.Tax,
			&order.FinalPrice,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
