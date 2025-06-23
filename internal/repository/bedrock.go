package repository

import (
	"errors"
	"minecrat_go/dto"
	"minecrat_go/helper/utils"
	"minecrat_go/model"

	"gorm.io/gorm"
)

type BedrockRepo interface {
	CreateWorld(req *dto.ServerParams) (*dto.ServerParams, error)
	EditWorld(req *dto.ServerParams, idWorld uint) error
	DeleteWorld(user uint, name string) error
	GetWorlds() ([]dto.GetWorlds, error)
	GetWorldAndPlayers(name string) (*dto.GetWorldAndPlayers, error)
	EnsurePlayerExists(xuid, name string, worldId uint) error
	GetPlayerRoleByName(name string) string
}

type bedrockRepo struct {
	db *gorm.DB
}

func NewBedrockRepo(db *gorm.DB) BedrockRepo {
	return &bedrockRepo{db}
}

func (r *bedrockRepo) CreateWorld(req *dto.ServerParams) (*dto.ServerParams, error) {
	newWorld := model.WorldServer{
		CreatorId:               &req.Creator,
		Name:                    req.Name,
		Port:                    req.Port,
		GameMode:                req.GameMode,
		Difficult:               req.Difficult,
		AllowCheat:              req.AllowCheat,
		MaxPlayer:               req.MaxPlayer,
		DefaultPermissionPlayer: req.DefaultPermissionPlayer,
		SeedWorld:               req.SeedWorld,
		ViewDistance:            req.ViewDistance,
	}

	if err := r.db.Debug().Model(&model.WorldServer{}).Create(&newWorld).Error; err != nil {
		return nil, err
	}
	return &dto.ServerParams{
		Name:                    newWorld.Name,
		Port:                    newWorld.Port,
		GameMode:                newWorld.GameMode,
		Difficult:               newWorld.Difficult,
		AllowCheat:              newWorld.AllowCheat,
		MaxPlayer:               newWorld.MaxPlayer,
		DefaultPermissionPlayer: newWorld.DefaultPermissionPlayer,
		SeedWorld:               newWorld.SeedWorld,
		ViewDistance:            req.ViewDistance,
	}, nil
}

func (r *bedrockRepo) EditWorld(req *dto.ServerParams, idWorld uint) error {
	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Port != 0 {
		updates["port"] = req.Port
	}
	if req.GameMode != "" {
		updates["game_mode"] = req.GameMode
	}
	if req.Difficult != "" {
		updates["difficult"] = req.Difficult
	}
	if req.AllowCheat {
		updates["allow_cheat"] = req.AllowCheat
	}
	if req.MaxPlayer != 0 {
		updates["max_player"] = req.MaxPlayer
	}
	if req.DefaultPermissionPlayer != "" {
		updates["default_permission_player"] = req.DefaultPermissionPlayer
	}
	if req.SeedWorld != "" {
		updates["seed_world"] = req.SeedWorld
	}
	if req.ViewDistance != 0 {
		updates["view_distance"] = req.ViewDistance
	}

	if len(updates) == 0 {
		return nil
	}
	err := r.db.Debug().Model(&model.WorldServer{}).Where("id = ? AND creator_id = ?", idWorld, req.Creator).Updates(&updates).Error
	if err != nil {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.ErrNotCreator
	}
	return nil
}

func (r *bedrockRepo) DeleteWorld(creator uint, name string) error {
	err := r.db.Debug().Model(&model.WorldServer{}).Where("name = ? AND creator_id = ?", name, creator).Delete(&model.WorldServer{}).Error
	if err != nil {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.ErrNotCreator
	}

	return nil
}

func (u *bedrockRepo) EnsurePlayerExists(xuid, name string, worldId uint) error {
	err := u.db.Where("xuid = ?", xuid).Error
	if err != nil {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := u.db.Create(&model.MemberRole{
			Xuid: xuid,
			Name: name,
			Role: "bocil",
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *bedrockRepo) GetPlayerRoleByName(name string) string {
	var role string
	err := r.db.Where("xuid = ?", name).Pluck("role", role)
	if err != nil {
		return ""
	}
	return role
}

func (r *bedrockRepo) GetWorlds() ([]dto.GetWorlds, error) {
	var result []model.WorldServer
	if err := r.db.Model(&model.WorldServer{}).Preload("MemberRole").Find(&result).Error; err != nil {
		return nil, err
	}

	response := make([]dto.GetWorlds, len(result))

	for _, r := range result {
		response = append(response, dto.GetWorlds{
			Creator: r.User.Username,
			Name:    r.Name,
			Port:    r.Port,
			Players: len(r.MemberRole),
		})

	}

	return response, nil

}

func (r *bedrockRepo) GetWorldAndPlayers(name string) (*dto.GetWorldAndPlayers, error) {
	var result model.WorldServer
	if err := r.db.Model(&model.WorldServer{}).Where("name = ?", name).Preload("MemberRole").First(&result).Error; err != nil {
		return nil, err
	}

	var responsePlayers []dto.Player
	for _, r := range result.MemberRole {
		responsePlayers = append(responsePlayers, dto.Player{
			Name: r.Name,
			Role: r.Role,
		})
	}

	return &dto.GetWorldAndPlayers{
		Creator:                 result.User.Username,
		Name:                    result.Name,
		Port:                    result.Port,
		Difficult:               result.Difficult,
		GameMode:                result.GameMode,
		MaxPlayer:               result.MaxPlayer,
		ViewDistance:            result.ViewDistance,
		AllowCheat:              result.AllowCheat,
		SeedWorld:               result.SeedWorld,
		DefaultPermissionPlayer: result.DefaultPermissionPlayer,
		Players:                 responsePlayers,
	}, nil

}
