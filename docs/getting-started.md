# Getting Started (Internal Beta)

Dynamically bridged Terraform providers are currently available for Pulumi Python.

To get started, run:

```console
$ GITHUB_TOKEN=$(gh auth token) pulumi plugin install resource terraform-provider
```

> [!NOTE]
> The `GITHUB_TOKEN=$(gh auth token)` is necessary because the `pulumi-terraform-provider`
> repo is currently private. This will not be the case in the public beta.

## Walk-through Demo

Start in a new project:

```console
$ mkdir demo-dynamic-terraform-provider && cd demo-dynamic-terraform-provider
$ pulumi new python
```

For the sake of keeping the demo run-able without credentials, I will use Hashicorp's
Random provider, but any provider will work the same way.

### Generate and install the SDK

We need to use `v1.126.0` of `pulumi` and `pulumi-language-python`.

```console
$ pulumi version
3.126.0
```

> [!TIP]
> Pulumi uses language servers to handle SDK generation. If you don't know what
> `pulumi-language-python` is, then it was installed alongside the `pulumi` binary at the
> same version.

We first create SDK for the provider:

```console
$ pulumi package gen-sdk terraform-provider --language python --out random -- hashicorp/random
```

To tell `pip` how to depend on the locally generated SDK, we need to edit `requirements.txt`:

```console
$ emacs requirements.txt
pulumi==3.126.0     # The interface is currently unstable, so we need to manually match versions.
-e ./random/python  # -e tells pip that this is a local provider
$ ./venv/bin/pip install -r requirements.txt
```

We have now installed the generated SDK into the local Python environment. We can consume
it as normal:

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
$ pulumi up
Previewing update (dev)

View in Browser (Ctrl+O): https://app.pulumi.com/pulumi/demo-dynamic-terraform-provider/dev/previews/1b034d96-3e55-4fb2-909a-222d72555b21

Loading policy packs...

     Type                 Name                                 Plan
 +   pulumi:pulumi:Stack  demo-dynamic-terraform-provider-dev  create
 +   └─ random:index:Pet  hi                                   create

Policies:
    ✅ pulumi-internal-policies@v0.0.6

Outputs:
    result: output<string>

Resources:
    + 2 to create

Do you want to perform this update? yes
Updating (dev)

View in Browser (Ctrl+O): https://app.pulumi.com/pulumi/demo-dynamic-terraform-provider/dev/updates/1

Loading policy packs...

     Type                 Name                                 Status
 +   pulumi:pulumi:Stack  demo-dynamic-terraform-provider-dev  created (0.89s)
 +   └─ random:index:Pet  hi                                   created (0.12s)

Policies:
    ✅ pulumi-internal-policies@v0.0.6

Outputs:
    result: "dynamic-routinely-related-foal"

Resources:
    + 2 created

Duration: 2s
```
