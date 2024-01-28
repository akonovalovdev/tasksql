package main

import (
	"github.com/jmoiron/sqlx"
)

type ParcelStore struct {
	db *sqlx.DB
}

func NewParcelStore(db *sqlx.DB) ParcelStore {
	return ParcelStore{db: db}
}

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	query := `INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)`
	res, err := s.db.NamedExec(query, p)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	query, args, err := sqlx.Named(`SELECT number, client, status, address, created_at FROM parcel WHERE number = :number`, map[string]interface{}{"number": number})
	if err != nil {
		return Parcel{}, err
	}
	p := Parcel{}
	err = s.db.Get(&p, query, args...)
	return p, err
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query, args, err := sqlx.Named(`SELECT number, client, status, address, created_at FROM parcel WHERE client = :client`, map[string]interface{}{"client": client})
	if err != nil {
		return nil, err
	}
	var parcels []Parcel
	err = s.db.Select(&parcels, query, args...)
	return parcels, err
}

func (s ParcelStore) SetStatus(number int, status string) error {
	query := `UPDATE parcel SET status = :status WHERE number = :number`
	_, err := s.db.NamedExec(query, map[string]interface{}{
		"status": status,
		"number": number,
	})
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	query := `UPDATE parcel SET address = :address WHERE number = :number AND status = :status`
	_, err := s.db.NamedExec(query, map[string]interface{}{
		"address": address,
		"number":  number,
		"status":  ParcelStatusRegistered,
	})
	return err
}

func (s ParcelStore) Delete(number int) error {
	query := `DELETE FROM parcel WHERE number = :number AND status = :status`
	_, err := s.db.NamedExec(query, map[string]interface{}{
		"number": number,
		"status": ParcelStatusRegistered,
	})
	return err
}
