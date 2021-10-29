# pacman

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
