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
	Logs   []string
	LogMu  sync.RWMutex
}

type BedrockUC interface {
	//world
	CreateServer(req *dto.ServerParams) error
	StopServer(name string) error
	StartServer(req *dto.StartServerReq) error
	DeleteWorld(user uint, name string) error
	EditWorld(req *dto.ServerParams, idWorld uint) error
	GetWorlds() ([]dto.GetWorlds, error)
	GetWorldAndPlayers(name string) (*dto.GetWorldAndPlayers, error)

	//command
	SendCommand(name string, command string) error

	//player
	KickPlayer(name string, playerName string) error
	BanPlayer(name string, playerName string) error
	CreateOrUpdatePermissions(req *dto.PermissionPlayer, worldName string) error
	DeletePermission(xuid, worldName string) error
	GetPermissionPlayer(name string) ([]dto.PermissionPlayer, error)
	GetServerLogs(name string) ([]string, error)

	//non import
	copyDir(src, dst string) error
	copyFile(src, dst string) error
	modifyProperties(req *dto.ServerParams) error
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
	src := "config/"
	dst := filepath.Join("data/servers", req.Name)

	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	if err := u.copyDir(src, dst); err != nil {
		return err
	}

	world, err := u.bedRepo.CreateWorld(req)
	if err != nil {
		return err
	}

	if err := u.modifyProperties(world); err != nil {
		return err
	}

	log.Printf("server %s is create with port %v", req.Name, req.Port)
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
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	combined := io.MultiReader(stdout, stderr)
	reader := io.TeeReader(combined, os.Stdout)

	if err := cmd.Start(); err != nil {
		return err
	}

	writer := bufio.NewWriter(stdin)

	u.s.Lock()
	u.servers[req.Name] = &BedrockServer{
		Cmd:    cmd,
		Writer: bufio.NewWriter(stdin),
		Port:   req.Port,
		Name:   dst,
	}
	u.s.Unlock()

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			u.HandleLogLine(line, writer, req.WorldId)

			u.s.RLock()
			server := u.servers[req.Name]
			u.s.RUnlock()

			if server != nil {
				server.LogMu.Lock()
				if len(server.Logs) > 1000 {
					server.Logs = server.Logs[1:]
				}
				server.Logs = append(server.Logs, line)
				server.LogMu.Unlock()
			}

		}
	}()

	log.Printf("Server %s (port: %d) is online", req.Name, req.Port)
	return nil
}

func (u *bedrockUC) HandleLogLine(line string, writer *bufio.Writer, worldId uint) {
	if strings.Contains(line, "Player connected:") {
		parts := strings.Split(line, "Player connected:")
		if len(parts) < 2 {
			return
		}
		data := strings.TrimSpace(parts[1])
		split := strings.Split(data, ", xuid: ")
		if len(split) < 2 {
			return
		}

		name := strings.TrimSpace(split[0])
		xuid := strings.TrimSpace(split[1])
		err := u.bedRepo.EnsurePlayerExists(xuid, name, worldId)
		if err != nil {
			return
		}
		return

	} else if strings.Contains(line, "Player Spawned:") {
		parts := strings.Split(line, "Player Spawned:")
		if len(parts) < 2 {
			return
		}
		data := strings.TrimSpace(parts[1])
		split := strings.Split(data, " xuid: ")
		if len(split) < 2 {
			return
		}

		name := strings.TrimSpace(split[0])
		xuid := strings.TrimSpace(strings.Split(split[1], ",")[0])
		role := u.bedRepo.GetPlayerRoleByName(xuid)
		if role == "" {
			role = "bocil"
		}
		roleUpper := strings.ToUpper(role)

		playerName := name
		if strings.Contains(name, " ") {
			playerName = fmt.Sprintf(`"%s"`, name)
		}

		cmds := []string{
			fmt.Sprintf(`tag %s add %s`, playerName, roleUpper),
			fmt.Sprintf(`scoreboard players set %s role 1`, playerName),
			fmt.Sprintf(`tellraw @a {"rawtext":[{"text":"[%s] %s joined the server"}]}`, roleUpper, name),
		}
		for _, cmd := range cmds {
			writer.WriteString(cmd + "\n")
			writer.Flush()
		}
		return

	} else if strings.Contains(line, "] [Chat]") {
		parts := strings.Split(line, "] [Chat]")
		if len(parts) < 2 {
			return
		}
		msg := strings.TrimSpace(parts[1])
		sep := strings.SplitN(msg, ": ", 2)
		if len(sep) < 2 {
			return
		}
		name := strings.TrimSpace(sep[0])
		message := sep[1]

		role := u.bedRepo.GetPlayerRoleByName(name)
		if role == "" {
			role = "bocil"
		}
		roleUpper := strings.ToUpper(role)

		tellraw := fmt.Sprintf(`tellraw @a {"rawtext":[{"text":"[%s] %s: %s"}]}`, roleUpper, name, message)
		writer.WriteString(tellraw + "\n")
		writer.Flush()

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

func (u *bedrockUC) SendCommand(name string, command string) error {
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

func (u *bedrockUC) DeleteWorld(user uint, name string) error {
	trgt := filepath.Join("data/servers", name)
	if err := os.RemoveAll(trgt); err != nil {
		return err
	}

	if err := u.bedRepo.DeleteWorld(user, name); err != nil {
		return err
	}

	fmt.Printf("%v menghapus server berna,ma %s", user, name)
	return nil
}

func (u *bedrockUC) EditWorld(req *dto.ServerParams, idWorld uint) error {
	if err := u.modifyProperties(req); err != nil {
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
		fmt.Println("Rel:", relPath)
		fmt.Println("Dst:", dstPath)

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

func (u *bedrockUC) modifyProperties(req *dto.ServerParams) error {
	dst := filepath.Join("data/servers", req.Name, "server.properties")
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

func (u *bedrockUC) KickPlayer(name string, playerName string) error {
	return u.SendCommand(name, "kick "+playerName)
}

func (u *bedrockUC) BanPlayer(name string, playerName string) error {
	return u.SendCommand(name, "ban "+playerName)
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
			newPermissions = append(newPermissions, dto.PermissionPlayer{
				Xuid:       r.Xuid,
				Permission: r.Permission,
			})
			break
		}
	}
	output, err := json.MarshalIndent(newPermissions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, output, 0644)

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
		return nil, fmt.Errorf("gagal lock/unlock async servers")
	}

	server.LogMu.RLock()
	logs := append([]string(nil), server.Logs...)
	server.LogMu.RUnlock()

	return logs, nil
}
