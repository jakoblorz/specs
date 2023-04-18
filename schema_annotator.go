package specs

import (
	"fmt"
	"log"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
)

// defaultParentSchemaAnnotatorMap is a map of annotator funcs that apply a certain validation rule to a field's parent openapi3.Schema
// Keys are taken from here: https://github.com/go-playground/validator/blob/b43d437012ec5766eee3a068f53c6581f8e64282/baked_in.go#L72
var defaultParentSchemaAnnotatorMap = map[string]ParentSchemaAnnotatorFunc{
	"required": requiredAnnotator,
}

type ParentSchemaAnnotatorFunc func(field *Field, schema *openapi3.Schema)

func requiredAnnotator(field *Field, schema *openapi3.Schema) {
	if schema.Required != nil {
		schema.Required = []string{}
	}
	schema.Required = append(schema.Required, field.Name)
}

// defaultSchemaAnnotatorMap is a map of annotator funcs that apply a certain validation rule to an openapi3.Schema
// Keys are taken from here: https://github.com/go-playground/validator/blob/b43d437012ec5766eee3a068f53c6581f8e64282/baked_in.go#L72
// The keys were chosen mostly during a quick scan regarding what could be expressed in openapi schemas, either using a regex,
// enums or other instructions
var defaultSchemaAnnotatorMap = map[string]SchemaAnnotatorFunc{
	"min":                           minAnnotator,
	"max":                           maxAnnotator,
	"eq":                            warnAnnotator,
	"eq_ignore_case":                warnAnnotator,
	"ne":                            warnAnnotator,
	"ne_ignore_case":                warnAnnotator,
	"lt":                            ltAnnotator,
	"lte":                           lteAnnotator,
	"gt":                            gtAnnotator,
	"gte":                           gteAnnotator,
	"alpha":                         warnAnnotator,
	"alphanum":                      warnAnnotator,
	"alphaunicode":                  warnAnnotator,
	"alphanumunicode":               warnAnnotator,
	"boolean":                       noopAnnotator,
	"numeric":                       warnAnnotator,
	"number":                        warnAnnotator,
	"hexadecimal":                   warnAnnotator,
	"hexcolor":                      warnAnnotator,
	"rgb":                           warnAnnotator,
	"rgba":                          warnAnnotator,
	"hsl":                           warnAnnotator,
	"hsla":                          warnAnnotator,
	"e164":                          warnAnnotator,
	"email":                         warnAnnotator,
	"url":                           warnAnnotator,
	"http_url":                      warnAnnotator,
	"uri":                           warnAnnotator,
	"urn_rfc2141":                   warnAnnotator, // RFC 2141
	"file":                          warnAnnotator,
	"filepath":                      warnAnnotator,
	"base64":                        warnAnnotator,
	"base64url":                     warnAnnotator,
	"base64rawurl":                  warnAnnotator,
	"contains":                      warnAnnotator,
	"containsany":                   warnAnnotator,
	"containsrune":                  warnAnnotator,
	"excludes":                      warnAnnotator,
	"excludesall":                   warnAnnotator,
	"excludesrune":                  warnAnnotator,
	"startswith":                    warnAnnotator,
	"endswith":                      warnAnnotator,
	"startsnotwith":                 warnAnnotator,
	"endsnotwith":                   warnAnnotator,
	"isbn":                          warnAnnotator,
	"isbn10":                        warnAnnotator,
	"isbn13":                        warnAnnotator,
	"uuid":                          warnAnnotator,
	"uuid3":                         warnAnnotator,
	"uuid4":                         warnAnnotator,
	"uuid5":                         warnAnnotator,
	"uuid_rfc4122":                  warnAnnotator,
	"uuid3_rfc4122":                 warnAnnotator,
	"uuid4_rfc4122":                 warnAnnotator,
	"uuid5_rfc4122":                 warnAnnotator,
	"ulid":                          warnAnnotator,
	"md4":                           warnAnnotator,
	"md5":                           warnAnnotator,
	"sha256":                        warnAnnotator,
	"sha384":                        warnAnnotator,
	"sha512":                        warnAnnotator,
	"ripemd128":                     warnAnnotator,
	"ripemd160":                     warnAnnotator,
	"tiger128":                      warnAnnotator,
	"tiger160":                      warnAnnotator,
	"tiger192":                      warnAnnotator,
	"ascii":                         warnAnnotator,
	"printascii":                    warnAnnotator,
	"multibyte":                     warnAnnotator,
	"datauri":                       warnAnnotator,
	"latitude":                      warnAnnotator,
	"longitude":                     warnAnnotator,
	"ssn":                           warnAnnotator,
	"ipv4":                          warnAnnotator,
	"ipv6":                          warnAnnotator,
	"ip":                            warnAnnotator,
	"cidrv4":                        warnAnnotator,
	"cidrv6":                        warnAnnotator,
	"cidr":                          warnAnnotator,
	"tcp4_addr":                     warnAnnotator,
	"tcp6_addr":                     warnAnnotator,
	"tcp_addr":                      warnAnnotator,
	"udp4_addr":                     warnAnnotator,
	"udp6_addr":                     warnAnnotator,
	"udp_addr":                      warnAnnotator,
	"ip4_addr":                      warnAnnotator,
	"ip6_addr":                      warnAnnotator,
	"ip_addr":                       warnAnnotator,
	"unix_addr":                     warnAnnotator,
	"mac":                           warnAnnotator,
	"hostname":                      warnAnnotator, // RFC 952
	"hostname_rfc1123":              warnAnnotator, // RFC 1123
	"fqdn":                          warnAnnotator,
	"unique":                        warnAnnotator,
	"oneof":                         warnAnnotator,
	"html":                          warnAnnotator,
	"html_encoded":                  warnAnnotator,
	"url_encoded":                   warnAnnotator,
	"dir":                           warnAnnotator,
	"dirpath":                       warnAnnotator,
	"json":                          warnAnnotator,
	"jwt":                           warnAnnotator,
	"hostname_port":                 warnAnnotator,
	"lowercase":                     warnAnnotator,
	"uppercase":                     warnAnnotator,
	"datetime":                      warnAnnotator,
	"timezone":                      warnAnnotator,
	"iso3166_1_alpha2":              warnAnnotator,
	"iso3166_1_alpha3":              warnAnnotator,
	"iso3166_1_alpha_numeric":       warnAnnotator,
	"iso3166_2":                     warnAnnotator,
	"iso4217":                       warnAnnotator,
	"iso4217_numeric":               warnAnnotator,
	"bcp47_language_tag":            warnAnnotator,
	"postcode_iso3166_alpha2":       warnAnnotator,
	"postcode_iso3166_alpha2_field": warnAnnotator,
	"bic":                           warnAnnotator,
	"semver":                        warnAnnotator,
	"dns_rfc1035_label":             warnAnnotator,
	"credit_card":                   warnAnnotator,
	"cve":                           warnAnnotator,
	"luhn_checksum":                 warnAnnotator,
	"mongodb":                       warnAnnotator,
	"cron":                          warnAnnotator,
}

