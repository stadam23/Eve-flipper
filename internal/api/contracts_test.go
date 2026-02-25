package api

import "testing"

func TestResolveContractTypeName(t *testing.T) {
	t.Run("prefers_live_esi_name", func(t *testing.T) {
		got := resolveContractTypeName(
			"Wightstorm Nirvana Booster IV",
			"Expired Wightstorm Nirvana Booster IV",
			90689,
		)
		want := "Expired Wightstorm Nirvana Booster IV"
		if got != want {
			t.Fatalf("resolveContractTypeName = %q, want %q", got, want)
		}
	})

	t.Run("falls_back_to_sde_name", func(t *testing.T) {
		got := resolveContractTypeName("Tritanium", "", 34)
		if got != "Tritanium" {
			t.Fatalf("resolveContractTypeName = %q, want Tritanium", got)
		}
	})

	t.Run("falls_back_to_type_id_label", func(t *testing.T) {
		got := resolveContractTypeName("", "", 12345)
		if got != "Type 12345" {
			t.Fatalf("resolveContractTypeName = %q, want Type 12345", got)
		}
	})
}
