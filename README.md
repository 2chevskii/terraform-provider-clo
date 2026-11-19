# CLO Terraform provider

This fork manages compute, disks, networking, projects, limits, managed databases, licenses and S3 users through the CLO v2 HTTP API. The provider exposes 22 resources and 48 data sources. Configuration, examples and import formats are in the [resource reference](docs/resources) and [data source reference](docs/data-sources). The [coverage report](API-COVERAGE.md) compares the fork with upstream v2.8.0.

## Install this fork from source

The following PowerShell commands build and install a local development version on Windows x64. Run them from the cloned fork with Go 1.25 or newer and Terraform installed. The local provider address is `2chevskii/clo`.

```powershell
$version = "2.8.1-dev.0"
$mirror = Join-Path $env:LOCALAPPDATA "clo-terraform/mirror"
$package = Join-Path $mirror "registry.terraform.io/2chevskii/clo/$version/windows_amd64"
New-Item -ItemType Directory -Force $package | Out-Null
go build -o (Join-Path $package "terraform-provider-clo_v$version.exe") .

$env:TF_CLI_CONFIG_FILE = Join-Path $env:LOCALAPPDATA "clo-terraform/terraform.rc"
$mirrorPath = $mirror.Replace('\', '/')
@"
provider_installation {
  filesystem_mirror {
    path = "$mirrorPath"
    include = ["registry.terraform.io/2chevskii/clo"]
  }
  direct {
    exclude = ["registry.terraform.io/2chevskii/clo"]
  }
}
"@ | Set-Content -Encoding ascii $env:TF_CLI_CONFIG_FILE
```

The CLI configuration applies to the current shell. If you already use a Terraform CLI configuration, you can merge the installation rules into that file. On other platforms, build the same Go package and use the corresponding Terraform platform directory and executable name.

## Configure Terraform

Place this configuration in a Terraform working directory. Set `CLO_API_AUTH_TOKEN` in the environment and run `terraform init`. The sample reads the projects visible to the token.

```hcl
terraform {
  required_providers {
    clo = {
      source  = "2chevskii/clo"
      version = "2.8.1-dev.0"
    }
  }
}

provider "clo" {
  auth_url = "https://api.clo.ru"
}

data "clo_projects" "available" {}
```

## Test the provider

Run `go test ./...` and `go vet ./...` from the repository root. Local tests use HTTP fixtures. The acceptance suite creates and deletes CLO infrastructure and requires `TF_ACC=1`, `CLO_API_AUTH_URL`, `CLO_API_AUTH_TOKEN`, and a disposable `CLO_API_PROJECT_ID`.
