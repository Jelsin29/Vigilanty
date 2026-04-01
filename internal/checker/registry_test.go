package checker

import "testing"

type stubChecker struct {
	name string
}

func (s *stubChecker) Name() string {
	return s.name
}

func (s *stubChecker) Check(ctx CheckContext) CheckResult {
	return CheckResult{Status: Passed}
}

func TestRegisterAndCreate(t *testing.T) {
	checkerType := "test-register-create"
	Register(checkerType, func(config map[string]interface{}) (Checker, error) {
		if config["name"] != "happy-path" {
			t.Fatalf("factory config[name] = %v, want happy-path", config["name"])
		}
		return &stubChecker{name: "created-checker"}, nil
	})

	instance, err := Create(checkerType, map[string]interface{}{"name": "happy-path"})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	if got := instance.Name(); got != "created-checker" {
		t.Fatalf("Create().Name() = %q, want %q", got, "created-checker")
	}
}

func TestCreateUnknownCheckerType(t *testing.T) {
	_, err := Create("checker-does-not-exist", nil)
	if err == nil {
		t.Fatal("Create() error = nil, want error")
	}
}

func TestDoubleRegistrationOverwritesFactory(t *testing.T) {
	checkerType := "test-double-registration"

	Register(checkerType, func(config map[string]interface{}) (Checker, error) {
		return &stubChecker{name: "first"}, nil
	})
	Register(checkerType, func(config map[string]interface{}) (Checker, error) {
		return &stubChecker{name: "second"}, nil
	})

	instance, err := Create(checkerType, nil)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	if got := instance.Name(); got != "second" {
		t.Fatalf("Create().Name() = %q, want %q", got, "second")
	}
}
