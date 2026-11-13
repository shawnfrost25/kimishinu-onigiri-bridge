## The Problem
Well, we already know how to create microservices and the like, but there is a problem: all of them are currently running on your localhost. None of them can be reached outside of it, and the moment you turn off your computer, they stop working, meaning they are no longer doing their jobs—which is something we clearly don't want. The main problem is simple: we don't have a way to keep the microservice running constantly, and it is fragilely connected to our local machine. This is what we fix via `opentofu`, which lets us write IaC (Infrastructure as Code) via Terraform using the HCL method. Thanks to it, we avoid logging into our provider (in my case, AWS) and clicking random buttons in the UI, which takes plenty of time. We simply write IaC and achieve the same results, just much faster. Let me explain the differences and details:
* **IaC + HCL:** IaC is a method; instead of using the UI, we use IaC (which simply means writing our infrastructure using code). Meanwhile, HCL (HashiCorp Configuration Language) is the syntax we use to write the code (through Terraform).
* **opentofu:** Think of this as the CLI tool for all of it. A simple mental model is to think of IaC and HCL as the `schema.sql`, while `opentofu` is `atlas`. See? They complement each other.

## Underlying Concept / How It Works
* I will explain how the `opentofu` commands and the concept work per se, what they do, and how they work.
* **"What is `terraform.tfstate`"** -> Think of it as a vault with all the information about the infrastructure. Since it contains all that information, it may also expose passwords and other sensitive values, which is why YOU MUST NEVER commit it to GitHub. If you do so, count the remaining days at the company you're in, because you will have to update your LinkedIn profile to "searching for a job." Just joking; this will not happen, since you're going to store it in a more secure place.
* **`tofu init`** -> We MUST run it at the start. It's similar to running `go mod init`, but different. We do this because `opentofu` has no idea how to interact with AWS, Azure, etc. When you create a folder named `infra`, or wherever you store your infrastructure, you need to run it. You must run it whenever you change settings or add a new module. In simple words: You need to run `tofu init` whenever you apply a change to anything within the `terraform {}` block, or when you add a `module`—run it only at these times. Why? Because its only job is to download plugins for `opentofu`; it won't touch anything on AWS or in your code. It only sets up `opentofu`.
* **`tofu plan`** -> Think of it as `git status`. It will read the `terraform.tfstate` in AWS and the current configurations you've written, compare them, and show you what will change. It's strongly recommended to always use it and review the changes, because infrastructure is not something to mess with.
* **`tofu apply`** -> This is the actual command that applies everything. It will update the `terraform.tfstate` file in AWS and apply everything declared in `resource` to AWS.

## Example Code and Explanation (OpenTofu settings)
### The `terraform` Block
The `terraform {}` block is the global configuration for OpenTofu itself. It does not build or touch any infrastructure in AWS. Instead, it acts as a setup list that tells OpenTofu which engine version to use and which external plugins it needs to download before doing any real work.
```hcl
terraform {
  required_version = "~> 1.6"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```
* **Phase 1:** Before doing anything, we declare the exact version of OpenTofu we expect to run. The `~>` symbol means "you can wander wherever you want within this family, but nowhere below or above." By writing `~> 1.6`, you are telling OpenTofu it is allowed to use `1.6.0` up to `1.9.99`, but it is strictly forbidden from jumping to `2.0` or dropping below `1.6`. This prevents accidental breaking changes.
* **Phase 2:** Because OpenTofu does not know how to talk to AWS out of the box, we use this block to list the external translation plugins it needs to download. Remember, this phase doesn't run anything—it just downloads the necessary tools. This is why we MUST run `tofu init` when changing something here, because we are changing how `opentofu` should interact with AWS.
* **Phase 3:** Inside the shopping list, we have to hard-lock ourselves to a specific provider:
* **The Name (`aws = {}`):** This is just a custom variable name you choose to store the provider. You could technically name it anything (you could write it as `sheena = {}`), but `aws` is the standard. You will use this variable name later when you actually configure the provider.
* **The Source (`source = "hashicorp/aws"`):** This tells OpenTofu exactly where on the internet to find the official AWS plugin.
* **The Version (`version = "~> 5.0"`):** Just like the engine, we lock the plugin version to the `5.x` family to ensure stability.
* **The Docker Analogy:** Think of writing `aws = { source = "hashicorp/..." }` like pulling a Docker image. OpenTofu is like the Docker engine—it needs to know exactly which "image" (provider) to download before it can run anything. In this phase, we are simply declaring the image so OpenTofu can download it and learn how to speak AWS's language.

