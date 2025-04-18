package utility

import "encoding/json"

func PrettyPrintJSON(data interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return string(bytes)
}
