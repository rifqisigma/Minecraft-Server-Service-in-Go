package utils

import (
	"fmt"
	"minecrat_go/dto"
	"regexp"
)

func IsValidEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

func ValidateReq(req *dto.ServerParams) error {
	if req.GameMode != "survival" && req.GameMode != "creative" && req.GameMode != "adventure" {
		return fmt.Errorf("gamemode salah")
	}
	if req.Difficult != "easy" && req.Difficult != "normal" && req.Difficult != "hard" {
		return fmt.Errorf("difficult salah")
	}

	if req.DefaultPermissionPlayer != "visitor" && req.DefaultPermissionPlayer != "member" && req.DefaultPermissionPlayer != "operator" {
		return fmt.Errorf("gamemode permission")
	}
	return nil
}
