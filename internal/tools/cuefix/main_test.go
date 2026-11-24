package main

import "testing"

func TestFix(t *testing.T) {
	tests := []struct {
		name, text, want string
	}{
		{
			"module path",
			"import \"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api\"\n",
			"import \"github.com/bemusedvermin/achlydesa-heptarchy/internal/core/api\"\n",
		},
		{
			"embedded field",
			"#Item: {\n\tData: #Data\n\tweight: int\n}\n",
			"#Item: {\n\t#Data\n\tweight: int\n}\n",
		},
		{
			"field of a struct type that isn't embedded",
			"#Item: {\n\tParent: #Data\n}\n",
			"#Item: {\n\tParent: #Data\n}\n",
		},
		{
			"a field on the top level",
			"Data: #Data\n",
			"Data: #Data\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := fix(test.text); got != test.want {
				t.Errorf("fix(%q) = %q, want %q", test.text, got, test.want)
			}
		})
	}
}
