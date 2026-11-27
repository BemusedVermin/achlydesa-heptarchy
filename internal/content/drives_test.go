package content

import "testing"

func TestDrivesAreChecked(t *testing.T) {
	data := validData + `
drives: {
	"drive.water": {name: "text.drive.water", kind: "need", urge: "1"}
	"drive.pay": {
		name: "text.drive.pay", kind: "desire", urge: "level"
		thresholds: [{at: 1, do: [{do: "damage", params: {}}]}]
	}
}
templates: "template.thing": drives: {"drive.water": 1, "drive.thirst": 2, "drive.pay": -1}
text: {"text.drive.water": "water", "text.drive.pay": "pay"}
`
	_, err := Load(writeData(t, data))
	requireErrContains(t, err,
		`drives["drive.water"].rise: required for a need`,
		`drives["drive.pay"].thresholds: a desire never kills`,
		`templates["template.thing"].drives["drive.thirst"]: unknown drive "drive.thirst"`,
		`templates["template.thing"].drives["drive.pay"]: a weight must not be negative`,
		`agents.weight: required when a template lists drives`,
	)
}

func TestDriveFormulas(t *testing.T) {
	registry := sampleRegistry(t)
	urge, err := registry.DriveUrge("drive.water", DriveEnv{Level: 6})
	if err != nil || urge <= 0 || urge >= 1 {
		t.Errorf("water urge at level 6 = %v, %v; want between 0 and 1", urge, err)
	}
	if rise, err := registry.DriveRise("drive.pay", DriveEnv{}); err != nil || rise != 0 {
		t.Errorf("pay rise = %v, %v; want 0 for a desire without a rise", rise, err)
	}
	if _, err := registry.DriveUrge("drive.safety", DriveEnv{Level: -40}); err == nil {
		t.Error("an urge below 0 was accepted")
	}
	weight, err := registry.DriveWeight(WeightEnv{Typical: 2, Roll: 0.5})
	if err != nil || weight != 2 {
		t.Errorf("weight at the middle roll = %v, %v; want the typical 2", weight, err)
	}
}
