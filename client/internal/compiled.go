// Code generated by fileb0x at "2020-03-22 14:46:44.689828 -0700 PDT m=+0.003781846" from config file "b0x.yml" DO NOT EDIT.
// modification hash(4ddd93ec81abcb42b40784c70d0b80ca.902253dd4a7873ead7b8164875470c16)

package internal

import (
	"bytes"

	"context"
	"io"
	"net/http"
	"os"
	"path"

	"golang.org/x/net/webdav"
)

var (
	// CTX is a context for webdav vfs
	CTX = context.Background()

	// FS is a virtual memory file system
	FS = webdav.NewMemFS()

	// Handler is used to server files through a http handler
	Handler *webdav.Handler

	// HTTP is the http file system
	HTTP http.FileSystem = new(HTTPFS)
)

// HTTPFS implements http.FileSystem
type HTTPFS struct {
	// Prefix allows to limit the path of all requests. F.e. a prefix "css" would allow only calls to /css/*
	Prefix string
}

// FileClientScriptsDaemonDownSh is "client/scripts/daemon-down.sh"
var FileClientScriptsDaemonDownSh = []byte("\x23\x21\x2f\x62\x69\x6e\x2f\x73\x68\x0a\x0a\x23\x20\x42\x61\x73\x69\x63\x20\x73\x63\x72\x69\x70\x74\x20\x66\x6f\x72\x20\x62\x72\x69\x6e\x67\x69\x6e\x67\x20\x64\x6f\x77\x6e\x20\x74\x68\x65\x20\x64\x61\x65\x6d\x6f\x6e\x2e\x0a\x0a\x73\x65\x74\x20\x2d\x65\x0a\x0a\x44\x41\x45\x4d\x4f\x4e\x5f\x4e\x41\x4d\x45\x3d\x69\x6e\x65\x72\x74\x69\x61\x2d\x64\x61\x65\x6d\x6f\x6e\x0a\x0a\x23\x20\x47\x65\x74\x20\x64\x61\x65\x6d\x6f\x6e\x20\x63\x6f\x6e\x74\x61\x69\x6e\x65\x72\x20\x61\x6e\x64\x20\x74\x61\x6b\x65\x20\x69\x74\x20\x64\x6f\x77\x6e\x20\x69\x66\x20\x69\x74\x20\x69\x73\x20\x72\x75\x6e\x6e\x69\x6e\x67\x2e\x0a\x41\x4c\x52\x45\x41\x44\x59\x5f\x52\x55\x4e\x4e\x49\x4e\x47\x3d\x60\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x70\x73\x20\x2d\x71\x20\x2d\x2d\x66\x69\x6c\x74\x65\x72\x20\x22\x6e\x61\x6d\x65\x3d\x24\x44\x41\x45\x4d\x4f\x4e\x5f\x4e\x41\x4d\x45\x22\x60\x0a\x69\x66\x20\x5b\x20\x21\x20\x2d\x7a\x20\x22\x24\x41\x4c\x52\x45\x41\x44\x59\x5f\x52\x55\x4e\x4e\x49\x4e\x47\x22\x20\x5d\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x72\x6d\x20\x2d\x66\x20\x24\x41\x4c\x52\x45\x41\x44\x59\x5f\x52\x55\x4e\x4e\x49\x4e\x47\x0a\x66\x69\x3b\x0a")

