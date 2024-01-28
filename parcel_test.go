package main

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestAddGetDelete(t *testing.T) {
	db, err := sqlx.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close the database: %v", err)
		}
	}()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel.Client, storedParcel.Client)
	assert.Equal(t, parcel.Status, storedParcel.Status)
	assert.Equal(t, parcel.Address, storedParcel.Address)
	assert.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db, err := sqlx.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close the database: %v", err)
		}
	}()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, storedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sqlx.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close the database: %v", err)
		}
	}()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newStatus, storedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sqlx.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close the database: %v", err)
		}
	}()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)
		parcels[i].Number = id
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Equal(t, len(parcels), len(storedParcels))

	for _, storedParcel := range storedParcels {
		for _, parcel := range parcels {
			if parcel.Number == storedParcel.Number {
				assert.Equal(t, parcel.Client, storedParcel.Client)
				assert.Equal(t, parcel.Status, storedParcel.Status)
				assert.Equal(t, parcel.Address, storedParcel.Address)
				assert.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)
			}
		}
	}
}

func TestRegister(t *testing.T) {
	db, err := sqlx.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close the database: %v", err)
		}
	}()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	parcel, err := service.Register(client, address)
	require.NoError(t, err)

	assert.Equal(t, client, parcel.Client)
	assert.Equal(t, address, parcel.Address)
	assert.Equal(t, ParcelStatusRegistered, parcel.Status)
}

func TestChangeAddress(t *testing.T) {
	db, err := sqlx.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close the database: %v", err)
		}
	}()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newAddress := "new test address"
	err = service.ChangeAddress(id, newAddress)
	require.NoError(t, err)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, storedParcel.Address)
}

func TestNextStatus(t *testing.T) {
	db, err := sqlx.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close the database: %v", err)
		}
	}()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	err = service.NextStatus(id)
	require.NoError(t, err)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, storedParcel.Status)
}