type SchemaAnnotatorFunc func(fieldTag *FieldTag, schema *openapi3.Schema)

var (
	warnAnnotator = func(fieldTag *FieldTag, schema *openapi3.Schema) {
		log.Printf("warn: %s operator is implemented using the warnAnnotator. Please implement the appropriate annotator using similar logic as in https://github.com/go-playground/validator/blob/b43d437012ec5766eee3a068f53c6581f8e64282/baked_in.go#L72", fieldTag.Operator)
	}
	noopAnnotator = func(fieldTag *FieldTag, schema *openapi3.Schema) {}
)

func minAnnotator(fieldTag *FieldTag, schema *openapi3.Schema) {
	gteAnnotator(fieldTag, schema)
}

func gteAnnotator(fieldTag *FieldTag, schema *openapi3.Schema) {
	i, err := strconv.ParseInt(fieldTag.Param, 10, 64)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s as int64: %w", fieldTag.Param, err))
	}
	f := float64(i)
	schema.Min = &f
}

func gtAnnotator(fieldTag *FieldTag, schema *openapi3.Schema) {
	gteAnnotator(fieldTag, schema)
	schema.ExclusiveMin = true
}

func maxAnnotator(fieldTag *FieldTag, schema *openapi3.Schema) {
	lteAnnotator(fieldTag, schema)
}

func lteAnnotator(fieldTag *FieldTag, schema *openapi3.Schema) {
	i, err := strconv.ParseInt(fieldTag.Param, 10, 64)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s as int64: %w", fieldTag.Param, err))
	}
	f := float64(i)
	schema.Max = &f
}

func ltAnnotator(fieldTag *FieldTag, schema *openapi3.Schema) {
	lteAnnotator(fieldTag, schema)
	schema.ExclusiveMax = true
}
