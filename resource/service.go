package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Service struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id        string  `json:"-" yaml:"-"`
	Name      string  `json:"name,omitempty" yaml:"name,omitempty"`
	Enabled   matcher `json:"enabled" yaml:"enabled"`
	Running   matcher `json:"running" yaml:"running"`
	Skip      bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
	RunLevels matcher `json:"runlevels,omitempty" yaml:"runlevels,omitempty"`
}

const (
	ServiceResourceKey  = "service"
	ServiceResourceName = "Service"
)

func init() {
	registerResource(ServiceResourceKey, &Service{})
}

func (s *Service) ID() string {
	if s.Name != "" && s.Name != s.id {
		return fmt.Sprintf("%s: %s", s.id, s.Name)
	}
	return s.id
}
func (s *Service) SetID(id string)  { s.id = id }
func (s *Service) SetSkip()         { s.Skip = true }
func (s *Service) TypeKey() string  { return ServiceResourceKey }
func (s *Service) TypeName() string { return ServiceResourceName }
func (s *Service) GetTitle() string { return s.Title }
func (s *Service) GetMeta() meta    { return s.Meta }
func (s *Service) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return s.id
}

func (s *Service) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), idKey{}, s.ID())
	skip := s.Skip
	sysservice := sys.NewService(ctx, s.GetName(), sys, util.Config{})

	var results []TestResult
	if s.Enabled != nil {
		results = append(results, ValidateValue(s, "enabled", s.Enabled, sysservice.Enabled, skip))
	}
	if s.Running != nil {
		results = append(results, ValidateValue(s, "running", s.Running, sysservice.Running, skip))
	}
	if s.RunLevels != nil {
		results = append(results, ValidateValue(s, "runlevels", s.RunLevels, sysservice.RunLevels, skip))
	}
	return results
}

func NewService(sysService system.Service, config util.Config) (*Service, error) {
	service := sysService.Service()
	enabled, err := sysService.Enabled()
	if err != nil {
		return nil, err
	}
	running, err := sysService.Running()
	if err != nil {
		return nil, err
	}
	return &Service{
		id:      service,
		Enabled: enabled,
		Running: running,
	}, nil
}
