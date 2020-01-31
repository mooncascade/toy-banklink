package dao

import (
	"database/sql"

	"github.com/google/uuid"
)

type paymentDAO struct {
	db *sql.DB
}

// GetPaymentDataResponse used to respond with received payment data
type GetPaymentDataResponse struct {
	UUID               string `json:"uuid"`
	ReceiverID         string `json:"receiver_id"`
	Amount             int    `json:"amount"`
	Status             string `json:"status"`
	TrueLayerPaymentID string `json:"truelayer_payment_id"`
}

// PaymentDAO interface defining DAO
type PaymentDAO interface {
	CreateTable() error
	MapToTruelayer(uuid, truelayerID string) error
	UpdatePaymentStatus(truelayerID, status string) error
	InsertPayment(receiverID string, amount int) (string, error)
	GetPayment(uuid string) (GetPaymentDataResponse, error)
	GetPaymentByTruelayerID(truelayerID string) (GetPaymentDataResponse, error)
}

//NewDAO Creates a new database access object and establishes connection with the database
func NewDAO() (PaymentDAO, error) {
	database, err := sql.Open("sqlite3", "./payments.db")
	if err != nil {
		return paymentDAO{}, err
	}

	return paymentDAO{
		db: database,
	}, nil
}

// CreateTable creates the payments table to the database
func (dao paymentDAO) CreateTable() error {
	_, err := dao.db.Exec(`CREATE TABLE IF NOT EXISTS payments (
		id string PRIMARY KEY NOT NULL,
		receiver_id TEXT NOT NULL,
		amount INTEGER NOT NULL,
		status TEXT DEFAULT unpaid,
		truelayer_payment_id TEXT DEFAULT '')`)
	if err != nil {
		return err
	}

	return nil
}

// MapToTruelayer adds truelayer payment ID to payment data
func (dao paymentDAO) MapToTruelayer(uuid, truelayerID string) error {
	_, err := dao.db.Exec("UPDATE payments SET truelayer_payment_id = ? WHERE id = ?", truelayerID, uuid)

	return err
}

// UpdatePaymentStatus updates payment status
func (dao paymentDAO) UpdatePaymentStatus(truelayerID, status string) error {
	_, err := dao.db.Exec("UPDATE payments SET status = ? WHERE truelayer_payment_id = ?", status, truelayerID)

	return err
}

// InsertPayment Inserts a new payment to database
func (dao paymentDAO) InsertPayment(receiverID string, amount int) (string, error) {
	uuidValue := uuid.New()

	_, err := dao.db.Exec(
		"INSERT INTO payments (id, receiver_id, amount) VALUES (?, ?, ?)",
		uuidValue.String(), receiverID, amount)
	if err != nil {
		return "", err
	}

	return uuidValue.String(), nil
}

// GetPayment gets payment data by payment UUID
func (dao paymentDAO) GetPayment(uuid string) (GetPaymentDataResponse, error) {
	var response GetPaymentDataResponse

	statement, err := dao.db.Prepare(`
		SELECT id, receiver_id, amount, status, truelayer_payment_id FROM payments WHERE id = ?`)
	if err != nil {
		return response, err
	}

	err = statement.QueryRow(uuid).Scan(
		&response.UUID, &response.ReceiverID,
		&response.Amount, &response.Status,
		&response.TrueLayerPaymentID)
	if err != nil {
		return response, err
	}

	return response, nil
}

// GetPaymentByTruelayerID gets payment data by truelayer ID
func (dao paymentDAO) GetPaymentByTruelayerID(truelayerID string) (GetPaymentDataResponse, error) {
	var response GetPaymentDataResponse

	statement, err := dao.db.Prepare(`
	SELECT id, receiver_id, amount, status, truelayer_payment_id FROM payments WHERE truelayer_payment_id = ?`)
	if err != nil {
		return response, err
	}

	err = statement.QueryRow(truelayerID).Scan(
		&response.UUID, &response.ReceiverID,
		&response.Amount, &response.Status,
		&response.TrueLayerPaymentID)
	if err != nil {
		return response, err
	}

	return response, nil
}
