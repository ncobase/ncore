package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSearchEngineConfigs_PreferDataSearchNamespace(t *testing.T) {
	v := viper.New()

	v.Set("data.elasticsearch.addresses", []string{"http://legacy:9200"})
	v.Set("data.elasticsearch.username", "legacy-user")
	v.Set("data.elasticsearch.password", "legacy-pass")

	v.Set("data.search.elasticsearch.addresses", []string{"http://search:9200"})
	v.Set("data.search.elasticsearch.username", "search-user")
	v.Set("data.search.elasticsearch.password", "search-pass")

	es := getElasticsearchConfigs(v)
	if len(es.Addresses) != 1 || es.Addresses[0] != "http://search:9200" {
		t.Fatalf("expected addresses from data.search.elasticsearch, got %v", es.Addresses)
	}
	if es.Username != "search-user" || es.Password != "search-pass" {
		t.Fatalf("expected credentials from data.search.elasticsearch, got %q/%q", es.Username, es.Password)
	}
}

func TestSearchEngineConfigs_FallbackToLegacyNamespace(t *testing.T) {
	v := viper.New()

	v.Set("data.elasticsearch.addresses", []string{"http://legacy:9200"})
	v.Set("data.elasticsearch.username", "legacy-user")
	v.Set("data.elasticsearch.password", "legacy-pass")

	es := getElasticsearchConfigs(v)
	if len(es.Addresses) != 1 || es.Addresses[0] != "http://legacy:9200" {
		t.Fatalf("expected addresses from data.elasticsearch, got %v", es.Addresses)
	}
	if es.Username != "legacy-user" || es.Password != "legacy-pass" {
		t.Fatalf("expected credentials from data.elasticsearch, got %q/%q", es.Username, es.Password)
	}
}

func TestOpenSearchConfigs_InsecureSkipTLSPreferSearchNamespace(t *testing.T) {
	v := viper.New()

	v.Set("data.opensearch.insecure_skip_tls", false)
	v.Set("data.search.opensearch.insecure_skip_tls", true)

	os := getOpenSearchConfigs(v)
	if os.InsecureSkipTLS != true {
		t.Fatalf("expected insecure_skip_tls from data.search.opensearch, got %v", os.InsecureSkipTLS)
	}
}

