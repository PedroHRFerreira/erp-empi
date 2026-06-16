package validation

import "testing"

func TestIsCPF(t *testing.T) {
	t.Parallel()

	if !IsCPF("529.982.247-25") {
		t.Fatal("expected valid cpf")
	}
	if IsCPF("111.111.111-11") {
		t.Fatal("expected repeated digits cpf to be invalid")
	}
}