// FileClientScriptsDaemonUpSh is "client/scripts/daemon-up.sh"
var FileClientScriptsDaemonUpSh = []byte("\x23\x21\x2f\x62\x69\x6e\x2f\x73\x68\x0a\x0a\x23\x20\x42\x61\x73\x69\x63\x20\x73\x63\x72\x69\x70\x74\x20\x66\x6f\x72\x20\x73\x65\x74\x74\x69\x6e\x67\x20\x75\x70\x20\x49\x6e\x65\x72\x74\x69\x61\x20\x72\x65\x71\x75\x69\x72\x65\x6d\x65\x6e\x74\x73\x20\x28\x64\x69\x72\x65\x63\x74\x6f\x72\x69\x65\x73\x2c\x20\x65\x74\x63\x29\x0a\x23\x20\x61\x6e\x64\x20\x62\x72\x69\x6e\x69\x6e\x67\x20\x74\x68\x65\x20\x64\x61\x65\x6d\x6f\x6e\x20\x6f\x6e\x6c\x69\x6e\x65\x2e\x0a\x0a\x73\x65\x74\x20\x2d\x65\x0a\x0a\x23\x20\x55\x73\x65\x72\x20\x61\x72\x67\x75\x6d\x65\x6e\x74\x73\x2e\x0a\x44\x41\x45\x4d\x4f\x4e\x5f\x52\x45\x4c\x45\x41\x53\x45\x3d\x22\x25\x5b\x31\x5d\x73\x22\x0a\x44\x41\x45\x4d\x4f\x4e\x5f\x50\x4f\x52\x54\x3d\x22\x25\x5b\x32\x5d\x73\x22\x0a\x48\x4f\x53\x54\x5f\x41\x44\x44\x52\x45\x53\x53\x3d\x22\x25\x5b\x33\x5d\x73\x22\x0a\x57\x45\x42\x48\x4f\x4f\x4b\x5f\x53\x45\x43\x52\x45\x54\x3d\x22\x25\x5b\x34\x5d\x73\x22\x0a\x0a\x23\x20\x49\x6e\x65\x72\x74\x69\x61\x20\x69\x6d\x61\x67\x65\x20\x64\x65\x74\x61\x69\x6c\x73\x2e\x0a\x44\x41\x45\x4d\x4f\x4e\x5f\x4e\x41\x4d\x45\x3d\x69\x6e\x65\x72\x74\x69\x61\x2d\x64\x61\x65\x6d\x6f\x6e\x0a\x49\x4d\x41\x47\x45\x3d\x75\x62\x63\x6c\x61\x75\x6e\x63\x68\x70\x61\x64\x2f\x69\x6e\x65\x72\x74\x69\x61\x3a\x24\x44\x41\x45\x4d\x4f\x4e\x5f\x52\x45\x4c\x45\x41\x53\x45\x0a\x0a\x23\x20\x49\x74\x20\x64\x6f\x65\x73\x6e\x27\x74\x20\x6d\x61\x74\x74\x65\x72\x20\x77\x68\x61\x74\x20\x70\x6f\x72\x74\x20\x74\x68\x65\x20\x64\x61\x65\x6d\x6f\x6e\x20\x72\x75\x6e\x73\x20\x6f\x6e\x20\x69\x6e\x20\x74\x68\x65\x20\x63\x6f\x6e\x74\x61\x69\x6e\x65\x72\x0a\x23\x20\x61\x73\x20\x6c\x6f\x6e\x67\x20\x61\x73\x20\x69\x74\x20\x69\x73\x20\x6d\x61\x70\x70\x65\x64\x20\x74\x6f\x20\x74\x68\x65\x20\x63\x6f\x72\x72\x65\x63\x74\x20\x44\x41\x45\x4d\x4f\x4e\x5f\x50\x4f\x52\x54\x2e\x0a\x43\x4f\x4e\x54\x41\x49\x4e\x45\x52\x5f\x50\x4f\x52\x54\x3d\x34\x33\x30\x33\x0a\x0a\x23\x20\x55\x73\x65\x72\x20\x70\x72\x6f\x6a\x65\x63\x74\x0a\x6d\x6b\x64\x69\x72\x20\x2d\x70\x20\x22\x24\x48\x4f\x4d\x45\x22\x2f\x69\x6e\x65\x72\x74\x69\x61\x2f\x70\x72\x6f\x6a\x65\x63\x74\x0a\x0a\x23\x20\x49\x6e\x65\x72\x74\x69\x61\x20\x64\x61\x74\x61\x0a\x6d\x6b\x64\x69\x72\x20\x2d\x70\x20\x22\x24\x48\x4f\x4d\x45\x22\x2f\x69\x6e\x65\x72\x74\x69\x61\x2f\x64\x61\x74\x61\x0a\x0a\x23\x20\x43\x6f\x6e\x66\x69\x67\x75\x72\x61\x74\x69\x6f\x6e\x0a\x6d\x6b\x64\x69\x72\x20\x2d\x70\x20\x22\x24\x48\x4f\x4d\x45\x22\x2f\x69\x6e\x65\x72\x74\x69\x61\x2f\x63\x6f\x6e\x66\x69\x67\x0a\x0a\x23\x20\x50\x65\x72\x73\x69\x73\x74\x65\x6e\x74\x20\x64\x61\x74\x61\x0a\x6d\x6b\x64\x69\x72\x20\x2d\x70\x20\x22\x24\x48\x4f\x4d\x45\x22\x2f\x69\x6e\x65\x72\x74\x69\x61\x2f\x70\x65\x72\x73\x69\x73\x74\x0a\x0a\x23\x20\x49\x6e\x65\x72\x74\x69\x61\x20\x73\x65\x63\x72\x65\x74\x73\x0a\x6d\x6b\x64\x69\x72\x20\x2d\x70\x20\x22\x24\x48\x4f\x4d\x45\x22\x2f\x2e\x69\x6e\x65\x72\x74\x69\x61\x0a\x6d\x6b\x64\x69\x72\x20\x2d\x70\x20\x22\x24\x48\x4f\x4d\x45\x22\x2f\x2e\x69\x6e\x65\x72\x74\x69\x61\x2f\x73\x73\x6c\x0a\x0a\x23\x20\x43\x68\x65\x63\x6b\x20\x69\x66\x20\x61\x6c\x72\x65\x61\x64\x79\x20\x72\x75\x6e\x6e\x69\x6e\x67\x20\x61\x6e\x64\x20\x74\x61\x6b\x65\x20\x64\x6f\x77\x6e\x20\x65\x78\x69\x73\x74\x69\x6e\x67\x20\x64\x61\x65\x6d\x6f\x6e\x2e\x0a\x41\x4c\x52\x45\x41\x44\x59\x5f\x52\x55\x4e\x4e\x49\x4e\x47\x3d\x24\x28\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x70\x73\x20\x2d\x71\x20\x2d\x2d\x66\x69\x6c\x74\x65\x72\x20\x22\x6e\x61\x6d\x65\x3d\x24\x44\x41\x45\x4d\x4f\x4e\x5f\x4e\x41\x4d\x45\x22\x29\x0a\x69\x66\x20\x5b\x20\x21\x20\x2d\x7a\x20\x22\x24\x41\x4c\x52\x45\x41\x44\x59\x5f\x52\x55\x4e\x4e\x49\x4e\x47\x22\x20\x5d\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x50\x75\x74\x74\x69\x6e\x67\x20\x65\x78\x69\x73\x74\x69\x6e\x67\x20\x49\x6e\x65\x72\x74\x69\x61\x20\x64\x61\x65\x6d\x6f\x6e\x20\x74\x6f\x20\x73\x6c\x65\x65\x70\x22\x0a\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x72\x6d\x20\x2d\x66\x20\x22\x24\x41\x4c\x52\x45\x41\x44\x59\x5f\x52\x55\x4e\x4e\x49\x4e\x47\x22\x20\x3e\x20\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x0a\x66\x69\x3b\x0a\x0a\x69\x66\x20\x5b\x20\x22\x24\x44\x41\x45\x4d\x4f\x4e\x5f\x52\x45\x4c\x45\x41\x53\x45\x22\x20\x21\x3d\x20\x22\x74\x65\x73\x74\x22\x20\x5d\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x23\x20\x44\x6f\x77\x6e\x6c\x6f\x61\x64\x20\x72\x65\x71\x75\x65\x73\x74\x65\x64\x20\x64\x61\x65\x6d\x6f\x6e\x20\x69\x6d\x61\x67\x65\x2e\x0a\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x44\x6f\x77\x6e\x6c\x6f\x61\x64\x69\x6e\x67\x20\x24\x49\x4d\x41\x47\x45\x22\x0a\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x70\x75\x6c\x6c\x20\x22\x24\x49\x4d\x41\x47\x45\x22\x20\x3e\x20\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x0a\x65\x6c\x73\x65\x0a\x20\x20\x20\x20\x23\x20\x4c\x6f\x61\x64\x20\x74\x65\x73\x74\x20\x62\x75\x69\x6c\x64\x20\x74\x68\x61\x74\x20\x73\x68\x6f\x75\x6c\x64\x20\x68\x61\x76\x65\x20\x62\x65\x65\x6e\x20\x73\x63\x70\x27\x64\x20\x69\x6e\x74\x6f\x0a\x20\x20\x20\x20\x23\x20\x74\x68\x65\x20\x56\x50\x53\x20\x61\x74\x20\x2f\x64\x61\x65\x6d\x6f\x6e\x2d\x69\x6d\x61\x67\x65\x2e\x0a\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x4c\x6f\x61\x64\x69\x6e\x67\x20\x24\x49\x4d\x41\x47\x45\x22\x0a\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x6c\x6f\x61\x64\x20\x2d\x69\x20\x2f\x64\x61\x65\x6d\x6f\x6e\x2d\x69\x6d\x61\x67\x65\x20\x3e\x20\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x0a\x66\x69\x0a\x0a\x23\x20\x52\x75\x6e\x20\x63\x6f\x6e\x74\x61\x69\x6e\x65\x72\x20\x77\x69\x74\x68\x20\x61\x63\x63\x65\x73\x73\x20\x74\x6f\x20\x74\x68\x65\x20\x68\x6f\x73\x74\x20\x64\x6f\x63\x6b\x65\x72\x20\x73\x6f\x63\x6b\x65\x74\x20\x61\x6e\x64\x20\x0a\x23\x20\x72\x65\x6c\x65\x76\x61\x6e\x74\x20\x68\x6f\x73\x74\x20\x64\x69\x72\x65\x63\x74\x6f\x72\x69\x65\x73\x20\x74\x6f\x20\x61\x6c\x6c\x6f\x77\x20\x66\x6f\x72\x20\x63\x6f\x6e\x74\x61\x69\x6e\x65\x72\x20\x63\x6f\x6e\x74\x72\x6f\x6c\x2e\x0a\x23\x20\x53\x65\x65\x20\x74\x68\x65\x20\x52\x45\x41\x44\x4d\x45\x20\x66\x6f\x72\x20\x6d\x6f\x72\x65\x20\x64\x65\x74\x61\x69\x6c\x73\x20\x6f\x6e\x20\x68\x6f\x77\x20\x74\x68\x69\x73\x20\x77\x6f\x72\x6b\x73\x3a\x0a\x23\x20\x68\x74\x74\x70\x73\x3a\x2f\x2f\x67\x69\x74\x68\x75\x62\x2e\x63\x6f\x6d\x2f\x75\x62\x63\x6c\x61\x75\x6e\x63\x68\x70\x61\x64\x2f\x69\x6e\x65\x72\x74\x69\x61\x23\x68\x6f\x77\x2d\x69\x74\x2d\x77\x6f\x72\x6b\x73\x0a\x65\x63\x68\x6f\x20\x22\x52\x75\x6e\x6e\x69\x6e\x67\x20\x64\x61\x65\x6d\x6f\x6e\x20\x6f\x6e\x20\x70\x6f\x72\x74\x20\x24\x44\x41\x45\x4d\x4f\x4e\x5f\x50\x4f\x52\x54\x22\x0a\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x72\x75\x6e\x20\x2d\x64\x20\x5c\x0a\x20\x20\x20\x20\x2d\x2d\x72\x65\x73\x74\x61\x72\x74\x20\x75\x6e\x6c\x65\x73\x73\x2d\x73\x74\x6f\x70\x70\x65\x64\x20\x5c\x0a\x20\x20\x20\x20\x2d\x70\x20\x22\x24\x44\x41\x45\x4d\x4f\x4e\x5f\x50\x4f\x52\x54\x22\x3a\x22\x24\x43\x4f\x4e\x54\x41\x49\x4e\x45\x52\x5f\x50\x4f\x52\x54\x22\x20\x5c\x0a\x20\x20\x20\x20\x2d\x76\x20\x2f\x76\x61\x72\x2f\x72\x75\x6e\x2f\x64\x6f\x63\x6b\x65\x72\x2e\x73\x6f\x63\x6b\x3a\x2f\x76\x61\x72\x2f\x72\x75\x6e\x2f\x64\x6f\x63\x6b\x65\x72\x2e\x73\x6f\x63\x6b\x20\x5c\x0a\x20\x20\x20\x20\x2d\x76\x20\x22\x24\x48\x4f\x4d\x45\x22\x3a\x2f\x61\x70\x70\x2f\x68\x6f\x73\x74\x20\x5c\x0a\x20\x20\x20\x20\x2d\x65\x20\x48\x4f\x4d\x45\x3d\x22\x24\x48\x4f\x4d\x45\x22\x20\x5c\x0a\x20\x20\x20\x20\x2d\x65\x20\x53\x53\x48\x5f\x4b\x4e\x4f\x57\x4e\x5f\x48\x4f\x53\x54\x53\x3d\x27\x2f\x61\x70\x70\x2f\x68\x6f\x73\x74\x2f\x2e\x73\x73\x68\x2f\x6b\x6e\x6f\x77\x6e\x5f\x68\x6f\x73\x74\x73\x27\x20\x5c\x0a\x20\x20\x20\x20\x2d\x2d\x6e\x61\x6d\x65\x20\x22\x24\x44\x41\x45\x4d\x4f\x4e\x5f\x4e\x41\x4d\x45\x22\x20\x5c\x0a\x20\x20\x20\x20\x22\x24\x49\x4d\x41\x47\x45\x22\x20\x22\x24\x48\x4f\x53\x54\x5f\x41\x44\x44\x52\x45\x53\x53\x20\x2d\x2d\x77\x65\x62\x68\x6f\x6f\x6b\x2e\x73\x65\x63\x72\x65\x74\x20\x24\x57\x45\x42\x48\x4f\x4f\x4b\x5f\x53\x45\x43\x52\x45\x54\x22\x20\x3e\x20\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x23\x20\x32\x3e\x26\x31\x0a")

