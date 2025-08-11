package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Package struct {
	Title      string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id         string  `json:"-" yaml:"-"`
	Name       string  `json:"name,omitempty" yaml:"name,omitempty"`
	Installed  matcher `json:"installed" yaml:"installed"`
	Versions   matcher `json:"versions,omitempty" yaml:"versions,omitempty"`
	Retry      bool    `json:"retry,omitempty" yaml:"retry,omitempty"`
	RetryDelay int     `json:"retry_delay,omitempty" yaml:"retry_delay,omitempty"`
	Skip       bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	PackageResourceKey  = "package"
	PackageResourceName = "Package"
)

func init() {
	registerResource(PackageResourceKey, &Package{})
}

func (p *Package) ID() string {
	if p.Name != "" && p.Name != p.id {
		return fmt.Sprintf("%s: %s", p.id, p.Name)
	}
	return p.id
}
func (p *Package) SetID(id string)  { p.id = id }
func (p *Package) SetSkip()         { p.Skip = true }
func (p *Package) TypeKey() string  { return PackageResourceKey }
func (p *Package) TypeName() string { return PackageResourceName }
func (p *Package) GetTitle() string { return p.Title }
func (p *Package) GetMeta() meta    { return p.Meta }
func (p *Package) GetName() string {
	if p.Name != "" {
		return p.Name
	}
	return p.id
}
func (p *Package) GetRetry() bool      { return p.Retry }
func (p *Package) GetRetryDelay() int  { return p.RetryDelay }

func (p *Package) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), idKey{}, p.ID())
	skip := p.Skip

	var results []TestResult
	
	// Handle retry logic for installed check
	if p.Retry {
		results = append(results, ValidateValueWithRetry(p, "installed", p.Installed, 
			func() (any, error) {
				sysPkg := sys.NewPackage(ctx, p.GetName(), sys, util.Config{})
				return sysPkg.Installed()
			}, skip, p.RetryDelay))
	} else {
		sysPkg := sys.NewPackage(ctx, p.GetName(), sys, util.Config{})
		results = append(results, ValidateValue(p, "installed", p.Installed, sysPkg.Installed, skip))
	}
	
	if shouldSkip(results) {
		skip = true
	}
	
	// Handle retry logic for versions check
	if p.Versions != nil {
		if p.Retry {
			results = append(results, ValidateValueWithRetry(p, "version", p.Versions,
				func() (any, error) {
					sysPkg := sys.NewPackage(ctx, p.GetName(), sys, util.Config{})
					return sysPkg.Versions()
				}, skip, p.RetryDelay))
		} else {
			sysPkg := sys.NewPackage(ctx, p.GetName(), sys, util.Config{})
			results = append(results, ValidateValue(p, "version", p.Versions, sysPkg.Versions, skip))
		}
	}
	
	return results
}

func NewPackage(sysPackage system.Package, config util.Config) (*Package, error) {
	name := sysPackage.Name()
	installed, _ := sysPackage.Installed()
	p := &Package{
		id:        name,
		Installed: installed,
	}
	if !contains(config.IgnoreList, "versions") {
		if versions, err := sysPackage.Versions(); err == nil {
			p.Versions = versions
		}
	}
	return p, nil
}
