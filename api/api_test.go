package api

import (
	"testing"
)

func TestSplitStringToPack(t *testing.T) {
	var packName string
	var integrationName string
	var unused string
	parseArray := []*string{&unused, &unused, &unused, &packName, &integrationName}

	err := parsePath("/api/packs/packname/intname", parseArray)
	if err != nil {
		t.Fail()
	}

	if packName != "packname" {
		t.Fail()
	}
}