// FileClientScriptsDockerSh is "client/scripts/docker.sh"
var FileClientScriptsDockerSh = []byte("\x23\x21\x2f\x62\x69\x6e\x2f\x73\x68\x0a\x0a\x23\x20\x42\x6f\x6f\x74\x73\x74\x72\x61\x70\x73\x20\x61\x20\x6d\x61\x63\x68\x69\x6e\x65\x20\x66\x6f\x72\x20\x64\x6f\x63\x6b\x65\x72\x2e\x0a\x0a\x73\x65\x74\x20\x2d\x65\x0a\x0a\x44\x4f\x43\x4b\x45\x52\x5f\x53\x4f\x55\x52\x43\x45\x3d\x68\x74\x74\x70\x73\x3a\x2f\x2f\x67\x65\x74\x2e\x64\x6f\x63\x6b\x65\x72\x2e\x63\x6f\x6d\x0a\x44\x4f\x43\x4b\x45\x52\x5f\x44\x45\x53\x54\x3d\x22\x2f\x74\x6d\x70\x2f\x67\x65\x74\x2d\x64\x6f\x63\x6b\x65\x72\x2e\x73\x68\x22\x0a\x0a\x73\x74\x61\x72\x74\x44\x6f\x63\x6b\x65\x72\x64\x28\x29\x20\x7b\x0a\x20\x20\x20\x20\x23\x20\x53\x74\x61\x72\x74\x20\x64\x6f\x63\x6b\x65\x72\x64\x20\x69\x66\x20\x69\x74\x20\x69\x73\x20\x6e\x6f\x74\x20\x6f\x6e\x6c\x69\x6e\x65\x0a\x20\x20\x20\x20\x69\x66\x20\x21\x20\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x73\x74\x61\x74\x73\x20\x2d\x2d\x6e\x6f\x2d\x73\x74\x72\x65\x61\x6d\x20\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x20\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x23\x20\x46\x61\x6c\x6c\x20\x62\x61\x63\x6b\x20\x74\x6f\x20\x73\x79\x73\x74\x65\x6d\x63\x74\x6c\x20\x69\x66\x20\x73\x65\x72\x76\x69\x63\x65\x20\x64\x6f\x65\x73\x6e\x22\x74\x20\x77\x6f\x72\x6b\x2c\x20\x6f\x74\x68\x65\x72\x77\x69\x73\x65\x20\x6a\x75\x73\x74\x20\x72\x75\x6e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x23\x20\x64\x6f\x63\x6b\x65\x72\x64\x20\x69\x6e\x20\x62\x61\x63\x6b\x67\x72\x6f\x75\x6e\x64\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x64\x6f\x63\x6b\x65\x72\x64\x20\x69\x73\x20\x6f\x66\x66\x6c\x69\x6e\x65\x20\x2d\x20\x73\x74\x61\x72\x74\x69\x6e\x67\x20\x64\x6f\x63\x6b\x65\x72\x64\x2e\x2e\x2e\x22\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x73\x65\x72\x76\x69\x63\x65\x20\x64\x6f\x63\x6b\x65\x72\x20\x73\x74\x61\x72\x74\x20\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x20\x5c\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x7c\x7c\x20\x73\x75\x64\x6f\x20\x73\x79\x73\x74\x65\x6d\x63\x74\x6c\x20\x73\x74\x61\x72\x74\x20\x64\x6f\x63\x6b\x65\x72\x20\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x20\x5c\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x7c\x7c\x20\x28\x20\x73\x75\x64\x6f\x20\x6e\x6f\x68\x75\x70\x20\x64\x6f\x63\x6b\x65\x72\x64\x20\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x20\x26\x20\x29\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x64\x6f\x63\x6b\x65\x72\x64\x20\x73\x74\x61\x72\x74\x65\x64\x22\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x23\x20\x50\x6f\x6c\x6c\x20\x75\x6e\x74\x69\x6c\x20\x64\x6f\x63\x6b\x65\x72\x64\x20\x69\x73\x20\x72\x75\x6e\x6e\x69\x6e\x67\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x77\x68\x69\x6c\x65\x20\x21\x20\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x73\x74\x61\x74\x73\x20\x2d\x2d\x6e\x6f\x2d\x73\x74\x72\x65\x61\x6d\x20\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x20\x3b\x20\x64\x6f\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x57\x61\x69\x74\x69\x6e\x67\x20\x66\x6f\x72\x20\x64\x6f\x63\x6b\x65\x72\x64\x20\x74\x6f\x20\x63\x6f\x6d\x65\x20\x6f\x6e\x6c\x69\x6e\x65\x2e\x2e\x2e\x22\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x73\x6c\x65\x65\x70\x20\x31\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x64\x6f\x6e\x65\x0a\x20\x20\x20\x20\x66\x69\x3b\x0a\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x64\x6f\x63\x6b\x65\x72\x64\x20\x69\x73\x20\x6f\x6e\x6c\x69\x6e\x65\x22\x0a\x7d\x0a\x0a\x23\x20\x53\x6b\x69\x70\x20\x69\x6e\x73\x74\x61\x6c\x6c\x61\x74\x69\x6f\x6e\x20\x69\x66\x20\x44\x6f\x63\x6b\x65\x72\x20\x69\x73\x20\x61\x6c\x72\x65\x61\x64\x79\x20\x69\x6e\x73\x74\x61\x6c\x6c\x65\x64\x2e\x0a\x69\x66\x20\x68\x61\x73\x68\x20\x64\x6f\x63\x6b\x65\x72\x20\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x44\x6f\x63\x6b\x65\x72\x20\x69\x6e\x73\x74\x61\x6c\x6c\x61\x74\x69\x6f\x6e\x20\x64\x65\x74\x65\x63\x74\x65\x64\x20\x2d\x20\x73\x6b\x69\x70\x70\x69\x6e\x67\x20\x69\x6e\x73\x74\x61\x6c\x6c\x22\x0a\x20\x20\x20\x20\x73\x74\x61\x72\x74\x44\x6f\x63\x6b\x65\x72\x64\x0a\x20\x20\x20\x20\x65\x78\x69\x74\x20\x30\x0a\x66\x69\x3b\x0a\x0a\x66\x65\x74\x63\x68\x66\x69\x6c\x65\x28\x29\x20\x7b\x0a\x20\x20\x20\x20\x23\x20\x41\x72\x67\x73\x3a\x0a\x20\x20\x20\x20\x23\x20\x20\x20\x24\x31\x20\x73\x6f\x75\x72\x63\x65\x20\x55\x52\x4c\x0a\x20\x20\x20\x20\x23\x20\x20\x20\x24\x32\x20\x64\x65\x73\x74\x69\x6e\x61\x74\x69\x6f\x6e\x20\x66\x69\x6c\x65\x2e\x0a\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x53\x61\x76\x69\x6e\x67\x20\x24\x31\x20\x74\x6f\x20\x24\x32\x22\x0a\x20\x20\x20\x20\x69\x66\x20\x68\x61\x73\x68\x20\x63\x75\x72\x6c\x20\x32\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x63\x75\x72\x6c\x20\x2d\x66\x73\x53\x4c\x20\x22\x24\x31\x22\x20\x2d\x6f\x20\x22\x24\x32\x22\x0a\x20\x20\x20\x20\x65\x6c\x69\x66\x20\x68\x61\x73\x68\x20\x77\x67\x65\x74\x20\x32\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x77\x67\x65\x74\x20\x2d\x4f\x20\x22\x24\x32\x22\x20\x22\x24\x31\x22\x0a\x20\x20\x20\x20\x65\x6c\x73\x65\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x72\x65\x74\x75\x72\x6e\x20\x31\x0a\x20\x20\x20\x20\x66\x69\x3b\x0a\x7d\x0a\x0a\x65\x63\x68\x6f\x20\x22\x49\x6e\x73\x74\x61\x6c\x6c\x69\x6e\x67\x20\x64\x6f\x63\x6b\x65\x72\x2e\x2e\x2e\x22\x0a\x0a\x23\x20\x41\x6d\x61\x7a\x6f\x6e\x20\x45\x43\x53\x20\x69\x6e\x73\x74\x61\x6e\x63\x65\x73\x20\x72\x65\x71\x75\x69\x72\x65\x20\x63\x75\x73\x74\x6f\x6d\x20\x69\x6e\x73\x74\x61\x6c\x6c\x0a\x69\x66\x20\x67\x72\x65\x70\x20\x2d\x71\x20\x41\x6d\x61\x7a\x6f\x6e\x20\x2f\x65\x74\x63\x2f\x73\x79\x73\x74\x65\x6d\x2d\x72\x65\x6c\x65\x61\x73\x65\x20\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x32\x3e\x26\x31\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x65\x63\x68\x6f\x20\x22\x41\x6d\x61\x7a\x6f\x6e\x4f\x53\x20\x64\x65\x74\x65\x63\x74\x65\x64\x22\x0a\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x79\x75\x6d\x20\x69\x6e\x73\x74\x61\x6c\x6c\x20\x2d\x79\x20\x64\x6f\x63\x6b\x65\x72\x0a\x65\x6c\x73\x65\x0a\x20\x20\x20\x20\x23\x20\x54\x72\x79\x20\x74\x6f\x20\x64\x6f\x77\x6e\x6c\x6f\x61\x64\x20\x75\x73\x69\x6e\x67\x20\x63\x75\x72\x6c\x20\x6f\x72\x20\x77\x67\x65\x74\x2c\x0a\x20\x20\x20\x20\x23\x20\x62\x65\x66\x6f\x72\x65\x20\x72\x65\x73\x6f\x72\x74\x69\x6e\x67\x20\x74\x6f\x20\x69\x6e\x73\x74\x61\x6c\x6c\x69\x6e\x67\x20\x63\x75\x72\x6c\x2e\x0a\x20\x20\x20\x20\x69\x66\x20\x66\x65\x74\x63\x68\x66\x69\x6c\x65\x20\x24\x44\x4f\x43\x4b\x45\x52\x5f\x53\x4f\x55\x52\x43\x45\x20\x24\x44\x4f\x43\x4b\x45\x52\x5f\x44\x45\x53\x54\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x73\x68\x20\x24\x44\x4f\x43\x4b\x45\x52\x5f\x44\x45\x53\x54\x0a\x20\x20\x20\x20\x65\x6c\x73\x65\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x61\x70\x74\x2d\x67\x65\x74\x20\x75\x70\x64\x61\x74\x65\x20\x26\x26\x20\x61\x70\x74\x2d\x67\x65\x74\x20\x2d\x79\x20\x69\x6e\x73\x74\x61\x6c\x6c\x20\x63\x75\x72\x6c\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x66\x65\x74\x63\x68\x66\x69\x6c\x65\x20\x24\x44\x4f\x43\x4b\x45\x52\x5f\x53\x4f\x55\x52\x43\x45\x20\x24\x44\x4f\x43\x4b\x45\x52\x5f\x44\x45\x53\x54\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x73\x68\x20\x24\x44\x4f\x43\x4b\x45\x52\x5f\x44\x45\x53\x54\x0a\x20\x20\x20\x20\x66\x69\x3b\x0a\x66\x69\x3b\x0a\x0a\x73\x74\x61\x72\x74\x44\x6f\x63\x6b\x65\x72\x64\x0a\x0a\x65\x63\x68\x6f\x20\x22\x44\x6f\x63\x6b\x65\x72\x20\x69\x6e\x73\x74\x61\x6c\x6c\x61\x74\x69\x6f\x6e\x20\x63\x6f\x6d\x70\x6c\x65\x74\x65\x22\x0a\x0a\x65\x78\x69\x74\x20\x30\x0a")

