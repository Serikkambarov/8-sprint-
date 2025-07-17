package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	db, err :=sql.Open("sqlite","./tracker.db")  // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
    number, err := store.Add(parcel)
	require.NoError(t, err, "добавление посылки должно пройти без ошибок")
	require.NotEmpty(t, number, "номер посылки не должен быть пустым (0)")
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
 	got, err := store.Get(number)
	require.NoError(t, err, "получение посылки должно пройти без ошибок")
	require.Equal(t, parcel.Client, got.Client)
	require.Equal(t, parcel.Address, got.Address)
	require.Equal(t, parcel.Status, got.Status)
	require.Equal(t, number, got.Number)
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(number)
	require.NoError(t, err, "удаление должно пройти без ошибок")
	_, err = store.Get(number)
	require.Equal(t, sql.ErrNoRows, err, "посылка должна быть удалена") 
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite","./tracker.db")// настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	require.NoError(t, err, "добавление посылки должно пройти без ошибок")
	require.NotEmpty(t, number, "номер посылки не должен быть пустым (0)")

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.NoError(t, err, "добавление посылки должно пройти без ошибок")

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	got, err := store.Get(number)
	require.NoError(t, err, "ошибка при получении посылки")
	require.Equal(t, newAddress, got.Address, "адрес должен быть обновлён")
	
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite","./tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	require.NoError(t, err, "добавление посылки должно пройти без ошибок")
	require.NotEmpty(t, number, "номер посылки не должен быть пустым (0)")
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent
	err  = store.SetStatus(number, newStatus)
	require.NoError(t, err, "добавление посылки должно пройти без ошибок")

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	got, err := store.Get(number)
	assert.NoError(t, err, "ошибка при получении посылки")
	assert.Equal(t, newStatus, got.Status, "статус должен быть обновлён")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite","./tracker.db")// настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err, "добавление посылки должно пройти без ошибок")
		require.NotEmpty(t, id, "номер посылки не должен быть пустым (0)")

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)// получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err)
    assert.Len(t, storedParcels, len(parcels), "должно вернуться столько же посылок, сколько было добавлено")
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	

	// check
	for _, parcel := range storedParcels {
		 
		expected, ok := parcelMap[parcel.Number]// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		assert.True(t, ok, "посылка с номером %v не найдена в ожидаемом списке", parcel.Number)// убедитесь, что все посылки из storedParcels есть в parcelMap
		assert.Equal(t, expected, parcel)// убедитесь, что значения полей полученных посылок заполнены верно
		
	}
}
