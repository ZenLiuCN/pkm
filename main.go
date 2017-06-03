package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	cSrc      string
	cTar      string
	cPkg      string
	cGit      bool
	cOver     bool
	cMode     string
	RunDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	z7a       = filepath.Join(RunDir, "7z")
	LogPath   = "pkm.log"
)

func main() {
	flag.StringVar(&cMode, "M", "P", "[P]ackMode,[I]nstallMode,[L]istMode\n        DEFAULT: P(PackMode)")
	flag.StringVar(&cSrc, "S", "", "InstallMode: path of where store package.7z \n        PackMode: where package src exists(EG: %GOPATH%\\src)\n        ListMode: package src path (EG: %GOPATH%\\src)")
	flag.StringVar(&cTar, "T", "", "InstallMode: path of package to install \n        PackMode: path of where to store package.7z \n        DEFAULT: same as -S")
	flag.StringVar(&cPkg, "P", "", "InstallMode and ListMode ONLY: Package full name ")
	flag.BoolVar(&cGit, "C", false, "PackMode ONLY: Remove .git folder \n        DEFAULT: false")
	flag.BoolVar(&cOver, "J", false, "PackMode and ListMode ONLY:\n        PackMode:Overwrite Package.7z\n        ListMode: List sub package depends \n        DEFAULT: false")
	flag.Parse()
	switch {
	case len(cSrc) == 0:
		flag.Usage()
		return
	case cMode != "P" && cMode != "I" && cMode != "L":
		fmt.Println("Must Define -M P|I|L (default P)")
		flag.Usage()
		return
	case len(cSrc) == 0:
		fmt.Println("Must Define -S")
		flag.Usage()
		return
	case cMode == "I" && len(cPkg) == 0:
		fmt.Println("Must Define -P PACKAGENAME when use InstallMode")
		flag.Usage()
		return
	case cMode == "L" && len(cPkg) == 0:
		fmt.Println("Must Define -P PACKAGENAME when use ListMode")
		flag.Usage()
		return
	}
	if len(cTar) == 0 {
		cTar = cSrc
	}
	if FileExist(LogPath) {
		os.Remove(LogPath)
	}
	Logf, er := os.Create(LogPath)
	if er != nil {
		fmt.Println("Create logfile error:", er)
		os.Exit(1)
	}
	Logf.Close()
	switch cMode {
	case "I":
		InstallPackage(cSrc, cTar, cPkg)
	case "P":
		PackPackage(cSrc, cTar, cGit, cOver)
	case "L":
		DepCheck(cPkg, cSrc, cOver)
	}
}
