package clickhouse

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceClickhouse() *schema.Resource {
	return &schema.Resource{
		ReadContext: schemautil.DatasourceServiceRead,
		Description: "The Clickhouse data source provides information about the existing Aiven Clickhouse service.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(clickhouseSchema(), "project", "service_name"),
	}
}
