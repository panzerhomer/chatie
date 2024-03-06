package repository

import (
	"chatie/internal/models"
	"context"
	"database/sql"
)

type roomRepo struct {
	db *sql.DB
}

type RoomRepository interface {
	AddRoom(_ context.Context, room *models.Room) error
	GetRoomByName(_ context.Context, name string) (*models.Room, error)
}

func NewRoomRepository(db *sql.DB) RoomRepository {
	return &roomRepo{db: db}
}

func (repo *roomRepo) AddRoom(_ context.Context, room *models.Room) error {
	query, err := repo.db.Prepare("INSERT INTO room(id, name, private) values(?,?,?)")
	if err != nil {
		return err
	}

	_, err = query.Exec(room.GetId(), room.GetName(), room.GetPrivate())
	if err != nil {
		return err
	}
	return nil
}

func (repo *roomRepo) GetRoomByName(_ context.Context, name string) (*models.Room, error) {
	query := repo.db.QueryRow("SELECT id, name, private FROM room where name = ? LIMIT 1", name)

	var room models.Room

	if err := query.Scan(&room.ID, &room.Name, &room.Private); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &room, nil

}
