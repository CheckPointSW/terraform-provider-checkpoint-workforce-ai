# DLP Data Types

DLP (Data Loss Prevention) data types define patterns of sensitive information — such as credit card numbers, social security numbers, or custom patterns defined by your organization. 
These are referenced in the `data_types` field of chats rule policies.

## Usage in Terraform

```hcl
policy = jsonencode({
  event_type = "file_upload"
  action     = "prevent"
  logging    = "enabled"
  services_and_application = {
    mode     = "selected"
    category = [{ category_id = 60531762 }] # Generative AI Tools: any AI app
  }
  data_types = [
    {
      id   = "cf0523c1-537e-4a4b-8bb8-084b7b9e0b45"
      name = "Credit Card Number"
      type = "PRE_DEFINED"
    }
  ]
})
```

## Managing Data Types

Data types are managed at the tenant level. You can explore the available predefined data types (e.g., Credit Card Number, SSN, IBAN) and custom data types created by your organization through the [DLP Datatypes API](https://app.swaggerhub.com/apis/Check-Point/checkpoint-ai-security/1.0.0#/DLP%20Datatypes).
