## The Problem
Well, we already know how to create microservices and the like, but there is a problem: all of them are currently running on your localhost. None of them can be reached outside of it, and the moment you turn off your computer, they stop working, meaning they are no longer doing their jobs—which is something we clearly don't want. The main problem is simple: we don't have a way to keep the microservice running constantly, and it is fragily connected to our local machine. This is what we fix via `opentofu`, which lets us write IaC (Infrastructure as Code) via Terraform using the HCL method. Thanks to it, we avoid logging into our provider (in my case, Azure) and clicking random buttons in the UI, which takes plenty of time. We simply write IaC and achieve the same results, just much faster. Let me explain the differences and details:
- **IaC + HCL:** IaC is a method; instead of using the UI, we use IaC (which simply means writing our infrastructure using code). Meanwhile, HCL (HashiCorp Configuration Language) is the syntax we use to write the code (through Terraform).
- **opentofu:** Think of this as the CLI tool for all of it. A simple mental model is to think of IaC and HCL as the `schema.sql`, while `opentofu` is `atlas`. See? They complement each other.

## Example Code and Explanation

### The `terraform` Block
The `terraform {}` block is the global configuration for OpenTofu itself. It does not build or touch any infrastructure in Azure. Instead, it acts as a setup list that tells OpenTofu which engine version to use and which external plugins it needs to download before doing any real work.
```hcl
terraform {
    required_version = "~> 1.6"
    required_providers {
        azurerm = {
            source  = "hashicorp/azurerm"
            version = "~> 3.0"
        }
    }
}
```

#### The Phases
* **Phase 1:** Before doing anything, we declare the exact version of OpenTofu we expect to run. The `~>` symbol means "you can wander wherever you want within this family, but nowhere below or above." By writing `~> 1.6`, you are telling OpenTofu it is allowed to use `1.6.0` up to `1.6.99`, but it is strictly forbidden from jumping to `1.7` or dropping below `1.6`. This prevents accidental breaking changes.
* **Phase 2:** Because OpenTofu does not know how to talk to Azure out of the box, we use this block to list the external translation plugins it needs to download. Remember, this phase doesn't run anything—it just downloads the necessary tools.
* **Phase 3:** Inside the shopping list, we have to hard-lock ourselves to a specific provider:
* **The Name (`azurerm = {}`):** This is just a custom variable name you choose to store the provider. You could technically name it anything, but `azurerm` is the standard. You will use this variable name later when you actually configure the provider.
* **The Source (`source = "hashicorp/azurerm"`):** This tells OpenTofu exactly where on the internet to find the official Azure plugin.
* **The Version (`version = "~> 3.0"`):** Just like the engine, we lock the plugin version to the `3.x` family to ensure stability.


* **The Docker Analogy:** Think of writing `azurerm = { source = "hash..." }` like pulling a Docker image. OpenTofu is like the Docker engine—it needs to know exactly which "image" (provider) to download before it can run anything. In this phase, we are simply declaring the image so OpenTofu can download it and learn how to speak Azure's language.

### The `provider` Block
```hcl
provider "azurerm" {
    features {}
}
```

#### The Phases
* **Phase 1:** The first thing we do after finishing the `terraform` block is actually start the created provider. We simply write `provider "azurerm"` (notice that the variable name coming after `provider` MUST match the variable name we used inside the `terraform` block to declare the required provider when we downloaded it). This whole phase is done simply to run the provider, nothing fancy.
* **Phase 2:** Then, inside this block (`provider`), we are forced to write `features {}`. Even if we don't mean anything by it, we are still forced to write it. Think of it as a "required syntax quirk". Ideally, inside this block (`provider`), we would write our credentials to log in to Azure, but it's not strictly necessary, fortunately, since we can do it via the CLI provided by Azure (which is `az`). Simply write `az login`, and it will redirect you to log in; then you'll have the credentials.

### The `resource_group` Block
```hcl
resource "azurerm_resource_group" "mimi_arg" {
  name     = "main"
  location = "austriaeast"
}
```

#### The Phases

* **Phase 1:** Since we now have a running provider, we can run Azure-specific commands. However, our first step is heavily locked down in 99% of cases, as Azure strictly requires your resources to stay in a group, so we create a group here. Notice that we use `resource`? `resource` is the command that tells Azure to create something instead of reading or editing. So we write `resource "azurerm_resource_group" "mimi_arg"` to tell Azure: "Hey Azure, create a resource group with the name 'main'." From now on, everything will be placed here (in this group).
* **Phase 2:** Since we just need to do some straightforward setup, I suggest checking the documentation and connecting to the closest location, rather than the one that sounds the prettiest.

### The `service_plan` Block
```hcl
resource "azurerm_service_plan" "mimi_asp" {
  name                = "kimishinu-backend"
  resource_group_name = azurerm_resource_group.mimi_arg.name
  location            = azurerm_resource_group.mimi_arg.location
  os_type             = "Linux"
  sku_name            = "F1"
}
```

#### The Phases
* **Phase 1:** Since we already have everything we need (the `terraform` block successfully defines what is expected from `opentofu`, and we have downloaded and are currently running a provider), we now declare "how much we are willing to spend." Think of a service plan like this: it is not an instance (VM) yet, it's just us telling Azure: "Hey, create a service plan that we will use to specify how many resources and which plan (tier) we are ready to allocate to the upcoming instances." Think of it as creating a document outlining what you can afford. This service must be declared as `azurerm_service_plan`—this is the core identifier. Next, we must provide it a `name` (which will be the name of the service plan we are about to create). After that, we must provide the `resource_group_name`, as Azure strictly expects everything to be tethered to a resource group. Then we need to give it a `location`, but we can use the same location we gave to our resource group. Finally, we specify what type of OS (operating system) we want our future instances to use via `os_type`.
* **Phase 2:** Lastly, we specify how much we are willing to spend and which tier we are going to use. In our case, we use `sku_name` to define our subscription level. We use `F1`, which is the free-tier subscription. From here, Azure will understand that it's for testing or small projects only, so it won't bother giving us highly powerful VMs, just basic ones.

## Reason Why Dropped
The reason is simple: I just don't like it. I have the Azure pack (via the GitHub Student Developer Pack), but the commands and everything else differ from what I used to know via AWS, so I am thinking of simply migrating to AWS (as initially planned).