### The `bucket` Block
As said, you won't store `terraform.tfstate` locally, which, honestly, is alright for local development, but you would rather store it inside a "bucket"—think of a bucket as a vault on AWS. By declaring a bucket, you must declare its name and the path where your `terraform.tfstate` is going to sit, as well as the region (I suggest declaring your region here based on where your company is located, or, if it's only you, simply declare the closest region). And so on, let me show the code and explain it.
```hcl
backend "s3" {
  bucket       = "kimishinu-poll-787391402772-eu-north-1-an"
  key          = "kimishinu/prod/terraform.tfstate"
  region       = "eu-north-1"
  encrypt      = true
  use_lockfile = true
}
```
* Let me simply explain how it works, divided into phases.
* **Phase 1** -> The first thing we do is declare `backend "s3"` (DO NOT mistake it for a simple Amazon S3 instance. Here it has a different meaning: it means "Simple Storage Service"—this is why it's called `"s3"`) since we are on AWS. This tells OpenTofu to store the `terraform.tfstate` inside a bucket, and then we specify the information about it.
* **Phase 2** -> As a second step, we specify the bucket name and its "folder" (key). When we write `bucket`, we are telling it to use our bucket's name, while when we write `key`, we are telling it where exactly to store the `terraform.tfstate` file.
* **Phase 3** -> Then we write `region` and specify which region we want our bucket to be stored in, because it's not like `opentofu` randomly starts looking around the world for it. It simply looks at the region and then searches for the bucket name, which MUST be unique.
  * Difference between the `backend "s3"` bucket and the region we will later declare when building VMs (instances):
    * **Inside `backend`** -> Think of it as declaring where the `opentofu` commands must go and run. Usually, you'd like to write the closest region to your local machine (or company).
    * **Inside `resource "aws_instance"`** -> When declaring the creation of a VM, we simply tell it where our VM will be located. It's preferable to choose the region with the most customers. For example, if you know that you created a new app in Italy, you would rather specify the region for Milan (which is `eu-south-1`) so the people using the app won't suffer from latency.
* **Phase 4** -> Let me give a quick explanation of what these two are:
  * **`encrypt`** -> This ensures that your `terraform.tfstate` is encrypted at rest when it is uploaded and saved inside the Amazon S3 bucket (inside our `backend "s3"`).
  * **`use_lockfile`** -> This prevents two users from changing the same file and deploying. For example, your whole team enters the infrastructure, starts changing it, and two people click to deploy at the same time. Without `use_lockfile`, the infrastructure could suffer corruption, while with it, it will simply tell the slower person, "Hang on, somebody is already doing stuff."
* **The catch** -> Since `tofu init` doesn't change anything, it will fail if we try to store the `terraform.tfstate` file inside `"kimishinu-poll"` when that bucket does not exist. There are two choices: we can create it via HCL—not really recommended, since we would end up in a "chicken-and-egg" paradox—or do it the sensible way: enter AWS and create it through the console. The steps are easy:
```plaintext
Click the console search panel -> write "buckets" -> click Buckets -> personalize it as you wish (copy its name if regional), and from now on, you have a bucket.
```

## Example Code and Explanation (General knowledge)
### Environmental variables
* Terraform itself needs a specific format for `.env` variables to run, since it uses its own way to declare variables, usually by creating a file as `<filename>.tfvars`. But we are going to use the best industry practice (in my opinion): for non-sensitive information, we use `<filename>.tfvars`, and AWS Secrets on deployment. I suggest separating variable declarations from the main logic. For example, put them inside a file called `variables.tf`, then, after declaring them, use a file called `<filename>.tfvars` to give values to the variables (non-sensitive ones). This way, we won't pollute the main structure. Here is what I mean.
* **inside `variables.tf`:**
```hcl
variable "region" {
  type        = string
  description = "The region where the instances are going to be located"
}

variable "author" {
  type        = string
  description = "The author of the AWS credentials"
}
```
* **inside `<filename>.tfvars`:**
```hcl
region = "us-east-1"
author = "<username>"
```
### IAM Policies
Well, there's a clear distinction from a normal environment, because AWS locks all the permissions of a virtual machine. When sending a binary to a VM, you're sending it to prison. Yup, and the IAM policies are the security guard allowing what it can do. At first, I used to confuse them a lot. I used to think that IAM policies were the ones granting permission for what our VM can and cannot do with connections, methods, and so on. But the main answer is simple: IAM policies handle the rights of how our VM can interact with AWS. For example, since all the VMs on AWS default to having no rights, they can't even interact with AWS, meaning they won't be able to fetch secrets and so on. This is why we must specify what they can and cannot do. For example, AWS is locked so that `opentofu` won't be able to interact with it unless we create an IAM user with the necessary rights. Let me give an example of what we need to do to grant our backend permission to read secrets from AWS (stored inside SSM). But here is something I wanted to say: **IT'S MATHEMATICALLY IMPOSSIBLE TO UNDERSTAND ALL OF IAM**—there are over 5,000 IAM-related things, and there are many more, so there is no possible way that you'll inspect your infrastructure and code and then say, "Hmm... I think we need the `ssm:GetParameter` action and also `ssm:GetParametersByPath`." Cool, though no fun. To find out which IAM permissions we actually need, we make the code tell us. This is my strategy:
* Instead of breaking everything, you need to create another bucket. One will be `kimishinu-poll-<variable for environment>`, and we create two buckets: one is `kimishinu-poll-production`, and the other is `kimishinu-poll-staging`. Then, we change all the names to include the suffix. You will also have two `<filename>.tfvars` files—one for production (which is going to create a good EC2 instance with good specs and everything), while the test `.tfvars` is going to create a weaker and less expensive one. The reason we do it is simple: first, we want to see how it acts on AWS without breaking the main environment. Also, this way, we can attach an empty IAM policy to it, causing `tofu apply` to give us an error with the missing IAM permissions. This clearly lets us add the required permissions instead of guessing and granting excessive privileges, which is a breach of AWS's main idea: least privilege = security. We don't want to give a VM the ability to delete files if all it needs to do is read them. So the main strategy is: Never push to production; always push to test first with the same IAM policies that production is slowly adopting, to see what breaks and what doesn't. We must give it a policy first, so we need to write at least this:
```hcl
resource "aws_iam_role" "backens_secret_fetch" {
  name = "test_policy"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sts:AssumeRole"
        ]
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}
```
* Quick explanation. When we write `aws_iam_role`, we need to declare multiple settings. We do it in a JSON style, but there is documentation for it, so it's not as complicated. We are currently defining which role our instance is going to get attached to. In this case, we wrote `sts:AssumeRole`. Let me give an analogy for what it means: Think of the credential token as a visitor security badge. Think of yourself as a senior engineer; the boss let you access the main infrastructure, but your simple employee badge won't do anything, and it won't give you the needed permissions. So you walk to the IAM Policy Document security guard, who will see your status (who you are and your role) and assume your role. Now you have the visitor security badge, meaning you can access the infrastructure. See? An "IAM Role" is simply us creating the badge and saying who can receive it. In our case, we clearly stated, "We can give it to AWS EC2 only." This is what we need to provide to our EC2 (no policy document or anything) if we want it to point out the missing IAM permissions and then we can attach them. But before that, here is what each IAM phase does. Since it is divided into multiple phases, here they are:
* **`aws_iam_role`** -> Before assigning everything, we must declare which IAM role we are going to create and to whom it is assignable. We will not simply give it to everyone, so we specify who can receive it and what type of permission it provides. In our case, we make it assignable only to AWS EC2 (`ec2.amazonaws.com`), and we assume their role, so we won't have to worry.
* **`aws_iam_policy`** -> We are happy: we just declared who can bear the role, but there's a difference. We need to specify what permissions the policy we are going to create has. For example, maybe our backend needs to read AWS secrets containing sensitive information such as API keys, so we declare a set of rules (for example: may the service that gets this policy be able to create files and also read environment variables?).
* **`aws_iam_role_policy_attachment`** -> Just as the name states, we simply attach the IAM role and the IAM policies together. Since we declared them separately, now they'll act as a set of rules. Without this command, they would remain a random policy and a random role; after this, they'll become a role with a set of policies.
* **`aws_iam_instance_profile`** -> This is the final part. Think of it as a wrapper that takes the roles with their set of rules and makes them available to be attached to the EC2.
* **FULL WORKFLOW**:
```plaintext
Create a Role (state who can attach it) -> Create the policy (give it the needed permissions) -> Attach the role and policy together -> Wrap it to make it compatible -> Simply state it inside the VM that needs those roles
```

// Continuation of Subnets and Network Groups tomorrow
