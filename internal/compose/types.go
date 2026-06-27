package compose

// ComposeDoc is the in-memory representation of a Compose document.
// Using map[string]interface{} avoids the need to enumerate every possible
// Compose key and allows arbitrary depth merging.
type ComposeDoc = map[string]interface{}

// GetServices safely returns the services map from a compose document.
func GetServices(doc ComposeDoc) map[string]interface{} {
	v, _ := doc["services"].(map[string]interface{})
	return v
}

// GetService returns a specific service map from a compose document.
func GetService(doc ComposeDoc, name string) (map[string]interface{}, bool) {
	services := GetServices(doc)
	if services == nil {
		return nil, false
	}
	svc, ok := services[name].(map[string]interface{})
	return svc, ok
}

// GetNetworks safely returns the top-level networks map from a compose document.
func GetNetworks(doc ComposeDoc) map[string]interface{} {
	v, _ := doc["networks"].(map[string]interface{})
	return v
}
