package main

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func DepCheck(pkg, src string, re bool) {
	x, er := ListPkgImports(pkg, src)
	if er != nil {
		fmt.Println("Get imports error:", er)
	}
	for i := range x {
		fmt.Println(x[i])
	}
}
func PackPackage(src, target string, clearGit, jumpOver bool) {

	AbsSrc, er := filepath.Abs(src)
	if er != nil {
		fmt.Println("Error convert path!", er)
		os.Exit(1)
	}
	AbsTar, er := filepath.Abs(target)
	if er != nil {
		fmt.Println("Error convert path!", er)
		os.Exit(1)
	}
	var Pkg []string
	Pkg, er = WalkPkg(AbsSrc)
	if er != nil {
		fmt.Println("Error walk path!", er)
		os.Exit(1)
	}
	fmt.Println("Find Packages:", Pkg)
	fmt.Println("pack packages")
	for i := range Pkg {
		name, er := filepath.Rel(AbsSrc, Pkg[i])
		if er != nil {
			fmt.Println("Error convert path!", Pkg[i], er)
		}
		name = strings.Replace(name, string(os.PathSeparator), "+", -1)
		if clearGit {
			fmt.Println("Try remove .git ", Pkg[i])
			RemoveGit(Pkg[i])
		}
		/*		fmt.Println("Gen Deps ", Pkg[i])
				GenDep(Pkg[i])*/
		fmt.Println("Packgeing ", Pkg[i])
		Pack(Pkg[i], name, AbsTar, jumpOver)
		dep := filepath.Join(Pkg[i], "dep.txt")
		if FileExist(dep) {
			os.RemoveAll(dep)
		}
	}
}

func InstallPackage(src, target, pkg string) {
	pkg = strings.Replace(pkg, "/", string(os.PathSeparator), -1)
	AbsSrc, er := filepath.Abs(src)
	if er != nil {
		fmt.Println("[S]ource path error:", er)
		os.Exit(1)
	}
	AbsTar, err := filepath.Abs(target)
	if err != nil {
		fmt.Println("[T]arget path error:", er)
		os.Exit(1)
	}
	if !IsDir(AbsSrc) {
		fmt.Println("[S]ource path error: not folder")
		os.Exit(1)
	}
	if !IsDir(AbsTar) {
		fmt.Println("[T]arget path error: not folder")
		os.Exit(1)
	}
	if ExtraPkg(pkg, AbsSrc, AbsTar) {
		fmt.Println("Install package success! may be some depends not installed,Pls check log!")
		return
	} else {
		fmt.Println("Install package Failed!")
	}
}

//ReadDirNames list all file names under path p
func ReadDirNames(p string) ([]string, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}
func IsDir(p string) bool {
	stat, er := os.Stat(p)
	if er != nil {
		return false
	}
	return stat.IsDir()
}

