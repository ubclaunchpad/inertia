
# Download our release binary builder
go get -u github.com/mitchellh/gox

# Specify platforms and release version
PLATFORMS="linux/amd64 linux/386 darwin/386 windows/amd64 windows/386"
RELEASE=$(git describe --tags)
echo "Building release $RELEASE"

# Build, tag and push Inertia Docker image
make daemon RELEASE=$RELEASE

# Build Inertia Go binaries for specified platforms
gox -output="inertia.$(git describe --tags).{{.OS}}.{{.Arch}}" \
    -ldflags "-w -s -X main.Version=$RELEASE" \
    -osarch="$PLATFORMS" \
