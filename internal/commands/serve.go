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

var Serve = &cobra.Command{
	Use:   "serve",
	Short: "Serves the proxy",
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("listen_port", 8080)
		proxyPort := viper.GetInt("listen_port")

		if proxyPort < 0 || proxyPort > 65_535 {
			log.Fatalf("Listen port '%d' is invalid", proxyPort)
		}

		viper.SetDefault("skip_tls_verify", false)
		skipTlsVerify := viper.GetBool("skip_tls_verify")

		viper.SetDefault("upstream_scheme", "https")
		upstreamScheme := viper.GetString("upstream_scheme")

		viper.SetDefault("upstream_host", "")
		upstreamHost := viper.GetString("upstream_host")

		viper.SetDefault("allowed_hosts", []string{})
		allowedHosts := viper.GetStringSlice("allowed_hosts")

		// Create a client credentials config from the Viper configuration.
		config := &clientcredentials.Config{
			ClientID:     viper.GetString("client_id"),
			ClientSecret: viper.GetString("client_secret"),
			TokenURL:     viper.GetString("token_endpoint"),
			Scopes:       viper.GetStringSlice("scopes"),
		}

		// Create a background context for the OAuth client to keep it alive.
		oauthContext := context.Background()

		// Create a reusable token source to store valid tokens obtained from
		// the authorization server.
		tokenSource := config.TokenSource(oauthContext)

		// Create a request director to pass to the reverse proxy. The director
		// is responsible for tampering with the request before dispatching it.
		director := func(req *http.Request) {

			// Ensure host and scheme are set properly, see:
			// https://stackoverflow.com/a/23166390/2532203
			req.URL.Scheme = upstreamScheme
			host := req.Host

			forwardHeader := req.Header.Get("Forward")

			if forwardHeader != "" {
				forwardUrl, err := url.Parse(forwardHeader)

				if err == nil {
					host = forwardUrl.Host
					req.URL.Scheme = forwardUrl.Scheme
				}
			}

			if host == "" && upstreamHost != "" {
				host = upstreamHost
			}

			if host == "" {
				host = req.Host
			}

			// If the allowed hosts list is used (has at least one entry), check
			// whether the list contains the resolved host. We only set the
			// upstream host if it is allowed.
			if len(allowedHosts) > 0 && contains(allowedHosts, host) {
				req.Host = host
			}

			// Fetch a token from the token source. This will reuse any valid
			// token still available in the source.
			token, err := tokenSource.Token()

			// Catch OAuth errors and log them, but don't crash the process
			if err != nil {
				log.Printf("OAuth error: %s\n", err)

				return
			}

			// Set the Authorization header to "Bearer <token>"
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

			log.Printf("Forwarding request to %s\n", req.URL)
		}

		// Create a reverse proxy
		proxy := &httputil.ReverseProxy{
			Director: director,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipTlsVerify,
				},
			},
		}

		// Start listening
		fmt.Printf("Listening on 0.0.0.0:%d\n", proxyPort)
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(proxyPort), proxy))
	},
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
