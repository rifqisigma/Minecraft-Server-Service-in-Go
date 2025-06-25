package usecase

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"minecrat_go/dto"
	"minecrat_go/internal/repository"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type BedrockServer struct {
	Cmd    *exec.Cmd
	Writer *bufio.Writer
	Port   int
	Name   string
	Id     uint
	Logs   []string
	LogMu  sync.RWMutex
}

type BedrockUC interface {
	//world
	CreateServer(req *dto.ServerParams) error
	StopServer(name string) error
	StartServer(req *dto.StartServerReq) error
	DeleteWorld(user uint, name string) error
	EditWorld(req *dto.ServerParams, idWorld uint, nameOld string) error
	GetWorlds() ([]dto.GetWorlds, error)
	GetWorldAndPlayers(name string) (*dto.GetWorldAndPlayers, error)
	SendCommandforAPI(name string, command string) error
	CreatePriority(req *dto.Allowlist, worldName string) error
	DeletePriority(xuid, worldName string) error

	//player
	KickPlayer(name string, playerName string) error
	BanPlayer(name string, playerName string) error
	CreateOrUpdatePermissions(req *dto.PermissionPlayer, worldName string) error
	DeletePermission(xuid, worldName string) error
	GetPermissionPlayer(name string) ([]dto.PermissionPlayer, error)
	GetServerLogs(name string) ([]string, error)
	GetPriority(name string) ([]dto.Allowlist, error)

	//non import
	handleLogLine(line string, worldId uint)
	copyDir(src, dst string) error
	copyFile(src, dst string) error
	modifyProperties(req *dto.ServerParams, worldname string) error
}

type bedrockUC struct {
	servers map[string]*BedrockServer
	s       sync.RWMutex
	bedRepo repository.BedrockRepo
}

func NewBedrockUC(bedRepo repository.BedrockRepo) BedrockUC {
	return &bedrockUC{
		servers: make(map[string]*BedrockServer),
		s:       sync.RWMutex{},
		bedRepo: bedRepo,
	}

}

func (u *bedrockUC) CreateServer(req *dto.ServerParams) error {
	src := "config/world_template/"
	dst := filepath.Join("data/servers", req.Name)

	worlddb, err := u.bedRepo.CreateWorld(req)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("remove old server failed: %w", err)
	}

	if err := u.copyDir(src, dst); err != nil {
		return fmt.Errorf("copy template failed: %w", err)
	}

	if err := u.modifyProperties(worlddb, worlddb.Name); err != nil {
		return fmt.Errorf("modify properties failed: %w", err)
	}

	log.Printf("Server %s created on port %d", req.Name, req.Port)
	return nil
}

func (u *bedrockUC) StartServer(req *dto.StartServerReq) error {
	dst := filepath.Join("data/servers", req.Name)

	cmd := exec.Command("./bedrock_server")
	cmd.Dir = dst

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return err
	}

	combined := io.MultiReader(stdout, stderr)
	reader := io.TeeReader(combined, os.Stdout)

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return err
	}

	server := &BedrockServer{
		Cmd:    cmd,
		Writer: bufio.NewWriter(stdin),
		Port:   req.Port,
		Name:   dst,
		Id:     req.WorldId,
		Logs:   make([]string, 0, 1001),
	}

	u.s.Lock()
	u.servers[req.Name] = server
	u.s.Unlock()

	go func() {
		defer func() {
			stdin.Close()
			stdout.Close()
			stderr.Close()
		}()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()

			u.handleLogLine(line, req.WorldId)

			server.LogMu.Lock()
			if len(server.Logs) > 1000 {
				server.Logs = server.Logs[1:]
			}
			server.Logs = append(server.Logs, line)
			server.LogMu.Unlock()
		}
	}()

	log.Printf("Server %s (port: %d) is online", req.Name, req.Port)
	return nil
}

func (u *bedrockUC) handleLogLine(line string, worldId uint) {

	if strings.Contains(line, "Player connected:") {
		log.Println("add player")
		parts := strings.Split(line, "Player connected:")
		if len(parts) < 2 {
			return
		}
		data := strings.TrimSpace(parts[1])
		split := strings.Split(data, ", xuid: ")
		if len(split) < 2 {
			return
		}

		xuid := strings.TrimSpace(split[1])
		err := u.bedRepo.EnsurePlayerExists(xuid, worldId)
		if err != nil {
			return
		}
		return

	}
}