// FileClientScriptsInertiaDownSh is "client/scripts/inertia-down.sh"
var FileClientScriptsInertiaDownSh = []byte("\x23\x21\x2f\x62\x69\x6e\x2f\x73\x68\x0a\x0a\x23\x20\x42\x61\x73\x69\x63\x20\x73\x63\x72\x69\x70\x74\x20\x66\x6f\x72\x20\x62\x72\x69\x6e\x67\x69\x6e\x67\x20\x64\x6f\x77\x6e\x20\x49\x6e\x65\x72\x74\x69\x61\x2e\x0a\x0a\x73\x65\x74\x20\x2d\x65\x0a\x0a\x23\x20\x52\x65\x6d\x6f\x76\x65\x20\x49\x6e\x65\x72\x74\x69\x61\x20\x66\x72\x6f\x6d\x20\x56\x50\x53\x0a\x73\x75\x64\x6f\x20\x72\x6d\x20\x2d\x72\x66\x20\x7e\x2f\x69\x6e\x65\x72\x74\x69\x61\x2f\x0a\x73\x75\x64\x6f\x20\x72\x6d\x20\x2d\x72\x66\x20\x7e\x2f\x2e\x69\x6e\x65\x72\x74\x69\x61\x2f\x0a")

// FileClientScriptsKeygenSh is "client/scripts/keygen.sh"
var FileClientScriptsKeygenSh = []byte("\x23\x21\x2f\x62\x69\x6e\x2f\x73\x68\x0a\x0a\x23\x20\x50\x72\x6f\x64\x75\x63\x65\x73\x20\x61\x20\x70\x75\x62\x6c\x69\x63\x2d\x70\x72\x69\x76\x61\x74\x65\x20\x6b\x65\x79\x2d\x70\x61\x69\x72\x20\x61\x6e\x64\x20\x6f\x75\x74\x70\x75\x74\x73\x20\x74\x68\x65\x20\x70\x75\x62\x6c\x69\x63\x20\x6b\x65\x79\x2e\x0a\x0a\x73\x65\x74\x20\x2d\x65\x0a\x0a\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x3d\x24\x48\x4f\x4d\x45\x2f\x2e\x73\x73\x68\x2f\x69\x64\x5f\x72\x73\x61\x5f\x69\x6e\x65\x72\x74\x69\x61\x5f\x64\x65\x70\x6c\x6f\x79\x0a\x50\x55\x42\x5f\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x3d\x24\x48\x4f\x4d\x45\x2f\x2e\x73\x73\x68\x2f\x69\x64\x5f\x72\x73\x61\x5f\x69\x6e\x65\x72\x74\x69\x61\x5f\x64\x65\x70\x6c\x6f\x79\x2e\x70\x75\x62\x0a\x0a\x23\x20\x49\x6e\x73\x74\x61\x6c\x6c\x20\x6f\x70\x65\x6e\x73\x73\x68\x20\x69\x66\x20\x73\x73\x68\x2d\x6b\x65\x79\x67\x65\x6e\x20\x69\x73\x20\x6e\x6f\x74\x20\x61\x76\x61\x69\x6c\x61\x62\x6c\x65\x0a\x69\x66\x20\x21\x20\x68\x61\x73\x68\x20\x73\x73\x68\x2d\x6b\x65\x79\x67\x65\x6e\x20\x32\x3e\x2f\x64\x65\x76\x2f\x6e\x75\x6c\x6c\x20\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x73\x75\x64\x6f\x20\x61\x70\x74\x2d\x67\x65\x74\x20\x69\x6e\x73\x74\x61\x6c\x6c\x20\x6f\x70\x65\x6e\x73\x73\x68\x2d\x63\x6c\x69\x65\x6e\x74\x20\x7c\x7c\x20\x73\x75\x64\x6f\x20\x61\x70\x74\x20\x69\x6e\x73\x74\x61\x6c\x6c\x20\x6f\x70\x65\x6e\x73\x73\x68\x2d\x63\x6c\x69\x65\x6e\x74\x0a\x66\x69\x3b\x0a\x0a\x23\x20\x43\x68\x65\x63\x6b\x20\x69\x66\x20\x64\x65\x73\x74\x69\x6e\x61\x74\x69\x6f\x6e\x20\x66\x69\x6c\x65\x20\x61\x6c\x72\x65\x61\x64\x79\x20\x65\x78\x69\x73\x74\x73\x0a\x69\x66\x20\x5b\x20\x2d\x66\x20\x22\x24\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x22\x20\x5d\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x69\x66\x20\x5b\x20\x21\x20\x2d\x66\x20\x22\x24\x50\x55\x42\x5f\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x22\x20\x5d\x3b\x20\x74\x68\x65\x6e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x23\x20\x49\x66\x20\x70\x75\x62\x6c\x69\x63\x20\x6b\x65\x79\x20\x64\x6f\x65\x73\x6e\x74\x20\x65\x78\x69\x73\x74\x2c\x20\x6d\x61\x6b\x65\x20\x69\x74\x2e\x0a\x20\x20\x20\x20\x20\x20\x20\x20\x73\x73\x68\x2d\x6b\x65\x79\x67\x65\x6e\x20\x2d\x79\x20\x2d\x66\x20\x22\x24\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x22\x20\x3e\x20\x22\x24\x50\x55\x42\x5f\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x22\x0a\x20\x20\x20\x20\x66\x69\x3b\x0a\x65\x6c\x73\x65\x0a\x20\x20\x20\x20\x23\x20\x47\x65\x6e\x65\x72\x61\x74\x65\x20\x6b\x65\x79\x20\x77\x69\x74\x68\x20\x6e\x6f\x20\x70\x61\x73\x73\x77\x6f\x72\x64\x2e\x0a\x20\x20\x20\x20\x73\x73\x68\x2d\x6b\x65\x79\x67\x65\x6e\x20\x2d\x66\x20\x22\x24\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x22\x20\x2d\x74\x20\x72\x73\x61\x20\x2d\x4e\x20\x27\x27\x0a\x66\x69\x0a\x0a\x73\x73\x68\x2d\x6b\x65\x79\x73\x63\x61\x6e\x20\x67\x69\x74\x68\x75\x62\x2e\x63\x6f\x6d\x20\x3e\x3e\x20\x7e\x2f\x2e\x73\x73\x68\x2f\x6b\x6e\x6f\x77\x6e\x5f\x68\x6f\x73\x74\x73\x0a\x0a\x63\x61\x74\x20\x22\x24\x50\x55\x42\x5f\x49\x44\x5f\x44\x45\x53\x54\x49\x4e\x41\x54\x49\x4f\x4e\x22\x0a")

