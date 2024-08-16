# Getting Started (Internal Beta)

Dynamically bridged Terraform providers are currently available for Pulumi Python.

To get started, run:

```console
GITHUB_TOKEN=$(gh auth token) pulumi plugin install resource terraform-provider
```

> [!NOTE]
> The `GITHUB_TOKEN=$(gh auth token)` is necessary because the `pulumi-terraform-provider`
> repo is currently private. This will not be the case in the public beta.

> [!TIP]
> `terraform-provider` is actually a Pulumi provider that "provides" Terraform providers.

## Testing

We're eager to learn more about how this method works for a wide variety of providers, so
please try it out with any TF provider you are familiar with and see if it gives you a
useful SDK. You can report problems in [#project-dynamic-bridged-providers](https://pulumi.slack.com/archives/C07A45FT70W) in slack or
by filing an issue in [pulumi-terraform-provider](https://github.com/pulumi/pulumi-terraform-provider/issues).

Any feedback is welcome!

## Walk-through Demo

Start in a new project:

```console
mkdir demo-dynamic-terraform-provider && cd demo-dynamic-terraform-provider
pulumi new python
```

For the sake of keeping the demo run-able without credentials, I will use Hashicorp's
Random provider, but any provider will work the same way.

### Prerequisites

We need to use *exactly* `v3.129.0` of `pulumi` and `pulumi-language-python`.

> [!WARNING]
> If you use the wrong version of `pulumi` or `pulumi-language-python` you may get an
> inscrutable error message.

```console
pulumi version
3.126.0
```

> [!TIP]
> Pulumi uses language servers to handle SDK generation. If you don't know what
> `pulumi-language-python` is, then it was installed alongside the `pulumi` binary at the
> same version.

Because `pulumi-terraform-provider` is a private repository, you will need a way to access
it from the CLI. I recommend installing [`gh`](https://cli.github.com/), which is what I will use for the
example. You may use whatever process you want so that `GITHUB_TOKEN=$(...)` sets a valid
GitHub token.

### Generate and install the SDK

We first create the SDK for the provider:

```console
pulumi package add terraform-provider hashicorp/random
Successfully generated a Python SDK for the random package at ./demo-dynamic-terraform-provider/sdks/random

To use this SDK in your Python project, run the following command:

  echo sdks/random >> requirements.txt

  pulumi install

You can then import the SDK in your Python code with:

  import pulumi_random as random

```

This will generate an SDK for `hashicorp/random` at the latest version in `./sdks/random`. The
SDK is named `pulumi_random`.

We now follow the instructions that `pulumi package add` gave us:

```console
echo sdks/random >> requirements.txt
```

At this point, our `requirements.txt` file reads as:

```pip
pulumi==3.129.0 # The interface is currently unstable, so we need to manually match versions.
./random/python # Our local terraform-provider-random SDK
```

We can then "finish" the installation:

```console
pulumi install
```

We have now installed the generated SDK into the local Python environment. We can consume
it as normal.

> [!NOTE]
> The UX for generating and installing the SDK is still a work in progress. We expect the
> final version of this instruction to read "Run `pulumi package use terraform-provider
> hashicorp/random` and follow the instructions it prints out."

### Creating a random resource

At this point, you can treat `pulumi_random` just like any other Pulumi SDK.

> [!TIP]
> `pulumi_random` *is* a full Pulumi SDK. It contains all information it needs to download
> `terraform-provider-random`. If you check in `random`, you won't even need to run
> `pulumi package gen-sdk ...` again.

Edit your `__main__.py` file to consume a resource from the newly generated SDK:

```python
"""A Python Pulumi program"""

import pulumi
import pulumi_random as random

pet = random.Pet("hello", length=3, prefix="dynamic")

pulumi.export("result", pet.id)
```

You can now run `pulumi up`:

```console
pulumi up
Previewing update (dev)

View in Browser (Ctrl+O): https://app.pulumi.com/pulumi/demo-dynamic-terraform-provider/dev/previews/1b034d96-3e55-4fb2-909a-222d72555b21

     Type                 Name                                 Plan
 +   pulumi:pulumi:Stack  demo-dynamic-terraform-provider-dev  create
 +   └─ random:index:Pet  hi                                   create

Outputs:
    result: output<string>

Resources:
    + 2 to create

Do you want to perform this update? yes
Updating (dev)

View in Browser (Ctrl+O): https://app.pulumi.com/pulumi/demo-dynamic-terraform-provider/dev/updates/1

     Type                 Name                                 Status
 +   pulumi:pulumi:Stack  demo-dynamic-terraform-provider-dev  created (0.89s)
 +   └─ random:index:Pet  hi                                   created (0.12s)

Outputs:
    result: "dynamic-routinely-related-foal"

Resources:
    + 2 created

Duration: 2s
```
