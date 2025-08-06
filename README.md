Use any Terraform provider with Pulumi

For complete documentation and examples, see the [Pulumi Registry](https://www.pulumi.com/registry/packages/terraform-provider/).

---

## Structure

`pulumi-terraform-provider` is developed in the [pulumi/pulumi-terraform-bridge](https://github.com/pulumi/pulumi-terraform-bridge)
repository in the [`./dynamic`](https://github.com/pulumi/pulumi-terraform-bridge/tree/master/dynamic) folder.

This repository hosts user-facing documentation and hosts releases.

## Releasing

To release `pulumi-terraform-provider`, tag this repository locally and then push tags:

```sh
git tag v<next>
git push --tags
```

This will kick off the `release` Workflow, which will create a release from 
[`pulumi-terraform-bridge@master`](https://github.com/pulumi/pulumi-terraform-bridge/tree/master).

> [!NOTE]
> Do _not_ create a release via the GitHub UI.  The `release` Workflow relies on a release
> asset from the previous release to generate release notes.
