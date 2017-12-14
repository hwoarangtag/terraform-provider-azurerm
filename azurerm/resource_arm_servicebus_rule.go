package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/arm/servicebus"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmServiceBusRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmServiceBusRuleCreate,
		Read:   resourceArmServiceBusRuleRead,
		Update: resourceArmServiceBusRuleCreate,
		Delete: resourceArmServiceBusRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"topic_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"filtertype": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"sqlexpression": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"location": deprecatedLocationSchema(),

			"subscription_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_group_name": resourceGroupNameSchema(),

			"auto_delete_on_idle": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"default_message_ttl": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"lock_duration": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"dead_lettering_on_message_expiration": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"enable_batched_operations": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"max_delivery_count": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"requires_session": {
				Type:     schema.TypeBool,
				Optional: true,
				// cannot be modified
				ForceNew: true,
			},

			// TODO: remove in the next major version
			"dead_lettering_on_filter_evaluation_exceptions": {
				Type:       schema.TypeBool,
				Optional:   true,
				Deprecated: "This field has been deprecated by Azure",
			},
		},
	}
}

func resourceArmServiceBusRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).serviceBusRulesClient
	log.Printf("[INFO] preparing arguments for Azure ARM ServiceBus Rule creation.")

	name := d.Get("name").(string)
	topicName := d.Get("topic_name").(string)
	subscriptionName := d.Get("subscription_name").(string)
	namespaceName := d.Get("namespace_name").(string)
	resGroup := d.Get("resource_group_name").(string)

	filterType := d.Get("filtertype").(string)
	sqlExpression := d.Get("sqlexpression").(string)

	parameters := servicebus.Rule{
		Ruleproperties: &servicebus.Ruleproperties{
			FilterType: servicebus.FilterType(filterType),
			SQLFilter: &servicebus.SQLFilter{
				SQLExpression: &sqlExpression,
			},
		},
	}

	_, err := client.CreateOrUpdate(resGroup, namespaceName, topicName, subscriptionName, name, parameters)
	if err != nil {
		return err
	}

	read, err := client.Get(resGroup, namespaceName, topicName, subscriptionName, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read ServiceBus Rule %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmServiceBusRuleRead(d, meta)
}

func resourceArmServiceBusRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).serviceBusRulesClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	namespaceName := id.Path["namespaces"]
	topicName := id.Path["topics"]
	name := id.Path["rules"]
	subscriptionName := id.Path["subscriptions"]

	log.Printf("[INFO] RuleID: %s, args: %s, %s, %s, %s", d.Id(), resGroup, namespaceName, topicName, name)

	resp, err := client.Get(resGroup, namespaceName, topicName, subscriptionName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure ServiceBus Rule %s: %+v", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	d.Set("namespace_name", namespaceName)
	d.Set("topic_name", topicName)

	if props := resp.Ruleproperties; props != nil {
		//TBD
	}

	return nil
}

func resourceArmServiceBusRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).serviceBusRulesClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	namespaceName := id.Path["namespaces"]
	topicName := id.Path["topics"]
	name := id.Path["rules"]
	subscriptionName := id.Path["subscriptions"]

	_, err = client.Delete(resGroup, namespaceName, topicName, subscriptionName, name)

	return err
}
