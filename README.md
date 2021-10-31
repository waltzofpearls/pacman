# pacman

[![Build Status][actions-badge]][actions-url]
[![MIT licensed][mit-badge]][mit-url]

[actions-badge]: https://github.com/waltzofpearls/pacman/workflows/ci/badge.svg
[actions-url]: https://github.com/waltzofpearls/pacman/actions?query=workflow%3Aci+branch%3Amain
[mit-badge]: https://img.shields.io/badge/license-MIT-green.svg
[mit-url]: https://github.com/waltzofpearls/pacman/blob/main/LICENSE

## Getting started

First, install certstrap so we can generate certs for mTLS. On MacOS use `brew install certstrap`, or
follow [this instruction](https://github.com/square/certstrap#building) to build from source.

Next, let's run it and seed the package registry. Initial run will prompt for entering passphrase from
`certstrap`. Just leave all the passphreases empty and proceed.

```shell
make run
make seed
```

After run and seed, `make list` will list packages from seeded registry. Add or remove package with:

```shell
make add name='package_name' deps='dep1 dep2'
make remove name='package_name'
```