func (u *bedrockUC) StopServer(name string) error {
	server, ok := u.servers[name]
	if !ok {
		return fmt.Errorf("server not found in map")
	}

	err := server.Cmd.Process.Kill()
	if err != nil {
		return err
	}

	u.s.Lock()
	delete(u.servers, name)
	u.s.Unlock()

	log.Printf("server %s is stop", name)
	return nil
}

func (u *bedrockUC) DeleteWorld(user uint, name string) error {
	trgt := filepath.Join("data/servers", name)
	if err := os.RemoveAll(trgt); err != nil {
		return err
	}

	if err := u.bedRepo.DeleteWorld(user, name); err != nil {
		return err
	}

	fmt.Printf("%v menghapus server bernama %s", user, name)
	return nil
}

func (u *bedrockUC) EditWorld(req *dto.ServerParams, idWorld uint, nameOld string) error {
	if err := u.modifyProperties(req, nameOld); err != nil {
		return err
	}
	if err := u.bedRepo.EditWorld(req, idWorld); err != nil {
		return err
	}
	return nil
}

func (u *bedrockUC) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		fmt.Println("Scanning:", path)
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)
		log.Println("Rel:", relPath)
		log.Println("Dst:", dstPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			return u.copyFile(path, dstPath)
		}
	})
}

func (u *bedrockUC) copyFile(src, dst string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}

	defer from.Close()

	to, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}

func (u *bedrockUC) modifyProperties(req *dto.ServerParams, worldname string) error {
	dst := filepath.Join("data/servers", worldname, "server.properties")
	input, err := os.ReadFile(dst)
	if err != nil {
		return err
	}

	content := string(input)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		//name
		if req.Name != "" && strings.HasPrefix(line, "server-name=") {
			lines[i] = "server-name=" + req.Name
		}

		//name
		if req.Name != "" && strings.HasPrefix(line, "level-name=") {
			lines[i] = "level-name=" + req.Name
		}
		//game mode
		if req.GameMode != "" && strings.HasPrefix(line, "gamemode=") {
			lines[i] = "gamemode=" + req.GameMode
		}
		//difficult
		if req.Difficult != "" && strings.HasPrefix(line, "difficulty=") {
			lines[i] = "difficulty=" + req.Difficult
		}
		//max player
		if req.MaxPlayer != 0 && strings.HasPrefix(line, "max-players=") {
			lines[i] = "max-players=" + strconv.Itoa(req.MaxPlayer)
		}
		//allow cheats
		if strings.HasPrefix(line, "allow-cheats=") {
			lines[i] = "allow-cheats=" + strconv.FormatBool(req.AllowCheat)
		}
		//seed
		if req.SeedWorld != "" && strings.HasPrefix(line, "level-seed=") {
			lines[i] = "level-seed=" + req.SeedWorld
		}
		//permission player
		if req.DefaultPermissionPlayer != "" && strings.HasPrefix(line, "default-player-permission-level=") {
			lines[i] = "default-player-permission-level=" + req.DefaultPermissionPlayer
		}
		//view distance
		if req.ViewDistance != 0 && strings.HasPrefix(line, "view-distance=") {
			lines[i] = "view-distance=" + strconv.Itoa(req.ViewDistance)
		}
		//port
		if req.Port != 0 && strings.HasPrefix(line, "server-port=") {
			lines[i] = "server-port=" + strconv.Itoa(req.Port)
		}
		if req.Port != 0 && strings.HasPrefix(line, "server-portv6=") {
			lines[i] = "server-portv6=" + strconv.Itoa(req.Port)
		}
	}

	output := strings.Join(lines, "\n")
	return os.WriteFile(dst, []byte(output), 0644)
}

func (u *bedrockUC) SendCommandforAPI(name string, command string) error {
	u.s.RLock()
	server, ok := u.servers[name]
	u.s.RUnlock()
	if !ok {
		return fmt.Errorf("server %s not found", name)
	}

	if server.Writer == nil {
		return fmt.Errorf("writer not initialized for server %s", name)
	}

	_, err := server.Writer.WriteString(command + "\n")
	if err != nil {
		return err
	}
	return server.Writer.Flush()
}

func (u *bedrockUC) KickPlayer(name string, playerName string) error {
	kick := fmt.Sprintf(`kick "%s"`, playerName)
	if err := u.SendCommandforAPI(name, kick); err != nil {
		return err
	}
	return nil
}

