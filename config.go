package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/hashicorp/hcl"
	"github.com/sirupsen/logrus"
)

type Site struct {
	Host          string  `json:"host"`
	AWSKey        string  `json:"awsKey"`
	AWSSecret     string  `json:"awsSecret"`
	AWSRegion     string  `json:"awsRegion"`
	AWSBucket     string  `json:"awsBucket"`
	AWSBucketPath string  `json:"awsBucketPath"`
	Users         []User  `json:"users"`
	Options       Options `json:"options"`
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Options struct {
	CORS     bool   `json:"cors"`
	Gzip     bool   `json:"gzip"`
	Website  bool   `json:"website"`
	Prefix   string `json:"prefix"`
	ForceSSL bool   `json:"forceSsl"`
	Proxied  bool   `json:"proxied"`
}

func (o Options) String() string {
	return fmt.Sprintf("[cors: %t]"+
		"[gzip: %t]"+
		"[website: %t]"+
		"[prefix: %s]"+
		"[ssl: %t]"+
		"[proxied: %t]",
		o.CORS,
		o.Gzip,
		o.Website,
		o.Prefix,
		o.ForceSSL,
		o.Proxied)
}

type sitesCfg []Site

func ConfiguredProxyHandler(configFile *os.File) (http.Handler, error) {
	var config []Site = make([]Site, 0)
	bytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		logrus.WithField("file", configFile.Name()).Errorf("Failed to read config")
	}

	err = hcl.Decode(&config, string(bytes))
	if err != nil {
		logrus.WithField("error", err).Error("Error decoding config")
	}

	logrus.WithField("count", len(config)).Info("Read sites")

	return createMulti(config)
}

func createMulti(sites []Site) (http.Handler, error) {
	var err error

	handler := NewHostDispatchingHandler()

	for i, site := range sites {
		err = site.validateWithHost()

		if err != nil {
			msg := fmt.Sprintf("%v in configuration at position %d", err, i)
			return nil, errors.New(msg)
		}

		logrus.WithFields(logrus.Fields{
			"host":    site.Host,
			"options": site.Options,
		}).Debug("Configuring site handlers")
		handler.HandleHost(site.Host, createSiteHandler(site))
	}

	return handler, nil
}

func createSiteHandler(s Site) http.Handler {
	var handler http.Handler

	proxy := NewS3Proxy(s.AWSKey, s.AWSSecret, s.AWSRegion, s.AWSBucket)
	handler = NewProxyHandler(proxy, s.Options.Prefix)

	if s.Options.Website {
		cfg, err := proxy.GetWebsiteConfig()
		if err != nil {
			fmt.Printf("warning: site for bucket %s configured with "+
				"website option but received error when retrieving "+
				"website config\n\t%v", s.AWSBucket, err)
		} else {
			handler = NewWebsiteHandler(handler, cfg)
		}
	}

	if s.Options.CORS {
		handler = corsHandler(handler)
	}

	if s.Options.Gzip {
		handler = handlers.CompressHandler(handler)
	}

	if len(s.Users) > 0 {
		handler = NewBasicAuthHandler(s.Users, handler)
	} else {
		fmt.Printf("warning: site for bucket %s has no configured users\n", s.AWSBucket)
	}

	if s.Options.ForceSSL {
		handler = NewSSLRedirectHandler(handler)
	}

	if s.Options.Proxied {
		handler = handlers.ProxyHeaders(handler)
	}

	return handler
}

func corsHandler(next http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{"*"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"HEAD", "GET", "OPTIONS"}),
	)(next)
}

func parseUsers(us string) ([]User, error) {
	if us == "" {
		return []User{}, nil
	}

	pairs := strings.Split(us, ",")
	users := make([]User, len(pairs))

	for i, p := range pairs {
		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			msg := fmt.Sprintf("Failed to parse user %s at position %d", p, i)
			return nil, errors.New(msg)
		}

		users[i] = User{
			Name:     parts[0],
			Password: parts[1],
		}
	}

	return users, nil
}

func (s Site) validateWithHost() error {
	if s.Host == "" {
		return errors.New("Host not specified")
	}

	return s.validate()
}

func (s Site) validate() error {
	if s.AWSKey == "" {
		return errors.New("AWS Key not specified")
	}

	if s.AWSSecret == "" {
		return errors.New("AWS Secret not specified")
	}

	if s.AWSRegion == "" {
		return errors.New("AWS Region not specified")
	}

	if s.AWSBucket == "" {
		return errors.New("AWS Bucket not specified")
	}

	return nil
}