func Pack(p, name, Target string, jumpOver bool) {

	lf, _ := os.OpenFile(LogPath, os.O_APPEND, 0644)
	defer lf.Close()

	tar := filepath.Join(Target, name+".7z")
	if FileExist(tar) {
		if jumpOver {
			os.Remove(tar)
		} else {
			fmt.Fprintln(lf, "pass ", tar)
			fmt.Println("pass ", tar)
			return
		}
	}

	cmd := exec.Command(z7a, "a", tar, p)
	if er := cmd.Start(); er != nil {
		fmt.Fprintln(lf, "Compress Error", er)
		fmt.Println("Compress Error", er)
		os.Exit(1)
	}
	cmd.Wait()

}
func FileExist(p string) bool {
	if _, ok := os.Stat(p); os.IsNotExist(ok) {
		return false
	}
	return true
}
func RemoveGit(p string) {
	git := filepath.Join(p, ".git")
	if FileExist(git) {
		if er := os.RemoveAll(git); er != nil {
			panic(er)
		}
	}

}
func WalkPkg(P string) ([]string, error) {
	var (
		Dir  []string //第二层
		TDir []string //第一层
		Pkg  []string
		er   error
		Abs  string
	)
	Abs, er = filepath.Abs(P) //绝对路径 **.src
	if er != nil {
		return nil, er
	}
	//读取第一层
	TDir, er = ReadDirNames(Abs) //读取源列表 src/...
	if er != nil {
		return nil, er
	}
	for I := 1; I < 5; I++ {
		var outD, loopD *[]string
		if I%2 == 0 { //偶数
			loopD = &Dir
			TDir = []string{}
			outD = &TDir
		} else {
			loopD = &TDir
			Dir = []string{}
			outD = &Dir

		}
		for k := range *loopD {
			//tp := filepath.Join(Abs, Dir[k]) //获取绝对路径 src/github.com/Abs
			if !strings.Contains((*loopD)[k], ":") {
				(*loopD)[k] = filepath.Join(Abs, (*loopD)[k])
			}
			TD1 := []string{} //下层目录
			Last := false
			if IsDir((*loopD)[k]) { //是目录
				TD, er := ReadDirNames((*loopD)[k]) //读取
				if er != nil {
					return nil, er
				}
				for i := range TD { //遍历读取的数据
					f := filepath.Join((*loopD)[k], TD[i])
					if IsDir(f) && strings.Index(TD[i], ".") != 0 { //跳过.文件和 .文件夹
						TD1 = append(TD1, filepath.Join((*loopD)[k], TD[i]))
					} else if strings.Contains(TD[i], ".git") ||
						strings.Contains(TD[i], ".go") ||
						strings.Contains(TD[i], "README") ||
						strings.Contains(TD[i], ".gitignore") { //包含git或源代码
						Last = true
					}
				}
				if Last {
					Pkg = append(Pkg, (*loopD)[k]) //添加到包
				} else {
					*outD = append(*outD, TD1...) //添加到继续遍历
				}
			}
		}
	}
	return Pkg, nil
}
func ExtraPkg(pk, src, tar string) bool {

	lf, _ := os.OpenFile(LogPath, os.O_APPEND, 0644)
	defer lf.Close()

	pk = strings.Replace(pk, "/", string(os.PathSeparator), -1)
	pkgT := filepath.Join(tar, pk)
	pkgf := filepath.Join(src, PkgToZip(pk)) + ".7z"
	//fmt.Fprintln(lf, "Package file1:", pkgf)
	if !FileExist(pkgf) {
		//check uper package is exists?
		PKS := strings.Split(pk, string(os.PathSeparator))
		for i := len(PKS); i >= 0; i-- {
			pkn := filepath.Join(src, PkgToZip(JoinPkg(PKS[:i]...))) + ".7z"
			//	fmt.Fprintln(lf, "Package file test1:", pkn)
			if !FileExist(pkn) {
				continue
			} else {
				pkgf = pkn
			}
		}
		if !FileExist(pkgf) {
			fmt.Fprintln(lf, "Package file2", pkgf, " not found!")
			fmt.Println("Package file", pkgf, " not found!")
			return false
		}
	}
	if !UnPack(pkgf, pkgT) {
		return false
	}
	dep := filepath.Join(tar, ZipToPkg(pk))
	if Deps, ok := DepLoad(pk, dep); ok && len(Deps) > 0 {
		for _, x := range Deps {
			//x =
			ExtraPkg(x, src, tar)
		}
	}
	return true
}
func JoinPkg(p ...string) string {
	return filepath.Join(p...)
}
func UnPack(f, p string) bool {

	lf, _ := os.OpenFile(LogPath, os.O_APPEND, 0644)
	defer lf.Close()

	tar := filepath.Dir(p)
	if FileExist(p) {
		fmt.Fprintln(lf, "Package already exists", p)
		fmt.Println("Package already exists")
		os.Exit(1)
	} else {
		os.MkdirAll(tar, 0644)
	}
	cmd := exec.Command(z7a, "x", "-o"+tar, f)
	cmd.Stdout = lf
	cmd.Stderr = lf
	if er := cmd.Start(); er != nil {
		fmt.Fprintln(lf, "Decompress error:", er)
		fmt.Println("Decompress error:", er)
		os.Exit(1)
	}
	cmd.Wait()
	return true
}

