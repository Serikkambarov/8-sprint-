package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec(`
	INSERT INTO parcel(client, address, status, created_at)
	VALUES (?, ?, ?, ?)`, p.Client, p.Address, p.Status, p.CreatedAt)

	if err != nil {
		return 0, err // ошибка при вставке
	}
	id, err := res.LastInsertId() // получаем ID новой посылки
	return int(id), err           // возвращаем ID (трек-номер)
}

// верните идентификатор последней добавленной записи

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	row := s.db.QueryRow(`SELECT number, client, address, status, created_at FROM parcel WHERE number = ?`, number)

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Address, &p.Status, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err

	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query(`SELECT number, client, address, status, created_at FROM parcel WHERE client = ?`, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// заполните срез Parcel данными из таблицы
	var res []Parcel
	for rows.Next() {
		var r Parcel
		err := rows.Scan(&r.Number, &r.Client, &r.Address, &r.Status, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec(`UPDATE parcel SET status = ? WHERE number  = ?`, status, number)
	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Обновляем адрес только если статус — registered
	res, err := s.db.Exec(`
		UPDATE parcel 
		SET address = ? 
		WHERE number = ? AND status = ?`, address, number, ParcelStatusRegistered)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("нельзя изменить адрес после отправки")
	}
	return nil
}


func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	res, err := s.db.Exec(`
		DELETE FROM parcel WHERE number  = ? AND status = ?`, number, ParcelStatusRegistered )
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("нельзя удалить посылку — либо она не найдена, либо уже отправлена")
	}
	return nil	
	

}
