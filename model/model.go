package model

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

type WorldServer struct {
	ID                      uint   `gorm:"primaryKey"`
	CreatorId               *uint  `gorm:"index"`
	Port                    int    `gorm:"not null;unique"`
	Name                    string `gorm:"not null; unique"`
	GameMode                string `gorm:"default:survival"`
	Difficult               string `gorm:"default:normal"`
	AllowCheat              bool   `gorm:"default:true"`
	ViewDistance            int    `gorm:"default:32"`
	SeedWorld               string
	MaxPlayer               int    `gorm:"default:10"`
	DefaultPermissionPlayer string `gorm:"default:member"`

	//fk
	User       *User        `gorm:"foreignKey:CreatorId;constraint:OnDelete:SET NULL"`
	MemberRole []MemberRole `gorm:"foreignKey:WorldServerId;constraint:OnDelete:CASCADE"`
}
type MemberRole struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null"`
	Role string `gorm:"not null;default:bocil"`
	Xuid string

	WorldServerId uint `gorm:"index"`
}
