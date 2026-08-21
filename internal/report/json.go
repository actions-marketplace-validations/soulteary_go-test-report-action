package report

import (
	"bytes"
	"encoding/json"

	"github.com/soulteary/go-test-report-action/internal/model"
)

// JSON renders the report as deterministic, pretty-printed JSON with a trailing
// newline. Packages and failures are already sorted by Build. No timestamps or
// absolute paths are included.
func JSON(rep model.Report) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(rep); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
