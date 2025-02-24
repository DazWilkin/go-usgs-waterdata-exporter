package waterdata

type GetInstantaneousValuesResponse struct {
	Name            string `json:"name"`
	DeclaredType    string `json:"declaredType"`
	Scope           string `json:"scope"`
	Value           Value  `json:"value"`
	Nil             bool   `json:"nil"`
	GlobalScope     bool   `json:"globalScope"`
	TypeSubstituted bool   `json:"typeSubstituted"`
}

type Criteria struct {
	LocationParam string `json:"locationParam"`
	VariableParam string `json:"variableParam"`
	// TODO Incomplete
}

type Note struct {
	Value string `json:"value"`
	Title string `json:"title"`
}

type QueryInfo struct {
	QueryURL string   `json:"queryURL"`
	Criteria Criteria `json:"criteria"`
	Note     []Note   `json:"note"`
}

type SiteCode struct {
	Value      string `json:"value"`
	Network    string `json:"network"`
	AgencyCode string `json:"AgencyCode"`
}

type SourceInfo struct {
	SiteName string     `json:"siteName"`
	SiteCode []SiteCode `json:"siteCode"`
}

type TimeSeries struct {
	SourceInfo SourceInfo         `json:"sourceInfo"`
	Variable   Variable           `json:"variable"`
	Values     []TimeSeries_Value `json:"values"`
	Name       string             `json:"name"`
}

type TimeSeries_Value struct {
	Value []TimeSeries_Value_Value `json:"value"`
}

type TimeSeries_Value_Value struct {
	Value      string   `json:"value"`
	Qualifiers []string `json:"qualifiers"`
	DateTime   string   `json:"dateTime"`
}

type Value struct {
	QueryInfo  QueryInfo    `json:"queryInfo"`
	TimeSeries []TimeSeries `json:"timeSeries"`
}

type Variable struct {
	VariableCode []VariableCode `json:"variableCode"`
	VariableName string         `json:"variableName"`
	// TODO Incomplete
}

func (x *Variable) Contains(value string) bool {
	for _, v := range x.VariableCode {
		if v.Value == value {
			return true
		}
	}

	return false
}

type VariableCode struct {
	Value      string `json:"value"`
	Network    string `json:"network"`
	Vocabulary string `json:"vocabulary"`
	VariableID int    `json:"variableID"`
	Default    bool   `json:"default"`
}
