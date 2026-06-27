package registry

import (
	"errors"
	"testing"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/extension/types"
)

type registryExtension struct {
	*types.OptionalImpl
	name string
}

func (e registryExtension) Name() string {
	return e.name
}

func (e registryExtension) Version() string {
	return "v1.0.0"
}

func (e registryExtension) Init(*config.Config, types.ManagerInterface) error {
	return nil
}

func (e registryExtension) GetMetadata() types.Metadata {
	return types.Metadata{Name: e.name}
}

func (e registryExtension) GetHandlers() types.Handler {
	return nil
}

func (e registryExtension) GetServices() types.Service {
	return nil
}

func (e registryExtension) Dependencies() []string {
	return nil
}

func resetRegistryForTest(t *testing.T) {
	t.Helper()

	mutex.Lock()
	defer mutex.Unlock()
	extensionRegistry = make(map[string]Entry)
}

func TestRegisterERejectsNilAndEmptyExtensions(t *testing.T) {
	resetRegistryForTest(t)

	if err := RegisterE(nil); !errors.Is(err, ErrNilExtension) {
		t.Fatalf("expected ErrNilExtension, got %v", err)
	}
	if err := RegisterE(registryExtension{OptionalImpl: &types.OptionalImpl{}}); !errors.Is(err, ErrEmptyExtensionName) {
		t.Fatalf("expected ErrEmptyExtensionName, got %v", err)
	}
}

func TestRegisterERejectsDuplicates(t *testing.T) {
	resetRegistryForTest(t)

	extension := registryExtension{OptionalImpl: &types.OptionalImpl{}, name: "demo"}
	if err := RegisterE(extension); err != nil {
		t.Fatalf("register extension: %v", err)
	}
	if err := RegisterE(extension); !errors.Is(err, ErrDuplicateExtension) {
		t.Fatalf("expected ErrDuplicateExtension, got %v", err)
	}
}

func TestRegisterToGroupWithWeakDepsEStoresMetadata(t *testing.T) {
	resetRegistryForTest(t)

	extension := registryExtension{OptionalImpl: &types.OptionalImpl{}, name: "demo"}
	if err := RegisterToGroupWithWeakDepsE(extension, "core", []string{"optional"}); err != nil {
		t.Fatalf("register extension: %v", err)
	}

	registered := GetExtensions()
	entry := registered["demo"]
	if entry.Group != "core" || len(entry.WeakDependencies) != 1 || entry.WeakDependencies[0] != "optional" {
		t.Fatalf("unexpected registry entry: %#v", entry)
	}
}