// FileClientScriptsTokenSh is "client/scripts/token.sh"
var FileClientScriptsTokenSh = []byte("\x23\x21\x2f\x62\x69\x6e\x2f\x73\x68\x0a\x0a\x73\x65\x74\x20\x2d\x65\x0a\x0a\x23\x20\x55\x73\x65\x72\x20\x61\x72\x67\x75\x6d\x65\x6e\x74\x2e\x0a\x52\x45\x4c\x45\x41\x53\x45\x3d\x25\x73\x0a\x0a\x23\x20\x47\x65\x6e\x65\x72\x61\x74\x65\x20\x61\x20\x64\x61\x65\x6d\x6f\x6e\x20\x74\x6f\x6b\x65\x6e\x20\x75\x73\x69\x6e\x67\x20\x43\x4c\x49\x20\x66\x6f\x72\x20\x41\x50\x49\x20\x72\x65\x71\x75\x65\x73\x74\x73\x2e\x0a\x73\x75\x64\x6f\x20\x64\x6f\x63\x6b\x65\x72\x20\x72\x75\x6e\x20\x2d\x2d\x72\x6d\x20\x5c\x0a\x20\x20\x20\x20\x2d\x76\x20\x24\x48\x4f\x4d\x45\x3a\x2f\x61\x70\x70\x2f\x68\x6f\x73\x74\x20\x5c\x0a\x20\x20\x20\x20\x2d\x65\x20\x53\x53\x48\x5f\x4b\x4e\x4f\x57\x4e\x5f\x48\x4f\x53\x54\x53\x3d\x27\x2f\x61\x70\x70\x2f\x68\x6f\x73\x74\x2f\x2e\x73\x73\x68\x2f\x6b\x6e\x6f\x77\x6e\x5f\x68\x6f\x73\x74\x73\x27\x20\x5c\x0a\x20\x20\x20\x20\x2d\x65\x20\x48\x4f\x4d\x45\x3d\x24\x48\x4f\x4d\x45\x20\x5c\x0a\x20\x20\x20\x20\x2d\x2d\x65\x6e\x74\x72\x79\x70\x6f\x69\x6e\x74\x3d\x69\x6e\x65\x72\x74\x69\x61\x64\x20\x5c\x0a\x20\x20\x20\x20\x75\x62\x63\x6c\x61\x75\x6e\x63\x68\x70\x61\x64\x2f\x69\x6e\x65\x72\x74\x69\x61\x3a\x24\x52\x45\x4c\x45\x41\x53\x45\x20\x74\x6f\x6b\x65\x6e\x0a")

