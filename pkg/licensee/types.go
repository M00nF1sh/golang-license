package licensee

type LicenseMeta struct {
	Title string `json:"title"`
}

type License struct {
	SpdxID string      `json:"spdx_id"`
	Meta   LicenseMeta `json:"meta"`
}

type Matcher struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

type MatchedFile struct {
	Filename          string  `json:"filename"`
	Content           string  `json:"content"`
	ContentNormalized string  `json:"content_normalized"`
	Matcher           Matcher `json:"matcher"`
	MatchedLicense    string  `json:"matched_license"`
	Attribution       string  `json:"attribution"`
}

type DetectionResult struct {
	Licenses     []License     `json:"licenses"`
	MatchedFiles []MatchedFile `json:"matched_files"`
}
