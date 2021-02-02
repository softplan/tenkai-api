// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Code generated by "go generate". DO NOT EDIT.

package stacktrace

import (
	"strings"

	radix "github.com/armon/go-radix"
)

var libraryPackages = newLibraryPackagesRadixTree(
	"vendor/golang_org",
	"archive/tar",
	"archive/zip",
	"bufio",
	"bytes",
	"compress/bzip2",
	"compress/flate",
	"compress/gzip",
	"compress/lzw",
	"compress/zlib",
	"container/heap",
	"container/list",
	"container/ring",
	"context",
	"crypto",
	"crypto/aes",
	"crypto/cipher",
	"crypto/des",
	"crypto/dsa",
	"crypto/ecdsa",
	"crypto/elliptic",
	"crypto/hmac",
	"crypto/internal/randutil",
	"crypto/internal/subtle",
	"crypto/md5",
	"crypto/rand",
	"crypto/rc4",
	"crypto/rsa",
	"crypto/sha1",
	"crypto/sha256",
	"crypto/sha512",
	"crypto/subtle",
	"crypto/tls",
	"crypto/x509",
	"crypto/x509/pkix",
	"database/sql",
	"database/sql/driver",
	"debug/dwarf",
	"debug/elf",
	"debug/gosym",
	"debug/macho",
	"debug/pe",
	"debug/plan9obj",
	"encoding",
	"encoding/ascii85",
	"encoding/asn1",
	"encoding/base32",
	"encoding/base64",
	"encoding/binary",
	"encoding/csv",
	"encoding/gob",
	"encoding/hex",
	"encoding/json",
	"encoding/pem",
	"encoding/xml",
	"errors",
	"expvar",
	"flag",
	"fmt",
	"go/ast",
	"go/build",
	"go/constant",
	"go/doc",
	"go/format",
	"go/importer",
	"go/internal/gccgoimporter",
	"go/internal/gcimporter",
	"go/internal/srcimporter",
	"go/parser",
	"go/printer",
	"go/scanner",
	"go/token",
	"go/types",
	"hash",
	"hash/adler32",
	"hash/crc32",
	"hash/crc64",
	"hash/fnv",
	"html",
	"html/template",
	"image",
	"image/color",
	"image/color/palette",
	"image/draw",
	"image/gif",
	"image/internal/imageutil",
	"image/jpeg",
	"image/png",
	"index/suffixarray",
	"internal/bytealg",
	"internal/cpu",
	"internal/nettrace",
	"internal/poll",
	"internal/race",
	"internal/singleflight",
	"internal/syscall/unix",
	"internal/syscall/windows",
	"internal/syscall/windows/registry",
	"internal/syscall/windows/sysdll",
	"internal/testenv",
	"internal/testlog",
	"internal/trace",
	"io",
	"io/ioutil",
	"log",
	"log/syslog",
	"math",
	"math/big",
	"math/bits",
	"math/cmplx",
	"math/rand",
	"mime",
	"mime/multipart",
	"mime/quotedprintable",
	"net",
	"net/http",
	"net/http/cgi",
	"net/http/cookiejar",
	"net/http/fcgi",
	"net/http/httptest",
	"net/http/httptrace",
	"net/http/httputil",
	"net/http/internal",
	"net/http/pprof",
	"net/internal/socktest",
	"net/mail",
	"net/rpc",
	"net/rpc/jsonrpc",
	"net/smtp",
	"net/textproto",
	"net/url",
	"os",
	"os/exec",
	"os/signal",
	"os/signal/internal/pty",
	"os/user",
	"path",
	"path/filepath",
	"plugin",
	"reflect",
	"regexp",
	"regexp/syntax",
	"runtime",
	"runtime/cgo",
	"runtime/debug",
	"runtime/internal/atomic",
	"runtime/internal/sys",
	"runtime/pprof",
	"runtime/pprof/internal/profile",
	"runtime/race",
	"runtime/trace",
	"sort",
	"strconv",
	"strings",
	"sync",
	"sync/atomic",
	"syscall",
	"testing",
	"testing/internal/testdeps",
	"testing/iotest",
	"testing/quick",
	"text/scanner",
	"text/tabwriter",
	"text/template",
	"text/template/parse",
	"time",
	"unicode",
	"unicode/utf16",
	"unicode/utf8",
	"unsafe",
	"go.elastic.co/apm",
)

func newLibraryPackagesRadixTree(k ...string) *radix.Tree {
	tree := radix.New()
	for _, k := range k {
		tree.Insert(k, true)
	}
	return tree
}

// RegisterLibraryPackage registers the given packages as being
// well-known library path prefixes. This must not be called
// concurrently with any other functions or methods in this
// package; it is expected to be used by init functions.
func RegisterLibraryPackage(pkg ...string) {
	for _, pkg := range pkg {
		libraryPackages.Insert(pkg, true)
	}
}

// RegisterApplicationPackage registers the given packages as being
// an application path. This must not be called concurrently with
// any other functions or methods in this package; it is expected
// to be used by init functions.
//
// It is not typically necessary to register application paths. If
// a package does not match a registered *library* package path
// prefix, then the path is considered an application path. This
// function exists for the unusual case that an application exists
// within a library (e.g. an example program).
func RegisterApplicationPackage(pkg ...string) {
	for _, pkg := range pkg {
		libraryPackages.Insert(pkg, false)
	}
}

// IsLibraryPackage reports whether or not the given package path is
// a library package. This includes known library packages
// (e.g. stdlib or apm-agent-go), vendored packages, and any packages
// with a prefix registered with RegisterLibraryPackage but not
// RegisterApplicationPackage.
func IsLibraryPackage(pkg string) bool {
	if strings.HasSuffix(pkg, "_test") {
		return false
	}
	if strings.Contains(pkg, "/vendor/") {
		return true
	}
	prefix, v, ok := libraryPackages.LongestPrefix(pkg)
	if !ok || v == false {
		return false
	}
	return prefix == pkg || pkg[len(prefix)] == '/'
}