func init() {
	err := CTX.Err()
	if err != nil {
		panic(err)
	}

	err = FS.Mkdir(CTX, "client/", 0777)
	if err != nil && err != os.ErrExist {
		panic(err)
	}

	err = FS.Mkdir(CTX, "client/scripts/", 0777)
	if err != nil && err != os.ErrExist {
		panic(err)
	}

	var f webdav.File

	f, err = FS.OpenFile(CTX, "client/scripts/daemon-down.sh", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(FileClientScriptsDaemonDownSh)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	f, err = FS.OpenFile(CTX, "client/scripts/daemon-up.sh", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(FileClientScriptsDaemonUpSh)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	f, err = FS.OpenFile(CTX, "client/scripts/docker.sh", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(FileClientScriptsDockerSh)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	f, err = FS.OpenFile(CTX, "client/scripts/inertia-down.sh", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(FileClientScriptsInertiaDownSh)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	f, err = FS.OpenFile(CTX, "client/scripts/keygen.sh", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(FileClientScriptsKeygenSh)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	f, err = FS.OpenFile(CTX, "client/scripts/token.sh", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(FileClientScriptsTokenSh)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	Handler = &webdav.Handler{
		FileSystem: FS,
		LockSystem: webdav.NewMemLS(),
	}

}

// Open a file
func (hfs *HTTPFS) Open(path string) (http.File, error) {
	path = hfs.Prefix + path

	f, err := FS.OpenFile(CTX, path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// ReadFile is adapTed from ioutil
func ReadFile(path string) ([]byte, error) {
	f, err := FS.OpenFile(CTX, path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))

	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(f)
	return buf.Bytes(), err
}

// WriteFile is adapTed from ioutil
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := FS.OpenFile(CTX, filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

// WalkDirs looks for files in the given dir and returns a list of files in it
// usage for all files in the b0x: WalkDirs("", false)
func WalkDirs(name string, includeDirsInList bool, files ...string) ([]string, error) {
	f, err := FS.OpenFile(CTX, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	fileInfos, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		filename := path.Join(name, info.Name())

		if includeDirsInList || !info.IsDir() {
			files = append(files, filename)
		}

		if info.IsDir() {
			files, err = WalkDirs(filename, includeDirsInList, files...)
			if err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}
