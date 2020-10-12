package commands

import (
	"net/http"
	"testing"
)

func TestExtractHost(t *testing.T) {
	type args struct {
		headers *http.Header
		config  Config
	}
	tests := []struct {
		name     string
		args     args
		hostname string
		scheme   string
	}{
		{
			name: "Uses upstream hostname and scheme",
			args: args{
				headers: &http.Header{
					"Host": []string{"foo.bar.com"},
				},
				config: Config{
					upstreamHost:   "www.example.com",
					upstreamScheme: "https",
				},
			},
			hostname: "www.example.com",
			scheme:   "https",
		},
		{
			name: "Uses upstream hostname and forward-proto scheme",
			args: args{
				headers: &http.Header{
					"Forward-Proto": []string{"http"},
				},
				config: Config{
					upstreamHost:   "www.example.com",
					upstreamScheme: "https",
				},
			},
			hostname: "www.example.com",
			scheme:   "http",
		},
		{
			name: "Uses forward hostname and forward scheme",
			args: args{
				headers: &http.Header{
					"Forward": []string{"http://foo.example.org"},
				},
				config: Config{
					upstreamHost:   "www.example.com",
					upstreamScheme: "https",
				},
			},
			hostname: "foo.example.org",
			scheme:   "http",
		},
		{
			name: "Uses forward hostname and forward-proto scheme",
			args: args{
				headers: &http.Header{
					"Forward":       []string{"http://foo.example.org"},
					"Forward-Proto": []string{"tcp"},
				},
				config: Config{
					upstreamHost: "www.example.com",
				},
			},
			hostname: "foo.example.org",
			scheme:   "tcp",
		},
		{
			name: "Uses hostname header as fallback",
			args: args{
				headers: &http.Header{
					"Host": []string{"foo.bar.com"},
				},
				config: Config{},
			},
			hostname: "foo.bar.com",
			scheme:   "https",
		},
		{
			name: "Returns blank host if not in allowed list",
			args: args{
				headers: &http.Header{
					"Host": []string{"www.example.io"},
				},
				config: Config{
					allowedHosts: []string{
						"www.example.com",
						"www.example.org",
					},
				},
			},
			hostname: "",
			scheme:   "https",
		},
		{
			name: "Returns blank host if upstream host not in allowed list",
			args: args{
				headers: &http.Header{},
				config: Config{
					allowedHosts: []string{
						"www.example.com",
						"www.example.org",
					},
					upstreamHost: "www.example.io",
				},
			},
			hostname: "",
			scheme:   "https",
		},
		{
			name: "Returns blank host if forward host not in allowed list",
			args: args{
				headers: &http.Header{
					"Forward": []string{"https://www.example.io"},
				},
				config: Config{
					allowedHosts: []string{
						"www.example.com",
						"www.example.org",
					},
				},
			},
			hostname: "",
			scheme:   "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostname, scheme := ExtractHost(tt.args.headers, tt.args.config)

			if hostname != tt.hostname {
				t.Errorf("ExtractHost() actual = %v, expected %v", hostname, tt.hostname)
			}

			if scheme != tt.scheme {
				t.Errorf("ExtractHost() actual = %v, expected %v", scheme, tt.scheme)
			}
		})
	}
}
