package compose

import "testing"

func TestIsBindMountSource(t *testing.T) {
	tests := []struct {
		src  string
		want bool
	}{
		{"/data", true},
		{"./data", true},
		{"../data", true},
		{"db_data", false},
		{"named_volume", false},
	}
	for _, tt := range tests {
		if got := isBindMountSource(tt.src); got != tt.want {
			t.Errorf("isBindMountSource(%q) = %v, want %v", tt.src, got, tt.want)
		}
	}
}

func TestApplySelinuxZString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"/host/data:/data", "/host/data:/data:z"},
		{"/host/data:/data:ro", "/host/data:/data:ro,z"},
		{"/host/data:/data:z", "/host/data:/data:z"},
		{"/host/data:/data:Z", "/host/data:/data:Z"},
		{"db_data:/data", "db_data:/data"},
		{"./local:/app", "./local:/app:z"},
	}
	for _, tt := range tests {
		if got := applySelinuxZString(tt.in); got != tt.want {
			t.Errorf("applySelinuxZString(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestApplySelinuxZObject(t *testing.T) {
	t.Run("bind mount gets selinux z", func(t *testing.T) {
		vol := map[string]interface{}{
			"type":   "bind",
			"source": "/host/data",
			"target": "/data",
		}
		applySelinuxZObject(vol)
		if vol["selinux"] != "z" {
			t.Errorf("selinux = %v, want z", vol["selinux"])
		}
	})

	t.Run("named volume unchanged", func(t *testing.T) {
		vol := map[string]interface{}{
			"type":   "volume",
			"source": "db_data",
			"target": "/data",
		}
		applySelinuxZObject(vol)
		if _, ok := vol["selinux"]; ok {
			t.Errorf("selinux should not be set on named volume")
		}
	})

	t.Run("existing selinux preserved", func(t *testing.T) {
		vol := map[string]interface{}{
			"type":    "bind",
			"source":  "/host/data",
			"target":  "/data",
			"selinux": "Z",
		}
		applySelinuxZObject(vol)
		if vol["selinux"] != "Z" {
			t.Errorf("selinux = %v, want Z", vol["selinux"])
		}
	})
}

func TestApplySelinuxToAllServices(t *testing.T) {
	doc := ComposeDoc{
		"services": map[string]interface{}{
			"web": map[string]interface{}{
				"volumes": []interface{}{
					"/host/www:/var/www",
					"static_data:/static",
				},
			},
		},
	}
	applySelinuxToAllServices(doc)

	svc := doc["services"].(map[string]interface{})["web"].(map[string]interface{})
	vols := svc["volumes"].([]interface{})
	if vols[0].(string) != "/host/www:/var/www:z" {
		t.Errorf("bind mount = %q, want :z suffix", vols[0])
	}
	if vols[1].(string) != "static_data:/static" {
		t.Errorf("named volume changed unexpectedly: %q", vols[1])
	}
}

func TestMergeNetworkList(t *testing.T) {
	t.Run("nil adds default and atrisos_net", func(t *testing.T) {
		got := mergeNetworkList(nil, "atrisos_net")
		list, ok := got.([]interface{})
		if !ok {
			t.Fatalf("got %T, want []interface{}", got)
		}
		if len(list) != 2 || list[0] != "default" || list[1] != "atrisos_net" {
			t.Errorf("got %v, want [default atrisos_net]", list)
		}
	})

	t.Run("list appends network", func(t *testing.T) {
		got := mergeNetworkList([]interface{}{"internal"}, "atrisos_net")
		list := got.([]interface{})
		if len(list) != 2 || list[1] != "atrisos_net" {
			t.Errorf("got %v", list)
		}
	})

	t.Run("list does not duplicate", func(t *testing.T) {
		existing := []interface{}{"atrisos_net", "internal"}
		got := mergeNetworkList(existing, "atrisos_net")
		list, ok := got.([]interface{})
		if !ok || len(list) != 2 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("map adds network key", func(t *testing.T) {
		got := mergeNetworkList(map[string]interface{}{"internal": nil}, "atrisos_net")
		m := got.(map[string]interface{})
		if _, ok := m["atrisos_net"]; !ok {
			t.Errorf("atrisos_net missing from map: %v", m)
		}
		if _, ok := m["internal"]; !ok {
			t.Errorf("internal network removed from map")
		}
	})
}
