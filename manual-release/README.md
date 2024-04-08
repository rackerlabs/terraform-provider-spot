Setup for Release
==================

1. Comment `signs` section from `.goreleaser.yml`
2. Choose the release version, latest version +1, for example, `0.0.3`
3. Copy the passphrase for signing the release assets.
3. Run script `manual-release/release.sh 0.0.3`. It will build release assets under `./dist` directory.
4. Verify a release assets [here](https://github.com/rackerlabs/terraform-provider-spot/releases).
5. Publish the release.

# Manual steps

To debug any release step, refer the following set of commands and their output.

```console
$ export NAME=spot
$ export VERSION=0.0.2

$ git checkout main
$ git tag v${VERSION}
$ git push origin v${VERSION}

$ goreleaser --snapshot --clean

$ cp terraform-registry-manifest.json dist/terraform-provider-${NAME}_${VERSION}_manifest.json

# The release directory 'dist/' contains archives with -SNAPSHOT-<commit-id> in their name
# This suffix is also added to the binaries inside archives
# We need to manually open each archive and remove -SNAPSHOT-<commit-id> from binary name

$ shasum -a 256 *.zip > terraform-provider-${NAME}_${VERSION}_SHA256SUMS
$ shasum -a 256 terraform-provider-${NAME}_${VERSION}_manifest.json >> terraform-provider-${NAME}_${VERSION}_SHA256SUMS
$ gpg --detach-sign terraform-provider-${NAME}_${VERSION}_SHA256SUMS
$ gpg --verify terraform-provider-${NAME}_${VERSION}_SHA256SUMS.sig terraform-provider-${NAME}_${VERSION}_SHA256SUMS
gpg: Signature made Fri 02 Feb 2024 02:19:10 AM IST
gpg:                using RSA key 4168A32E08C104352C677A3BBE807C25E49FC5D6
gpg: Good signature from "Platform9 <terraform@platform9.com>" [ultimate]
```
