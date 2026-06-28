package compose

import "testing"

func TestGetServices(t *testing.T) {
	doc := ComposeDoc{
		"services": map[string]interface{}{
			"web": map[string]interface{}{"image": "nginx"},
		},
	}
	services := GetServices(doc)
	if len(services) != 1 {
		t.Fatalf("got %d services, want 1", len(services))
	}
	if _, ok := services["web"]; !ok {
		t.Error("web service missing")
	}

	if got := GetServices(ComposeDoc{}); got != nil {
		t.Errorf("empty doc should return nil, got %v", got)
	}
}

func TestGetService(t *testing.T) {
	doc := ComposeDoc{
		"services": map[string]interface{}{
			"api": map[string]interface{}{"image": "alpine"},
		},
	}
	svc, ok := GetService(doc, "api")
	if !ok || svc["image"] != "alpine" {
		t.Errorf("GetService(api) = %v, %v", svc, ok)
	}
	if _, ok := GetService(doc, "missing"); ok {
		t.Error("expected missing service to return false")
	}
}

func TestGetNetworks(t *testing.T) {
	doc := ComposeDoc{
		"networks": map[string]interface{}{
			"atrisos_net": map[string]interface{}{"external": true},
		},
	}
	nets := GetNetworks(doc)
	if len(nets) != 1 {
		t.Fatalf("got %d networks, want 1", len(nets))
	}
}
