#!/bin/bash


VERSION_FILE="./version"


CURRENT_VERSION=$(cat "$VERSION_FILE")

parse_version() {
    if [[ "$1" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)([a-z]([0-9]+))?$ ]]; then
        MAJOR=${BASH_REMATCH[1]}
        MINOR=${BASH_REMATCH[2]}
        PATCH=${BASH_REMATCH[3]}
        PRE_RELEASE=${BASH_REMATCH[4]}
        PRE_NUM=${BASH_REMATCH[5]}
    else
        echo "Invalid version format: $1"
        exit 1
    fi
}

parse_version "$CURRENT_VERSION"

case $1 in
    manual)
        if [ -z "$2" ]; then
            echo "Manual version required. Usage: $0 manual <version>"
            exit 1
        fi
        NEW_VERSION=$2
        ;;

    major)
        NEW_VERSION="$((MAJOR + 1)).0.0"
        ;;

    minor)
        NEW_VERSION="$MAJOR.$((MINOR + 1)).0"
        ;;

    prerelease)
        if [ -n "$PRE_RELEASE" ]; then
            NEW_VERSION="$MAJOR.$MINOR.$PATCH${PRE_RELEASE:0:1}$((PRE_NUM + 1))"
        else
            NEW_VERSION="$MAJOR.$MINOR.$((PATCH + 1))a0"
        fi
        ;;

    patch)
        if [ -n "$PRE_RELEASE" ]; then
            NEW_VERSION="$MAJOR.$MINOR.$PATCH"
        else
            NEW_VERSION="$MAJOR.$MINOR.$((PATCH + 1))"
        fi
        ;;

    *)
        echo "Invalid version type: $1"
        echo "Usage: $0 {major|minor|patch|prerelease|manual <version>}"
        exit 1
        ;;
esac

echo "$NEW_VERSION" > "$VERSION_FILE"

if [ -z "$(git config --get branch.$(git branch --show-current).remote)" ]; then
    git push --set-upstream origin $(git branch --show-current)
fi

git add "$VERSION_FILE"
git commit -m "$NEW_VERSION"
git push

if git rev-parse "v$NEW_VERSION" >/dev/null 2>&1; then
    echo "Tag v$NEW_VERSION already exists. Skipping release creation."
else
    PRERELEASE_FLAG=""
    if [[ $1 == "prerelease" || ($1 == "manual" && "$NEW_VERSION" =~ [a-z][0-9]+$) ]]; then
        PRERELEASE_FLAG="--prerelease"
    fi

    gh release create "v$NEW_VERSION" \
        --generate-notes \
        --target=$(git branch --show-current) \
        $PRERELEASE_FLAG
fi

echo "$NEW_VERSION"