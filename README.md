# CRDiff - Compare Kubernetes CRDs & Detect Breaking Changes

CRDiff is a small utility to compare 2 versions of a set of CRDs, reporting all the changes and listing breaking changes between the two versions. CRDiff can compare many CRDs at once and has helpers to make usage in CI/CD systems especially convenient.

CRDiff is using the OpenAPI differ [oasdiff](https://github.com/tufin/oasdiff) to compare 2 OpenAPI schemas. For this reason it's possible that not all Kubernetes-exclusive features (like `x-kubernetes-...` annotations) are supported yet.

## Features

* Compare Kubernetes CRDs (apiextensions v1beta1 and v1).
* Reports all differences and/or just breaking changes.
* Can compare either single CRDs or entire directories recursively.
* Output as either pretty text or nerdy JSON, depending on your needs.

## Installation

Either [download the latest release](https://github.com/xrstf/crdiff/releases) or build for yourself using Go 1.20+:

```bash
go install go.xrstf.de/crdiff
```

## Usage

```bash
Compare Kubernetes CRDs

Usage:
  crdiff [command]

Available Commands:
  breaking    Compare two or more CRD files/directories and print all breaking differences
  diff        Compare two or more CRD files/directories and print the differences
  help        Help about any command
  version     Print the application version and then exit

Flags:
  -h, --help      help for crdiff
  -v, --verbose   enable verbose logging

Use "crdiff [command] --help" for more information about a command.
```

### Simples Case: Compare 2 versions of the same CRD

Just use the `diff` sub command and specify 2 YAML files (first one is base, or the old one, and the second is revision, i.e. the one with changes).

```bash
crdiff diff mycrd-old.yaml mycrd-new.yaml
```

Each of these YAML files can contain unrelated resources. CRDiff will filter out anything that is not looking like a Kubernetes CRD. So you can for example use CRDiff upstream YAML files that contain _everything_ (Deployments, Secrets and CRDs, among others).

### Compare 2 versions of the same set of CRDs

Instead of just giving a single file, you can also specify directories for either base or revision CRDs. CRDiff will read the directory recursively, find all `*.yaml` and `.yml` files and extract all CRDs from those. Then it will compile all found CRDs, reporting on those that were added and removed as well.

```bash
crdiff diff old-crds/ new-crds/
```

### See Breaking Changes

To only see breaking changes, use the `breaking` instead of `diff` subcommand:

```bash
crdiff breaking old-crds/ new-crds/
```

Note that `breaking` will exit with a non-zero status code when there are breaking changes. `diff` will always exit with 0 (unless an error occurs).

## License

MIT
