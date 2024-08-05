package additionaldisk

import (

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceAdditionalDisk() *schema.Resource {
	return &schema.Resource{
		Create: Create,
		Read:   Read,
		Update: Update,
		Delete: Delete,

		Schema: Schema(),
	}
}