func (u *bedrockUC) BanPlayer(name string, playerName string) error {
	ban := fmt.Sprintf(`ban "%s"`, playerName)
	if err := u.SendCommandforAPI(name, ban); err != nil {
		return err
	}
	return nil
}

func (u *bedrockUC) CreateOrUpdatePermissions(req *dto.PermissionPlayer, worldName string) error {
	var resultFile []dto.PermissionPlayer

	path := filepath.Join("data/servers", worldName, "permissions.json")
	result, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(result, &resultFile); err != nil {
		return err
	}

	found := false

	for i, r := range resultFile {
		if r.Xuid == req.Xuid {
			resultFile[i].Permission = req.Permission
			found = true
			break
		}
	}

	if !found {
		resultFile = append(resultFile, dto.PermissionPlayer{
			Xuid:       req.Xuid,
			Permission: req.Permission,
		})
	}

	output, err := json.MarshalIndent(resultFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, output, 0644)

}

func (u *bedrockUC) CreatePriority(req *dto.Allowlist, worldName string) error {
	var resultFile []dto.Allowlist

	path := filepath.Join("data/servers", worldName, "allowlist.json")
	result, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(result, &resultFile); err != nil {
		return err
	}

	found := false

	for i, r := range resultFile {
		if r.Xuid == req.Xuid {
			resultFile[i].Priority = req.Priority
			found = true
			break
		}
	}

	if !found {
		resultFile = append(resultFile, dto.Allowlist{
			Xuid:     req.Xuid,
			Name:     req.Name,
			Priority: req.Priority,
		})
	}

	output, err := json.MarshalIndent(resultFile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, output, 0644)

}

func (u *bedrockUC) DeletePriority(xuid, worldName string) error {
	var resultFile []dto.Allowlist

	path := filepath.Join("data/servers", worldName, "allowlist.json")
	result, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(result, &resultFile); err != nil {
		return err
	}

	var newPriority []dto.Allowlist
	for _, r := range resultFile {
		if r.Xuid != xuid {
			continue
		}
		newPriority = append(newPriority, dto.Allowlist{
			Xuid:     r.Xuid,
			Name:     r.Name,
			Priority: r.Priority,
		})
	}

	output, err := json.MarshalIndent(newPriority, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, output, 0644)

}

func (u *bedrockUC) DeletePermission(xuid, worldName string) error {
	var resultFile []dto.PermissionPlayer

	path := filepath.Join("data/servers", worldName, "permissions.json")
	result, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(result, &resultFile); err != nil {
		return err
	}

	var newPermissions []dto.PermissionPlayer
	for _, r := range resultFile {
		if r.Xuid != xuid {
			continue
		}
		newPermissions = append(newPermissions, dto.PermissionPlayer{
			Xuid:       r.Xuid,
			Permission: r.Permission,
		})
	}
	output, err := json.MarshalIndent(newPermissions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, output, 0644)

}

func (u *bedrockUC) GetPriority(name string) ([]dto.Allowlist, error) {
	var resultFile []dto.Allowlist

	path := filepath.Join("data/servers", name, "allowlist.json")
	result, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(result, &resultFile); err != nil {
		return nil, err
	}

	return resultFile, nil

}

func (u *bedrockUC) GetPermissionPlayer(name string) ([]dto.PermissionPlayer, error) {
	var resultFile []dto.PermissionPlayer

	path := filepath.Join("data/servers", name, "permissions.json")
	result, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(result, &resultFile); err != nil {
		return nil, err
	}

	return resultFile, nil

}

func (u *bedrockUC) GetWorlds() ([]dto.GetWorlds, error) {
	return u.bedRepo.GetWorlds()
}

func (u *bedrockUC) GetWorldAndPlayers(name string) (*dto.GetWorldAndPlayers, error) {
	return u.bedRepo.GetWorldAndPlayers(name)
}

func (u *bedrockUC) GetServerLogs(name string) ([]string, error) {

	u.s.RLock()
	server, ok := u.servers[name]
	u.s.RUnlock()

	if !ok {
		return nil, fmt.Errorf("gagal lock/unlock async server")
	}

	server.LogMu.RLock()
	logs := append([]string(nil), server.Logs...)
	server.LogMu.RUnlock()

	return logs, nil
}