func DepLoad(pkg, src string) ([]string, bool) {
	pk, er := ListPkgImports(pkg, src)
	return pk, er == nil
}

//ZipToPkg convert package 7z name to package name
func ZipToPkg(p string) string {
	return strings.Replace(p, "+", string(os.PathSeparator), -1)
}

//PkgToZip convert package name to package 7z name(with out .7z)
func PkgToZip(p string) string {
	return strings.Replace(p, string(os.PathSeparator), "+", -1)
}

func ListImports(importPath, srcPath, rootPath string) ([]string, error) {
	ctxt := build.Default
	ctxt.BuildTags = []string{}
	ctxt.GOPATH = os.Getenv("GOPATH")
	pkg, err := ctxt.Import(importPath, srcPath, build.AllowBinary)
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			return nil, fmt.Errorf("fail to get imports(%s): %v", importPath, err)
		}
		//log.Warn("Getting imports: %v", err)
	}
	rawImports := pkg.Imports
	numImports := len(rawImports)
	//Not Test
	//rawImports = append(rawImports, pkg.TestImports...)
	//numImports = len(rawImports)
	imports := make([]string, 0, numImports)
	for _, name := range rawImports {
		if IsGoRepoPath(name) {
			continue
		} else if strings.HasPrefix(name, rootPath) {
			moreImports, err := ListImports(name, srcPath, rootPath)
			if err != nil {
				return nil, err
			}
			imports = append(imports, moreImports...)
			continue
		}

		imports = append(imports, name)
	}
	return imports, nil
}
func IsGoRepoPath(importPath string) bool {
	return goRepoPath[importPath]
}

var goRepoPath = map[string]bool{}

func init() {
	for p := range standardPath {
		for {
			goRepoPath[p] = true
			i := strings.LastIndex(p, "/")
			if i < 0 {
				break
			}
			p = p[:i]
		}
	}
}

