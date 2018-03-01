
PLATFORMS="linux/amd64 linux/386 darwin/amd64 darwin/386 windows/amd64 windows/386"
RELEASE=$(git describe --tags)
echo "Building release $RELEASE"

# Build, tag and push Inertia Docker image
make docker RELEASE=$RELEASE

# Build Inertia Go binaries for specified platforms
gox -output="inertia_$(git describe --tags).{{.OS}}.{{.Arch}}" \
    -ldflags "-X main.Version=$RELEASE" \
    -osarch="$PLATFORMS" \
