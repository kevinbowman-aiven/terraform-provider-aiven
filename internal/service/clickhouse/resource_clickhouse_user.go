package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenClickhouseUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The actual name of the Clickhouse user.").ForceNew().Build(),
	},
	"password": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "The password of the clickhouse user.",
	},
	"uuid": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "UUID of the clickhouse user.",
	},
	"required": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates if a clickhouse user is required",
	},
}

func ResourceClickhouseUser() *schema.Resource {
	return &schema.Resource{
		Description:        "The Clickhouse User resource allows the creation and management of Aiven Clikhouse Users.",
		DeprecationMessage: betaDeprecationMessage,
		CreateContext:      resourceClickhouseUserCreate,
		ReadContext:        resourceClickhouseUserRead,
		DeleteContext:      resourceClickhouseUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClickhouseUserState,
		},

		Schema: aivenClickhouseUserSchema,
	}
}

func resourceClickhouseUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	u, err := client.ClickhouseUser.Create(
		projectName,
		serviceName,
		username,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, u.User.UUID))

	if err := d.Set("password", u.User.Password); err != nil {
		return diag.FromErr(err)
	}

	return resourceClickhouseUserRead(ctx, d, m)
}

func resourceClickhouseUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, uuid := schemautil.SplitResourceID3(d.Id())
	user, err := client.ClickhouseUser.Get(projectName, serviceName, uuid)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", user.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("uuid", user.UUID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("required", user.Required); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceClickhouseUserDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, uuid := schemautil.SplitResourceID3(d.Id())
	err := client.ClickhouseUser.Delete(projectName, serviceName, uuid)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceClickhouseUserState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<uuid>", d.Id())
	}

	di := resourceClickhouseUserRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get clickhouse user: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
