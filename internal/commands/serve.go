package commands

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
)

type Config struct {
	skipTlsVerify  bool
	upstreamScheme string
	upstreamHost   string
	allowedHosts   []string
	clientID       string
	clientSecret   string
	tokenURL       string
	scopes         []string
	listenPort     int
}

var Serve = &cobra.Command{
	Use:   "serve",
	Short: "Serves the proxy",
	Run: func(cmd *cobra.Command, args []string) {
		config := MergeConfig()

		// Create a client credentials Config from the Viper configuration.
		oauth := &clientcredentials.Config{
			ClientID:     config.clientID,
			ClientSecret: config.clientSecret,
			TokenURL:     config.tokenURL,
			Scopes:       config.scopes,
		}

		// Create a background context for the OAuth client to keep it alive.
		oauthContext := context.Background()

		// Create a reusable token source to store valid tokens obtained from
		// the authorization server.
		tokenSource := oauth.TokenSource(oauthContext)

		// Create a request director to pass to the reverse proxy. The director
		// is responsible for tampering with the request before dispatching it.
		director := func(req *http.Request) {
			host, scheme := ExtractHost(
				&req.Header,
				config,
			)

			req.Host = host
			req.URL.Host = host
			req.URL.Scheme = scheme

			// Fetch a token from the token source. This will reuse any valid
			// token still available in the source.
			token, err := tokenSource.Token()

			// Catch OAuth errors and log them, but don't crash the process
			if err != nil {
				log.Printf("OAuth error: %s\n", err)

				return
			}

			// Set the Authorization header to "Bearer <token>"
			req.Header.Set(
				"Authorization",
				fmt.Sprintf("Bearer %s", token.AccessToken),
			)

			log.Printf("Forwarding request to %s://%s\n", scheme, host)
		}

		// Create a reverse proxy
		proxy := &httputil.ReverseProxy{
			Director: director,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: config.skipTlsVerify,
				},
			},
		}

		// Start listening
		fmt.Printf("Listening on 0.0.0.0:%d\n", config.listenPort)
		log.Fatal(http.ListenAndServe(
			":"+strconv.Itoa(config.listenPort),
			proxy,
		))
	},
}

func MergeConfig() Config {
	viper.SetDefault("listen_port", 8080)
	viper.SetDefault("skip_tls_verify", false)
	viper.SetDefault("upstream_scheme", "https")
	viper.SetDefault("upstream_host", "")
	viper.SetDefault("allowed_hosts", []string{})

	return Config{
		skipTlsVerify:  viper.GetBool("skip_tls_verify"),
		upstreamScheme: viper.GetString("upstream_scheme"),
		upstreamHost:   viper.GetString("upstream_host"),
		allowedHosts:   viper.GetStringSlice("allowed_hosts"),
		listenPort:     viper.GetInt("listen_port"),
		clientID:       viper.GetString("client_id"),
		clientSecret:   viper.GetString("client_secret"),
		tokenURL:       viper.GetString("token_endpoint"),
		scopes:         viper.GetStringSlice("scopes"),
	}
}

func ExtractHost(headers *http.Header, config Config) (string, string) {

	// Ensure hostname and scheme are set properly, see:
	// https://stackoverflow.com/a/23166390/2532203
	scheme := "https"

	if config.upstreamScheme != "" {
		scheme = config.upstreamScheme
	}

	host := ""

	hostHeader := headers.Get("Host")

	if hostHeader != "" {
		host = hostHeader
	}

	if config.upstreamHost != "" {
		host = config.upstreamHost
	}

	forwardHeader := headers.Get("Forward")

	if forwardHeader != "" {
		forwardUrl, err := url.Parse(forwardHeader)

		if err == nil {
			host = forwardUrl.Host
			scheme = forwardUrl.Scheme
		}
	}

	forwardProtoHeader := headers.Get("Forward-Proto")

	if forwardProtoHeader != "" {
		scheme = forwardProtoHeader
	}

	// If the allowed hosts list is used (has at least one entry), check whether
	// the list contains the resolved hostname. We only set the upstream hostname if it
	// is allowed.
	if len(config.allowedHosts) > 0 && !contains(config.allowedHosts, host) {
		log.Println("Blocked disallowed host", host)

		return "", scheme
	}

	return host, scheme
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
