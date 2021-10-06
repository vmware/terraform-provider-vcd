---
name: Bug Report
about: If something isn't working as expected.

---

Hello,

Thank you for opening an issue. Please note that we try to keep the Terraform issue tracker reserved
for bug reports and feature requests. For general usage questions, please see:
https://www.terraform.io/community.html.

### Terraform Version
Run `terraform -v` to show the version. If you are not running the latest version of Terraform, please upgrade because your issue may have already been fixed.

### Affected Resource(s)
Please list the resources as a list, for example:
- vcd_vapp_vm
- vcd_network_routed_v2

If this issue appears to affect multiple resources, it may be an issue with Terraform's core, so please mention this.

### Terraform Configuration Files

```hcl
Copy-paste your Terraform configurations here. Please **shrink HCL replicating the bug to minimum** 
without modules. It will help us to respond and replicate the problem quicker. It may also uncover
some HCL error while doing so.

For bigger configs you may use [Gist(s)](https://gist.github.com) or a service like Dropbox and
share a link to the ZIP file.
```

### Debug Output
Please provide a link to GitHub [Gist(s)](https://gist.github.com) containing complete debug
output. You can enable debug by using the commands below:
```shell
export TF_LOG_PATH=tf.log            
export TF_LOG=TRACE                  
export GOVCD_LOG_FILE=go-vcloud-director.log
export GOVCD_LOG=true     
```

On Windows the command instead of `export` is `set`.

Replicate the issue after setting the environment variables listed above and it should create two
new files in the working directory: `tf.log` and `go-vcloud-director.log`. The `tf.log` is a general
Terraform debug log (more information about it is in
https://www.terraform.io/docs/internals/debugging.html) while the `go-vcloud-director.log` is a
specific log file for `terraform-provider-vcd` containing debug information about performed API
calls. Please attach both of them to your [Gist](https://gist.github.com).

### Panic Output
If Terraform produced a panic, please provide a link to a GitHub Gist containing the output of the `crash.log`.

### Expected Behavior
What should have happened?

### Actual Behavior
What actually happened?

### Steps to Reproduce
Please list the steps required to reproduce the issue, for example:
1. `terraform apply`

### User Access rights 
Information about user used. Role and/or more exact rights if it is customized.

### Important Factoids
Is there anything atypical about your accounts that we should know?

### References
Are there any other GitHub issues (open or closed) or Pull Requests that should be linked here? For example:
- Issue #0000
