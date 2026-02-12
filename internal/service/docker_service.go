package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"godockerui/internal/model"
)

type RuntimeService interface {
	Dashboard(ctx context.Context) (model.Dashboard, error)
	ListContainers(ctx context.Context) ([]model.Container, error)
	ListImages(ctx context.Context) ([]model.Image, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RestartContainer(ctx context.Context, id string) error
}

type DockerService struct {
	provider provider
}

type provider interface {
	dashboard(context.Context) (model.Dashboard, error)
	listContainers(context.Context) ([]model.Container, error)
	listImages(context.Context) ([]model.Image, error)
	startContainer(context.Context, string) error
	stopContainer(context.Context, string) error
	restartContainer(context.Context, string) error
}

func NewDockerService(dockerBin string, dockerErr error) *DockerService {
	if dockerErr != nil {
		return &DockerService{provider: newSimulatedProvider()}
	}

	p := &cliProvider{dockerBin: dockerBin}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := p.raw(ctx, "version", "--format", "{{.Server.Version}}"); err != nil {
		return &DockerService{provider: newSimulatedProvider()}
	}
	return &DockerService{provider: p}
}

func (s *DockerService) Dashboard(ctx context.Context) (model.Dashboard, error) {
	return s.provider.dashboard(ctx)
}

func (s *DockerService) ListContainers(ctx context.Context) ([]model.Container, error) {
	return s.provider.listContainers(ctx)
}

func (s *DockerService) ListImages(ctx context.Context) ([]model.Image, error) {
	return s.provider.listImages(ctx)
}

func (s *DockerService) StartContainer(ctx context.Context, id string) error {
	return s.provider.startContainer(ctx, id)
}

func (s *DockerService) StopContainer(ctx context.Context, id string) error {
	return s.provider.stopContainer(ctx, id)
}

func (s *DockerService) RestartContainer(ctx context.Context, id string) error {
	return s.provider.restartContainer(ctx, id)
}

type cliProvider struct {
	dockerBin string
}

func (p *cliProvider) dashboard(ctx context.Context) (model.Dashboard, error) {
	containers, err := p.listContainers(ctx)
	if err != nil {
		return model.Dashboard{}, err
	}
	images, err := p.listImages(ctx)
	if err != nil {
		return model.Dashboard{}, err
	}

	up := 0
	for _, c := range containers {
		if c.State == "running" {
			up++
		}
	}

	return model.Dashboard{
		ContainersTotal: len(containers),
		ContainersUp:    up,
		ContainersDown:  len(containers) - up,
		ImagesTotal:     len(images),
		Simulated:       false,
	}, nil
}

func (p *cliProvider) listContainers(ctx context.Context) ([]model.Container, error) {
	out, err := p.output(ctx, "ps", "-a", "--no-trunc", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []model.Container{}, nil
	}

	result := make([]model.Container, 0, len(lines))
	for _, line := range lines {
		var item struct {
			ID      string `json:"ID"`
			Names   string `json:"Names"`
			Image   string `json:"Image"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Command string `json:"Command"`
		}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("parse container output: %w", err)
		}
		result = append(result, model.Container{
			ID:      shortID(item.ID),
			Name:    item.Names,
			Image:   item.Image,
			State:   normalizeState(item.State),
			Status:  item.Status,
			Command: item.Command,
		})
	}
	return result, nil
}

func (p *cliProvider) listImages(ctx context.Context) ([]model.Image, error) {
	out, err := p.output(ctx, "images", "--no-trunc", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []model.Image{}, nil
	}

	result := make([]model.Image, 0, len(lines))
	for _, line := range lines {
		var item struct {
			ID         string `json:"ID"`
			Repository string `json:"Repository"`
			Tag        string `json:"Tag"`
			Size       string `json:"Size"`
		}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("parse image output: %w", err)
		}
		tag := strings.TrimSpace(item.Repository + ":" + item.Tag)
		result = append(result, model.Image{
			ID:       shortID(item.ID),
			RepoTags: tag,
			SizeMB:   parseSizeMB(item.Size),
			Created:  0,
		})
	}
	return result, nil
}

func (p *cliProvider) startContainer(ctx context.Context, id string) error {
	if err := ValidateContainerID(id); err != nil {
		return err
	}
	return p.raw(ctx, "start", id)
}

func (p *cliProvider) stopContainer(ctx context.Context, id string) error {
	if err := ValidateContainerID(id); err != nil {
		return err
	}
	return p.raw(ctx, "stop", id)
}

func (p *cliProvider) restartContainer(ctx context.Context, id string) error {
	if err := ValidateContainerID(id); err != nil {
		return err
	}
	return p.raw(ctx, "restart", id)
}

func (p *cliProvider) output(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, p.dockerBin, args...)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker %s failed: %v (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(data)))
	}
	return string(data), nil
}

func (p *cliProvider) raw(ctx context.Context, args ...string) error {
	_, err := p.output(ctx, args...)
	return err
}

type simulatedProvider struct {
	mu         sync.Mutex
	containers []model.Container
	images     []model.Image
}

func newSimulatedProvider() *simulatedProvider {
	now := time.Now().Unix()
	return &simulatedProvider{
		containers: []model.Container{
			{ID: "9df8a4a8ce2a", Name: "traefik-proxy", Image: "traefik:v3", State: "running", Status: "Up 3 hours", Command: "traefik --api.insecure=true"},
			{ID: "f41c8f4b7a9d", Name: "postgres-db", Image: "postgres:16", State: "running", Status: "Up 1 day", Command: "docker-entrypoint.sh postgres"},
			{ID: "52c0c9150e4f", Name: "redis-cache", Image: "redis:7", State: "exited", Status: "Exited (0) 2 hours ago", Command: "redis-server"},
		},
		images: []model.Image{
			{ID: "sha256:1a8d", RepoTags: "traefik:v3", SizeMB: 142, Created: now - 3600*48},
			{ID: "sha256:8c42", RepoTags: "postgres:16", SizeMB: 386, Created: now - 3600*72},
			{ID: "sha256:9f20", RepoTags: "redis:7", SizeMB: 114, Created: now - 3600*24},
		},
	}
}

func (p *simulatedProvider) dashboard(context.Context) (model.Dashboard, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	up := 0
	for _, c := range p.containers {
		if c.State == "running" {
			up++
		}
	}
	return model.Dashboard{
		ContainersTotal: len(p.containers),
		ContainersUp:    up,
		ContainersDown:  len(p.containers) - up,
		ImagesTotal:     len(p.images),
		Simulated:       true,
	}, nil
}

func (p *simulatedProvider) listContainers(context.Context) ([]model.Container, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]model.Container(nil), p.containers...), nil
}

func (p *simulatedProvider) listImages(context.Context) ([]model.Image, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]model.Image(nil), p.images...), nil
}

func (p *simulatedProvider) startContainer(_ context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx, err := p.findContainer(id)
	if err != nil {
		return err
	}
	p.containers[idx].State = "running"
	p.containers[idx].Status = "Up just now"
	return nil
}

func (p *simulatedProvider) stopContainer(_ context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx, err := p.findContainer(id)
	if err != nil {
		return err
	}
	p.containers[idx].State = "exited"
	p.containers[idx].Status = "Exited (0) just now"
	return nil
}

func (p *simulatedProvider) restartContainer(_ context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx, err := p.findContainer(id)
	if err != nil {
		return err
	}
	p.containers[idx].State = "running"
	p.containers[idx].Status = "Up less than a second"
	return nil
}

func (p *simulatedProvider) findContainer(id string) (int, error) {
	for i := range p.containers {
		if p.containers[i].ID == id {
			return i, nil
		}
	}
	return -1, errors.New("container not found: " + id)
}

func shortID(id string) string {
	if strings.HasPrefix(id, "sha256:") {
		id = strings.TrimPrefix(id, "sha256:")
	}
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func normalizeState(state string) string {
	if strings.Contains(strings.ToLower(state), "up") || strings.ToLower(state) == "running" {
		return "running"
	}
	return "exited"
}

func parseSizeMB(size string) int64 {
	parts := strings.Fields(strings.TrimSpace(strings.ToUpper(size)))
	if len(parts) == 0 {
		return 0
	}

	var value float64
	_, err := fmt.Sscanf(parts[0], "%f", &value)
	if err != nil {
		return 0
	}

	unit := "MB"
	if len(parts) > 1 {
		unit = parts[1]
	} else if strings.HasSuffix(parts[0], "GB") || strings.HasSuffix(parts[0], "MB") || strings.HasSuffix(parts[0], "KB") || strings.HasSuffix(parts[0], "B") {
		// docker can return values like 132MB without spaces.
		suffixes := []string{"GB", "MB", "KB", "B"}
		for _, s := range suffixes {
			if strings.HasSuffix(parts[0], s) {
				unit = s
				trimmed := strings.TrimSuffix(parts[0], s)
				if _, err := fmt.Sscanf(trimmed, "%f", &value); err != nil {
					return 0
				}
				break
			}
		}
	}

	switch unit {
	case "GB":
		return int64(value * 1024)
	case "KB":
		return int64(value / 1024)
	case "B":
		return int64(value / 1024 / 1024)
	default:
		return int64(value)
	}
}

func ValidateContainerID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("container id is required")
	}
	return nil
}
