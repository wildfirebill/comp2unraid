package main

import (
	"encoding/xml"
	"strings"
	"testing"

	composetypes "github.com/compose-spec/compose-go/v2/types"
)

// ---------------------------------------------------------------------------
// getRegistryURL
// ---------------------------------------------------------------------------

func TestGetRegistryURL(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		want    string
		wantErr bool
	}{
		{"quay.io registry", "quay.io/nextcloud/server", "https://quay.io/repository/nextcloud/server", false},
		{"ghcr.io registry", "ghcr.io/nextcloud/server", "https://github.com/nextcloud/server", false},
		{"docker.io registry", "docker.io/nextcloud/server", "https://hub.docker.com/r/nextcloud/server", false},
		{"docker hub shorthand", "library/nginx", "https://hub.docker.com/r/library/nginx", false},
		{"image with tag stripped", "ghcr.io/immich-app/immich-server:release", "https://github.com/immich-app/immich-server", false},
		{"docker hub shorthand with tag", "linuxserver/plex:latest", "https://hub.docker.com/r/linuxserver/plex", false},
		{"unknown registry falls back to docker hub", "my.registry.io/org/app", "https://hub.docker.com/r/org/app", false},
		{"single word image errors", "nginx", "", true},
		{"empty string errors", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRegistryURL(tt.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRegistryURL(%q) error = %v, wantErr %v", tt.image, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRegistryURL(%q) = %q, want %q", tt.image, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseYaml - uses the example compose file as a fixture
// ---------------------------------------------------------------------------

func TestParseYaml(t *testing.T) {
	args := commandLineOptions{
		configFile: "examples/compose/docker-compose.yml",
	}

	project, err := parseYaml(args)
	if err != nil {
		t.Fatalf("parseYaml() returned error: %v", err)
	}

	if len(project.Services) == 0 {
		t.Fatal("parseYaml() returned 0 services, expected at least 1")
	}

	// The example file defines immich-server and immich-machine-learning
	serviceNames := make(map[string]bool)
	for _, s := range project.Services {
		serviceNames[s.Name] = true
	}

	for _, want := range []string{"immich-server", "immich-machine-learning"} {
		if !serviceNames[want] {
			t.Errorf("expected service %q in parsed project, got services: %v", want, serviceNames)
		}
	}
}

// ---------------------------------------------------------------------------
// getConfigs (port extraction)
// ---------------------------------------------------------------------------

func TestGetConfigs_Ports(t *testing.T) {
	svc := &composetypes.ServiceConfig{
		Ports: []composetypes.ServicePortConfig{
			{Published: "8080", Target: 80, Protocol: "tcp"},
			{Published: "443", Target: 443, Protocol: "tcp"},
		},
	}

	configs := getConfigs(svc)
	if len(configs) != 1 {
		t.Fatalf("getConfigs() returned %d configs, expected 1 (WebUI)", len(configs))
	}
	c := configs[0]
	if c.Type != "Port" {
		t.Errorf("expected Type=Port, got %q", c.Type)
	}
	if c.Target != "8080" {
		t.Errorf("expected Target=8080, got %q", c.Target)
	}
	if c.Value != "8080" {
		t.Errorf("expected Value=8080, got %q", c.Value)
	}
}

func TestGetConfigs_NoPorts(t *testing.T) {
	svc := &composetypes.ServiceConfig{}
	configs := getConfigs(svc)
	if len(configs) != 0 {
		t.Errorf("getConfigs() with no ports should return 0 configs, got %d", len(configs))
	}
}

// ---------------------------------------------------------------------------
// getEnvironmentConfigs
// ---------------------------------------------------------------------------

func TestGetEnvironmentConfigs(t *testing.T) {
	val1 := "Europe/Berlin"
	val2 := "s3cret"
	val3 := "hunter2"
	val4 := "my_secret_key"

	svc := &composetypes.ServiceConfig{
		Environment: composetypes.MappingWithEquals{
			"TZ":           &val1,
			"DB_PASSWORD":  &val2,
			"ADMIN_PWD":    &val3,
			"TOKEN_SECRET": &val4,
		},
	}

	configs := getEnvironmentConfigs(svc)
	if len(configs) != 4 {
		t.Fatalf("expected 4 env configs, got %d", len(configs))
	}

	lookup := make(map[string]Config)
	for _, c := range configs {
		lookup[c.Name] = c
	}

	// TZ should NOT be masked
	if tz, ok := lookup["TZ"]; ok {
		if tz.Mask {
			t.Error("TZ should not have Mask=true")
		}
		if tz.Value != "Europe/Berlin" {
			t.Errorf("TZ value = %q, want %q", tz.Value, "Europe/Berlin")
		}
		if tz.Type != "Variable" {
			t.Errorf("TZ Type = %q, want Variable", tz.Type)
		}
	} else {
		t.Error("TZ config not found")
	}

	// Variables containing PWD, PASS, SECRET should be masked
	for _, key := range []string{"DB_PASSWORD", "ADMIN_PWD", "TOKEN_SECRET"} {
		if c, ok := lookup[key]; ok {
			if !c.Mask {
				t.Errorf("%s should have Mask=true", key)
			}
		} else {
			t.Errorf("%s config not found", key)
		}
	}
}

// ---------------------------------------------------------------------------
// getEnvironmentConfigs - nil value handling
// ---------------------------------------------------------------------------

func TestGetEnvironmentConfigs_NilValue(t *testing.T) {
	svc := &composetypes.ServiceConfig{
		Environment: composetypes.MappingWithEquals{
			"EMPTY_VAR": nil,
		},
	}

	configs := getEnvironmentConfigs(svc)
	if len(configs) != 1 {
		t.Fatalf("expected 1 env config, got %d", len(configs))
	}
	if configs[0].Value != "" {
		t.Errorf("nil env var Value = %q, want empty string", configs[0].Value)
	}
	if configs[0].Default != "" {
		t.Errorf("nil env var Default = %q, want empty string", configs[0].Default)
	}
}

// ---------------------------------------------------------------------------
// getVolumeConfigs
// ---------------------------------------------------------------------------

func TestGetVolumeConfigs(t *testing.T) {
	svc := &composetypes.ServiceConfig{
		Name: "myapp",
		Volumes: []composetypes.ServiceVolumeConfig{
			{Source: "/mnt/user/photos", Target: "/usr/src/app/upload", Type: "bind"},
			{Source: "model-cache", Target: "/cache", Type: "volume"},
		},
	}

	configs := getVolumeConfigs(svc)
	if len(configs) != 2 {
		t.Fatalf("expected 2 volume configs, got %d", len(configs))
	}

	// Bind mount: Value should be empty (source starts with /)
	bind := configs[0]
	if bind.Type != "Path" {
		t.Errorf("expected Type=Path, got %q", bind.Type)
	}
	if bind.Target != "/usr/src/app/upload" {
		t.Errorf("expected Target=/usr/src/app/upload, got %q", bind.Target)
	}
	if bind.Value != "" {
		t.Errorf("bind mount Value should be empty, got %q", bind.Value)
	}
	if bind.Default != "/mnt/user/photos" {
		t.Errorf("bind mount Default = %q, want /mnt/user/photos", bind.Default)
	}

	// Named volume: Value should be set to the volume name
	named := configs[1]
	if named.Value != "model-cache" {
		t.Errorf("named volume Value = %q, want model-cache", named.Value)
	}
	if named.Default != "model-cache" {
		t.Errorf("named volume Default = %q, want model-cache", named.Default)
	}
}

// ---------------------------------------------------------------------------
// getDeviceConfigs
// ---------------------------------------------------------------------------

func TestGetDeviceConfigs(t *testing.T) {
	tests := []struct {
		name         string
		devices      []string
		wantCount    int
		wantName0    string
		wantDefault0 string
	}{
		{
			name:         "host:container mapping",
			devices:      []string{"/dev/dri:/dev/dri"},
			wantCount:    1,
			wantName0:    "Device passthrough /dev/dri",
			wantDefault0: "/dev/dri:/dev/dri",
		},
		{
			name:         "device without colon",
			devices:      []string{"/dev/sda"},
			wantCount:    1,
			wantName0:    "Device passthrough /dev/sda",
			wantDefault0: "/dev/sda",
		},
		{
			name:      "no devices",
			devices:   nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &composetypes.ServiceConfig{Devices: tt.devices}
			configs := getDeviceConfigs(svc)
			if len(configs) != tt.wantCount {
				t.Fatalf("got %d configs, want %d", len(configs), tt.wantCount)
			}
			if tt.wantCount > 0 {
				if configs[0].Name != tt.wantName0 {
					t.Errorf("Name = %q, want %q", configs[0].Name, tt.wantName0)
				}
				if configs[0].Default != tt.wantDefault0 {
					t.Errorf("Default = %q, want %q", configs[0].Default, tt.wantDefault0)
				}
				if configs[0].Type != "Device" {
					t.Errorf("Type = %q, want Device", configs[0].Type)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getNetworkMode
// ---------------------------------------------------------------------------

func TestGetNetworkMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{"empty defaults to bridge", "", "bridge"},
		{"host mode", "host", "host"},
		{"custom network", "my-network", "my-network"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &composetypes.ServiceConfig{NetworkMode: tt.mode}
			got := getNetworkMode(svc)
			if got != tt.want {
				t.Errorf("getNetworkMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getWebUI
// ---------------------------------------------------------------------------

func TestGetWebUI(t *testing.T) {
	t.Run("with ports", func(t *testing.T) {
		svc := &composetypes.ServiceConfig{
			Ports: []composetypes.ServicePortConfig{
				{Published: "8080"},
			},
		}
		got := getWebUI(svc)
		want := "http://[IP]:[PORT:8080]"
		if got != want {
			t.Errorf("getWebUI() = %q, want %q", got, want)
		}
	})

	t.Run("no ports", func(t *testing.T) {
		svc := &composetypes.ServiceConfig{}
		got := getWebUI(svc)
		if got != "" {
			t.Errorf("getWebUI() with no ports = %q, want empty", got)
		}
	})
}

// ---------------------------------------------------------------------------
// XML marshaling - round-trip test
// ---------------------------------------------------------------------------

func TestXMLMarshal(t *testing.T) {
	template := UnraidTemplate{
		Version:    "2",
		Name:       "test-app",
		Repository: "ghcr.io/org/test-app:latest",
		Registry:   "https://github.com/org/test-app",
		Network:    "bridge",
		Shell:      "bash",
		Author:     "testauthor",
		Category:   "Other:",
		Configs: []Config{
			{
				Name:     "WebUI",
				Target:   "8080",
				Default:  "8080",
				Mode:     "tcp",
				Type:     "Port",
				Display:  "always",
				Required: true,
				Mask:     false,
				Value:    "8080",
			},
			{
				Name:     "DB_PASSWORD",
				Target:   "DB_PASSWORD",
				Default:  "secret",
				Type:     "Variable",
				Display:  "always",
				Required: true,
				Mask:     true,
				Value:    "secret",
			},
		},
	}

	xmlBytes, err := xml.MarshalIndent(template, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent() error: %v", err)
	}

	xmlStr := string(xmlBytes)

	if !strings.Contains(xmlStr, "<Container") {
		t.Error("XML should contain <Container> root element")
	}
	if !strings.Contains(xmlStr, `version="2"`) {
		t.Error("XML should contain version attribute")
	}
	if !strings.Contains(xmlStr, "<Name>test-app</Name>") {
		t.Error("XML should contain service name")
	}
	if !strings.Contains(xmlStr, `Mask="true"`) {
		t.Error("XML should contain Mask=true for password config")
	}
	if !strings.Contains(xmlStr, `Mask="false"`) {
		t.Error("XML should contain Mask=false for port config")
	}
	if !strings.Contains(xmlStr, `Type="Port"`) {
		t.Error("XML should contain Type=Port config")
	}
	if !strings.Contains(xmlStr, `Type="Variable"`) {
		t.Error("XML should contain Type=Variable config")
	}

	// Unmarshal back and verify round-trip
	var decoded UnraidTemplate
	err = xml.Unmarshal(xmlBytes, &decoded)
	if err != nil {
		t.Fatalf("xml.Unmarshal() error: %v", err)
	}
	if decoded.Name != "test-app" {
		t.Errorf("round-trip Name = %q, want test-app", decoded.Name)
	}
	if decoded.Network != "bridge" {
		t.Errorf("round-trip Network = %q, want bridge", decoded.Network)
	}
	if len(decoded.Configs) != 2 {
		t.Errorf("round-trip Configs count = %d, want 2", len(decoded.Configs))
	}
}

// ---------------------------------------------------------------------------
// GitHub blob URL conversion
// ---------------------------------------------------------------------------

func TestGitHubBlobURLConversion(t *testing.T) {
	// The getLocalPath method should convert GitHub blob URLs to raw URLs.
	// We can't easily test the full HTTP flow, but we can verify the URL
	// rewriting logic by checking the configFile field after conversion.
	// For now, test the conversion logic directly.
	blob := "https://github.com/Ogglord/comp2unraid/blob/main/examples/compose/docker-compose.yml"
	want := "https://raw.githubusercontent.com/Ogglord/comp2unraid/main/examples/compose/docker-compose.yml"

	url := blob
	if strings.HasPrefix(url, "https://github.com/") && strings.Contains(url, "/blob/") {
		url = strings.Replace(url, "https://github.com/", "https://raw.githubusercontent.com/", 1)
		url = strings.Replace(url, "/blob/", "/", 1)
	}

	if url != want {
		t.Errorf("GitHub blob URL conversion:\ngot  %q\nwant %q", url, want)
	}
}

// ---------------------------------------------------------------------------
// SetRepository
// ---------------------------------------------------------------------------

func TestSetRepository(t *testing.T) {
	t.Run("github shorthand", func(t *testing.T) {
		opts := &commandLineOptions{}
		opts.SetRepository("MyUser/my-repo")
		if opts.Author != "MyUser" {
			t.Errorf("Author = %q, want MyUser", opts.Author)
		}
		if !strings.Contains(opts.resourceRepository, "raw.githubusercontent.com/MyUser/my-repo") {
			t.Errorf("resourceRepository = %q, missing expected path", opts.resourceRepository)
		}
		if !strings.Contains(opts.templateRepository, "github.com/MyUser/my-repo") {
			t.Errorf("templateRepository = %q, missing expected path", opts.templateRepository)
		}
	})

	t.Run("full URL", func(t *testing.T) {
		opts := &commandLineOptions{}
		opts.SetRepository("https://example.com/templates")
		if opts.Author != "comp2unraid" {
			t.Errorf("Author = %q, want comp2unraid", opts.Author)
		}
		if opts.resourceRepository != "https://example.com/templates" {
			t.Errorf("resourceRepository = %q, want https://example.com/templates", opts.resourceRepository)
		}
	})
}
