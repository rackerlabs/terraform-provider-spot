# Steps for Publishing a Release

1. Comment `signs` section from `.goreleaser.yml`
2. Choose the release version, latest version +1, for example, `0.0.3`
3. Copy the passphrase for signing the release assets.
3. Run script `manual-release/release.sh 0.0.3`. It will build release assets under `./dist` directory.
4. Verify a release assets [here](https://github.com/rackerlabs/terraform-provider-spot/releases).
5. Publish the release.