var standardPath = map[string]bool{
	"builtin": true,

	// go list -f '"{{.ImportPath}}": true,'  std
	"archive/tar":               true,
	"archive/zip":               true,
	"bufio":                     true,
	"bytes":                     true,
	"compress/bzip2":            true,
	"compress/flate":            true,
	"compress/gzip":             true,
	"compress/lzw":              true,
	"compress/zlib":             true,
	"container/heap":            true,
	"container/list":            true,
	"container/ring":            true,
	"context":                   true,
	"crypto":                    true,
	"crypto/aes":                true,
	"crypto/cipher":             true,
	"crypto/des":                true,
	"crypto/dsa":                true,
	"crypto/ecdsa":              true,
	"crypto/elliptic":           true,
	"crypto/hmac":               true,
	"crypto/md5":                true,
	"crypto/rand":               true,
	"crypto/rc4":                true,
	"crypto/rsa":                true,
	"crypto/sha1":               true,
	"crypto/sha256":             true,
	"crypto/sha512":             true,
	"crypto/subtle":             true,
	"crypto/tls":                true,
	"crypto/x509":               true,
	"crypto/x509/pkix":          true,
	"database/sql":              true,
	"database/sql/driver":       true,
	"debug/dwarf":               true,
	"debug/elf":                 true,
	"debug/gosym":               true,
	"debug/macho":               true,
	"debug/pe":                  true,
	"debug/plan9obj":            true,
	"encoding":                  true,
	"encoding/ascii85":          true,
	"encoding/asn1":             true,
	"encoding/base32":           true,
	"encoding/base64":           true,
	"encoding/binary":           true,
	"encoding/csv":              true,
	"encoding/gob":              true,
	"encoding/hex":              true,
	"encoding/json":             true,
	"encoding/pem":              true,
	"encoding/xml":              true,
	"errors":                    true,
	"expvar":                    true,
	"flag":                      true,
	"fmt":                       true,
	"go/ast":                    true,
	"go/build":                  true,
	"go/constant":               true,
	"go/doc":                    true,
	"go/format":                 true,
	"go/importer":               true,
	"go/internal/gccgoimporter": true,
	"go/internal/gcimporter":    true,
	"go/parser":                 true,
	"go/printer":                true,
	"go/scanner":                true,
	"go/token":                  true,
	"go/types":                  true,
	"hash":                      true,
	"hash/adler32":              true,
	"hash/crc32":                true,
	"hash/crc64":                true,
	"hash/fnv":                  true,
	"html":                      true,
	"html/template":             true,
	"image":                     true,
	"image/color":               true,
	"image/color/palette":       true,
	"image/draw":                true,
	"image/gif":                 true,
	"image/internal/imageutil":  true,
	"image/jpeg":                true,
	"image/png":                 true,
	"index/suffixarray":         true,
	"internal/race":             true,
	"internal/singleflight":     true,
	"internal/testenv":          true,
	"internal/trace":            true,
	"io":                        true,
	"io/ioutil":                 true,
	"log":                       true,
	"log/syslog":                true,
	"math":                      true,
	"math/big":                  true,
	"math/cmplx":                true,
	"math/rand":                 true,
	"mime":                      true,
	"mime/multipart":            true,
	"mime/quotedprintable":      true,
	"net":                     true,
	"net/http":                true,
	"net/http/cgi":            true,
	"net/http/cookiejar":      true,
	"net/http/fcgi":           true,
	"net/http/httptest":       true,
	"net/http/httputil":       true,
	"net/http/internal":       true,
	"net/http/pprof":          true,
	"net/internal/socktest":   true,
	"net/mail":                true,
	"net/rpc":                 true,
	"net/rpc/jsonrpc":         true,
	"net/smtp":                true,
	"net/textproto":           true,
	"net/url":                 true,
	"os":                      true,
	"os/exec":                 true,
	"os/signal":               true,
	"os/user":                 true,
	"path":                    true,
	"path/filepath":           true,
	"reflect":                 true,
	"regexp":                  true,
	"regexp/syntax":           true,
	"runtime":                 true,
	"runtime/cgo":             true,
	"runtime/debug":           true,
	"runtime/internal/atomic": true,
	"runtime/internal/sys":    true,
	"runtime/pprof":           true,
	"runtime/race":            true,
	"runtime/trace":           true,
	"sort":                    true,
	"strconv":                 true,
	"strings":                 true,
	"sync":                    true,
	"sync/atomic":             true,
	"syscall":                 true,
	"testing":                 true,
	"testing/iotest":          true,
	"testing/quick":           true,
	"text/scanner":            true,
	"text/tabwriter":          true,
	"text/template":           true,
	"text/template/parse":     true,
	"time":                    true,
	"unicode":                 true,
	"unicode/utf16":           true,
	"unicode/utf8":            true,
	"unsafe":                  true,
}

func ListPkgImports(pkg, pkgPath string) ([]string, error) {
	if strings.Contains(pkg, string(os.PathSeparator)) {
		pkg = strings.Replace(pkg, string(os.PathSeparator), "/", -1)
	}
	x, err := ListImports(pkg, pkgPath, pkg)
	if err != nil {
		return x, err
	}
	SortViaLen(x)
	x1 := make([]string, 0, len(x))
	for i := range x {
		if !strings.HasPrefix(x[i], pkg) {
			if !SearchHasPrefix(x1, x[i]) {
				x1 = append(x1, x[i])
			}
		}
	}
	return x1, nil
}
func SearchMatch(a []string, x string) bool {
	//sort.Strings(a)
	//return sort.Search(len(a), func(i int) bool { return a[i] == x })
	for _, v := range a {
		if v == x {
			return true
		}
	}
	return false
}
func SearchHasPrefix(a []string, x string) bool {
	for _, v := range a {
		switch {
		case len(v) < len(x):
			if strings.HasPrefix(x, v) {
				return true
			}
		default:
			if strings.HasPrefix(v, x) {
				return true
			}
		}

	}
	return false
}
func SortViaLen(a []string) {
	sort.Sort(ByLength(a))
}

type ByLength []string

func (s ByLength) Len() int {
	return len(s)
}
func (s ByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}
