package kafka

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/service/kafka/cache"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenKafkaACLSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice([]string{"admin", "read", "readwrite", "write"}, false),
		Description:  schemautil.Complex("Kafka permission to grant.").ForceNew().PossibleValues("admin", "read", "readwrite", "write").Build(),
	},
	"topic": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("Topic name pattern for the ACL entry.").ForceNew().Build(),
	},
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(\*$|[a-zA-Z0-9-_?][a-zA-Z0-9-_?*]+)$`), "username should be alphanumeric"),
		Description:  schemautil.Complex("Username pattern for the ACL entry.").ForceNew().Build(),
	},
}

func ResourceKafkaACL() *schema.Resource {
	return &schema.Resource{
		Description:   "The Resource Kafka ACL resource allows the creation and management of ACLs for an Aiven Kafka service.",
		CreateContext: resourceKafkaACLCreate,
		ReadContext:   resourceKafkaACLRead,
		DeleteContext: resourceKafkaACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKafkaACLState,
		},

		Schema: aivenKafkaACLSchema,
	}
}

func resourceKafkaACLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	acl, err := client.KafkaACLs.Create(
		project,
		serviceName,
		aiven.CreateKafkaACLRequest{
			Permission: d.Get("permission").(string),
			Topic:      d.Get("topic").(string),
			Username:   d.Get("username").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, acl.ID))

	return resourceKafkaACLRead(ctx, d, m)
}

func resourceKafkaACLRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, aclID := schemautil.SplitResourceID3(d.Id())
	acl, err := cache.ACLCache{}.Read(project, serviceName, aclID, client)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = copyKafkaACLPropertiesFromAPIResponseToTerraform(d, &acl, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaACLDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, aclID := schemautil.SplitResourceID3(d.Id())
	err := client.KafkaACLs.Delete(projectName, serviceName, aclID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaACLState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<acl_id>", d.Id())
	}

	di := resourceKafkaACLRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get kafka acl: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func copyKafkaACLPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	acl *aiven.KafkaACL,
	project string,
	serviceName string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("topic", acl.Topic); err != nil {
		return err
	}
	if err := d.Set("permission", acl.Permission); err != nil {
		return err
	}
	if err := d.Set("username", acl.Username); err != nil {
		return err
	}

	return nil
}
