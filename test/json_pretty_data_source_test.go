package test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJSONPrettyDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "test_json_pretty" "basic" {
  json = "{\"name\":\"John\",\"age\":30}"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_json_pretty.basic", "id"),
					resource.TestCheckResourceAttr("data.test_json_pretty.basic", "json", "{\"name\":\"John\",\"age\":30}"),
					resource.TestCheckResourceAttr("data.test_json_pretty.basic", "result", "{\n  \"age\": 30,\n  \"name\": \"John\"\n}"),
				),
			},
		},
	})
}

func TestAccJSONPrettyDataSource_CustomIndent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "test_json_pretty" "custom_indent" {
  json = "{\"name\":\"John\",\"age\":30}"
  indent = 4
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_json_pretty.custom_indent", "id"),
					resource.TestCheckResourceAttr("data.test_json_pretty.custom_indent", "json", "{\"name\":\"John\",\"age\":30}"),
					resource.TestCheckResourceAttr("data.test_json_pretty.custom_indent", "result", "{\n    \"age\": 30,\n    \"name\": \"John\"\n}"),
				),
			},
		},
	})
}

func TestAccJSONPrettyDataSource_WithNewline(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "test_json_pretty" "with_newline" {
  json = "{\"name\":\"John\",\"age\":30}"
  newline = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_json_pretty.with_newline", "id"),
					resource.TestCheckResourceAttr("data.test_json_pretty.with_newline", "json", "{\"name\":\"John\",\"age\":30}"),
					resource.TestCheckResourceAttr("data.test_json_pretty.with_newline", "result", "{\n  \"age\": 30,\n  \"name\": \"John\"\n}\n"),
				),
			},
		},
	})
}

func TestAccJSONPrettyDataSource_WithBothParams(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "test_json_pretty" "both_params" {
  json = "{\"name\":\"John\",\"age\":30}"
  indent = 3
  newline = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_json_pretty.both_params", "id"),
					resource.TestCheckResourceAttr("data.test_json_pretty.both_params", "json", "{\"name\":\"John\",\"age\":30}"),
					resource.TestCheckResourceAttr("data.test_json_pretty.both_params", "result", "{\n   \"age\": 30,\n   \"name\": \"John\"\n}\n"),
				),
			},
		},
	})
}

func TestAccJSONPrettyDataSource_Complex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "test_json_pretty" "complex" {
  json = "{\"users\":[{\"id\":1,\"name\":\"Alice\",\"roles\":[\"admin\",\"user\"]},{\"id\":2,\"name\":\"Bob\",\"roles\":[\"user\"]}],\"total\":2,\"metadata\":{\"created\":\"2024-01-01\",\"version\":\"1.0\"}}"
  indent = 2
  newline = false
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_json_pretty.complex", "id"),
					resource.TestCheckResourceAttrSet("data.test_json_pretty.complex", "result"),
				),
			},
		},
	})
}

func TestAccJSONPrettyDataSource_InvalidJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "test_json_pretty" "invalid" {
  json = "invalid json"
}
`,
				ExpectError: regexp.MustCompile(`Invalid JSON string`),
			},
		},
	})
}
