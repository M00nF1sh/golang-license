package licensee

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/m00nf1sh/golang-license/pkg/module"
	"os/exec"
)

type Detector interface {
	Detect(dependency module.Dependency) (DetectionResult, error)
}

func NewDetector() Detector {
	return &detector{}
}

// detector based on https://github.com/licensee/licensee
type detector struct {
}

func (d *detector) Detect(dependency module.Dependency) (DetectionResult, error) {
	d.checkLicensee()
	cmd := exec.Command("licensee", "detect", "--json", "--no-readme", dependency.Module.Dir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return DetectionResult{}, err
	}
	ret := DetectionResult{}
	if err := json.Unmarshal(out.Bytes(), &ret); err != nil {
		return DetectionResult{}, err
	}
	return ret, nil
}

func (d *detector) checkLicensee() error {
	_, err := exec.LookPath("licensee")
	if err != nil {
		return errors.New("please install licensee in your system: https://github.com/licensee/licensee")
	}
	return nil
}
