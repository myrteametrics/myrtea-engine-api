package export

type CSVParameters struct {
	Columns           []string
	ColumnsLabel      []string
	FormatColumnsData map[string]string
	Separator         rune
	Limit             int64
	ChunkSize         int64
	FileName          string
}

// Equals compares two CSVParameters
func (p CSVParameters) Equals(Params CSVParameters) bool {
	if p.FileName != Params.FileName {
		return false
	}
	if p.Separator != Params.Separator {
		return false
	}
	if p.Limit != Params.Limit {
		return false
	}
	if p.ChunkSize != Params.ChunkSize {
		return false
	}
	if len(p.Columns) != len(Params.Columns) {
		return false
	}
	for i, column := range p.Columns {
		if column != Params.Columns[i] {
			return false
		}
	}
	if len(p.ColumnsLabel) != len(Params.ColumnsLabel) {
		return false
	}
	for i, columnLabel := range p.ColumnsLabel {
		if columnLabel != Params.ColumnsLabel[i] {
			return false
		}
	}
	if len(p.FormatColumnsData) != len(Params.FormatColumnsData) {
		return false
	}
	for key, value := range p.FormatColumnsData {
		if value != Params.FormatColumnsData[key] {
			return false
		}
	}
	return true
}
