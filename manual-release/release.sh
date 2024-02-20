#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <version> without v"
    exit 1
fi

VERSION=$1

export NAME=spot

git checkout main

TAG_NAME="v${VERSION}"
git tag ${TAG_NAME}
git push origin ${TAG_NAME}

goreleaser --snapshot --clean

find dist/ -type d -name 'terraform-provider-*' -exec rm -r {} +

SHORT_COMMIT_ID=$(git log -1 --format=%h | cat)

cp terraform-registry-manifest.json dist/terraform-provider-${NAME}_${VERSION}_manifest.json

# The release directory 'dist/' contains archives with -SNAPSHOT-<commit-id> in their name
# using following code to get rid of the -SNAPSHOT-<commit-id> suffix from binaries inside archives
# and archives itself
pushd ./dist
for zip_file in ./*.zip; do
    mkdir workspace
    unzip "$zip_file" -d workspace
    pushd workspace
    bin_filename=$(find . -type f -name "terraform-provider-*")
    new_bin_filename=$(echo $bin_filename | sed "s/-SNAPSHOT-${SHORT_COMMIT_ID}//")
    mv $bin_filename $new_bin_filename
    new_zip_file=$(echo $zip_file | sed "s/-SNAPSHOT-${SHORT_COMMIT_ID}//")
    zip -r $new_zip_file *
    mv $new_zip_file ../
    popd
    rm -rf workspace/
    rm -f $zip_file
done

rm -f ./terraform-provider-${NAME}_${VERSION}-SNAPSHOT-${SHORT_COMMIT_ID}_SHA256SUMS
rm -f ./artifacts.json ./metadata.json ./config.yaml

shasum -a 256 *.zip > terraform-provider-${NAME}_${VERSION}_SHA256SUMS
shasum -a 256 terraform-provider-${NAME}_${VERSION}_manifest.json >> terraform-provider-${NAME}_${VERSION}_SHA256SUMS
gpg --detach-sign terraform-provider-${NAME}_${VERSION}_SHA256SUMS

# gpg --verify terraform-provider-${NAME}_${VERSION}_SHA256SUMS.sig terraform-provider-${NAME}_${VERSION}_SHA256SUMS

gh release create --title ${TAG_NAME} --generate-notes --draft --verify-tag ${TAG_NAME}

for asset in ./*; do
    gh release upload ${TAG_NAME} $asset
done

popd

# Publish the release
# DO THIS MANUALLY AFTER CHECKING THE DRAFT
# gh release edit ${TAG_NAME} --draft false

