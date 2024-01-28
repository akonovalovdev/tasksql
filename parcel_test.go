package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	stmt, err := db.Prepare(`INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`)
	assert.NoError(t, err)
	res, err := stmt.Exec(parcel.Client, parcel.Status, parcel.Address, parcel.CreatedAt)
	assert.NoError(t, err)
	id, err := res.LastInsertId()
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// get
	stmt, err = db.Prepare(`SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`)
	assert.NoError(t, err)
	row := stmt.QueryRow(id)
	storedParcel := Parcel{}
	err = row.Scan(&storedParcel.Number, &storedParcel.Client, &storedParcel.Status, &storedParcel.Address, &storedParcel.CreatedAt)
	assert.NoError(t, err)
	assert.Equal(t, parcel.Client, storedParcel.Client)
	assert.Equal(t, parcel.Status, storedParcel.Status)
	assert.Equal(t, parcel.Address, storedParcel.Address)

	// delete
	stmt, err = db.Prepare(`DELETE FROM parcel WHERE number = ? AND status = ?`)
	assert.NoError(t, err)
	_, err = stmt.Exec(id, ParcelStatusRegistered)
	assert.NoError(t, err)

	// check
	stmt, err = db.Prepare(`SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`)
	assert.NoError(t, err)
	row = stmt.QueryRow(id)
	err = row.Scan(&storedParcel.Number, &storedParcel.Client, &storedParcel.Status, &storedParcel.Address, &storedParcel.CreatedAt)
	assert.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	assert.NoError(t, err)

	// check
	storedParcel, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, storedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	assert.NoError(t, err)

	// check
	storedParcel, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, newStatus, storedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, id)

	// get by client
	storedParcels, err := store.GetByClient(parcel.Client)
	assert.NoError(t, err)
	assert.Len(t, storedParcels, 1)

	// check
	assert.Equal(t, parcel.Client, storedParcels[0].Client)
	assert.Equal(t, parcel.Status, storedParcels[0].Status)
	assert.Equal(t, parcel.Address, storedParcels[0].Address)
}
