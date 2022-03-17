/*
 * rpm2deb utility to convert Mariner rpm to deb format
 *
 */

package main

import(
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"github.com/DavidGamba/go-getoptions"
	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	"microsoft.com/pkggen/rpm2deb/debparser"
)

const version string = "0.1"

func validateInputFile(fpath string) error {
	logger.Println("Checking file", fpath)
	_, err := os.Stat(fpath)

	if err != nil {
		return err
	}

	ext := filepath.Ext(fpath)
	if ext != ".rpm" {
		return errors.New("Wrong File Format : " + ext)
	}
	return nil
}

func doCreateAndDoWriteIO(fname string, mod os.FileMode, rd io.Reader, pkg string) (int64, error) {
	var file *os.File
	var err error
	file, err = os.Create(fname)
	if err != nil {
		logger.Printf("Error: %v :%v file creation failed for %v\n", err, fname, pkg)
		return -1, err
	}
	defer file.Close()
	err = os.Chmod(fname, mod)
	if err != nil {
		logger.Printf("Error: %v :%v file chmod failed for %v\n", err, fname, pkg)
		return -1, err
	}

	/* Write */
	return io.Copy(file, rd)
}

func createRulesFile(pkg string) (int64, error) {
	content := "#!/usr/bin/make -f\n"
	content += "DH_VERBOSE = 1\nPACKAGE=\\$(shell dh_listpackages)\n"
	content += "build:\n\tdh_testdir\n"
	content += "clean:\n\tdh_testdir\n\tdh_testroot\n\tdh_clean -d\n"
	content += "binary-indep: build\n"
	content += "binary-arch: build\n\tdh_testdir\n\tdh_testroot\n\tdh_prep\n\tdh_installdirs\n\tdh_installdocs\n\tdh_installchangelogs\n\t"
	content += "find . -maxdepth 1 -mindepth 1 -not -name debian -print0 | xargs -0 -r -i cp -a {} debian/$(PACKAGE)\n\t"
	content += "dh_compress\n\tdh_makeshlibs\n\tdh_installdeb\n\t#dh_shlibdeps\n\tdh_gencontrol\n\tdh_md5sums\n\tdh_builddeb\n"
	content += "binary: binary-indep binary-arch\n"
	content += ".PHONY: build clean binary-indep binary-arch binary"

	return doCreateAndDoWriteIO("rules", 0755, strings.NewReader(content), pkg)
}

func getPackageName(s []string) (int, string) {
	const alpha = "abcdefghijklmnopqrstuvwxyz"
	var output string
	var ret int
	for idx, c := range s {
		if len(c) == 0 || !strings.ContainsAny(strings.ToLower(c), alpha) || strings.Contains(c, "_") {
			ret = idx
			break
		}
		if (idx != 0) {
			output += "-"
		}
		output += c
	}
	logger.Printf("pkgName: %v, idx: %v\n", output, ret)
	return ret, output
}

func createControlFile(pkgName string, pkg debparser.Package) (int64, error) {
	content := "Source: rtd\n"
	content += "Section: admin\n"
	content += "Priority: optional\n"
	content += "Build-Depends: debhelper (>= 7.0.50)\n"
	content += "Maintainer: Praveen Kumar <kumarpraveen@microsoft.com>\n"
	content += "Standards-Version: 1.0.0\n"
	content += "#Vcs-Git:\n"
	content += "#Homepage:\n"
	content += "\n"		/* New line is required for Next Stanza */
	_, pkgname := getPackageName(strings.Split(pkgName, "-"))
	content += "Package: " + strings.ToLower(pkgname) +"\n"
	content += "Architecture: all\n"
//	content += "Depends: debhelper (>= 7), ${misc:Depends}, ${go:Depends}, rpm (>= 2.4.4-2), dpkg-dev, make, cpio, rpm2cpio\n"
	content += "Depends: " +  strings.Join(pkg.Depends, ", ") + "\n"
	content += "Suggests: patch, bzip2, lsb-rpm, lintian, lzma\n"
	content += "Description: Converts Mariner's RPM package to Debian package. This tool only works upon existing binaries and there is NO build performed\n"

	return doCreateAndDoWriteIO("control", 0644, strings.NewReader(content), pkgName)
}

func createCompatFile(pkg string) (int64, error) {
	content := "9\n"
	return doCreateAndDoWriteIO("compat", 0644, strings.NewReader(content), pkg)
}

func createChangelogFile(pkg string) (int64, error) {
	content := "rtd ("+ version + ") experimental; urgency=low\n\n"
	content += "  * Initial Release.\n\n"
	content += "  -- Praveen Kumar <kumarpraveen@microsoft.com>  Thu, 22 Nov 2021 13:21:48 -0800\n"

	return doCreateAndDoWriteIO("changelog", 0644, strings.NewReader(content), pkg)
}

func createReleaseFile(pkg string, repoName string) (int64, error) {
	content := "Archive: stable\n"
	content += "Component: main\n"
	content += "Origin: Micorosft\n"
	content += "Label:" + repoName + "\n"
	content += "Architecture: all"

	return doCreateAndDoWriteIO("release", 0755, strings.NewReader(content), pkg)
}

func createDebPkgMandatoryFiles(pkgName string, pkg debparser.Package) error {
	_, err := createRulesFile(pkgName)
	if err != nil {
		return err
	}

	_, err = createControlFile(pkgName, pkg)
	if err != nil {
		return err
	}

	_, err = createChangelogFile(pkgName)
	if err != nil {
		return err
	}

	_, err = createCompatFile(pkgName)
	if err != nil {
		return err
	}
	return nil
}

func getDebFilesGenerated(pkgName string, pkg debparser.Package) error {
	err := os.Mkdir("debian", 0755)
	if err != nil {
		logger.Println(err)
		return err
	}

	curDir, _ := os.Getwd()
	err = os.Chdir("debian")
	if err != nil {
		logger.Println(err)
		return err
	}

	err = createDebPkgMandatoryFiles(pkgName, pkg)
	if err != nil {
		logger.Println(err)
		return err
	}

	err = os.Chdir(curDir)
	if err != nil {
		logger.Println(err)
		return err
	}
	return nil
}

func CreateAndExtractRPM(fdir string, fname string) (string, error) {

	if len(fname) <=4 {
		logger.Println("Invalid filename provided")
		return "", errors.New("Invalid filename provided")
	}

	err := os.Chdir(fdir);
	if err != nil {
		logger.Println(err)
		return "", err
	}

	/* filename to be .rpm format name */
	pkgName := fname[:len(fname)-4]
	err = os.Mkdir(pkgName, 0755)
	if err != nil {
		logger.Println(err)
		return "", err
	}

	err = os.Chdir(pkgName);
	if err != nil {
		logger.Println(err)
		return "", err
	}

	rpm2cpioCmd := exec.Command("rpm2cpio", "../"+fname)
	cpioCmd := exec.Command("cpio", "--extract", "--make-directories", "--no-absolute-filenames", "--preserve-modification-time")
	rd, wr := io.Pipe()
	rpm2cpioCmd.Stdout = wr
	cpioCmd.Stdin = rd

	rpm2cpioCmd.Start()
	cpioCmd.Start()

	rpm2cpioCmd.Wait()
	wr.Close()

	cpioCmd.Wait()
	rd.Close()

	return pkgName, nil
}

func createRepo(pkgLocation string, repoName string) error {
	/*
	 * Check if repo already exists. In that case add the packages from
	 * package location to the repo 
	 */
	output, err := exec.Command("aptly", "repo", "show", repoName).Output()
	if err != nil {
		logger.Println(err)
	}

	if string(output) != "" {
		/* repo exists, add packages from dir */
		err = addPackageToRepo(pkgLocation, repoName)

		/* update the repo with changed packages */
		if err != nil {
			cmd := exec.Command("aptly", "publish", "update", "--skip-signing", repoName)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Start()
			if err != nil {
				logger.Println(err)
			}
			cmd.Wait()
		}

	} else {
		/* Create a repo with distribution set as test and name as mentioned */

		/* Before creating the repo make sure it is created in the current directory */
		/* edit aptly conf file for the same */
		output, err := exec.Command("pwd").Output()
		if err != nil {
			logger.Println(err)
		}
		output_string := string(output)
		output_string = strings.TrimSuffix(output_string, "\n")

		err = changeRepoDir(string(output_string))
		if err != nil {
			logger.Println(err)
		}

		cmd := exec.Command("aptly", "repo", "create", "-distribution=test", "-component=main", repoName)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Start()
		if err != nil {
			logger.Println(err)
		}
		cmd.Wait()

		err = addPackageToRepo(pkgLocation, repoName)

		if err != nil {
			/* if packages added correctly, publish the changes locally */
			cmd := exec.Command("aptly", "publish", "repo", "-architectures=\"amd64\"",  "-skip-signing", repoName)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Start()
			if err != nil {
				logger.Println(err)
			}
			cmd.Wait()
		}
	}

	return err
}
func createRepoManually(pkgLocation string, repoName string) error {
	/*
	 * Move to the pkg directory
	 */
	err := os.Chdir(pkgLocation);
	if err != nil {
		logger.Println(err)
		return err
	}

	/*
	 * create the required directories
	 */
	cmd := exec.Command("bash", "-c", `mkdir dists; mkdir dists/stable; mkdir dists/stable/main`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()

	if err != nil {
		logger.Println(err)
		return err
	}
	//move all packages to main
	cmd = exec.Command("bash", "-c", `mv *.deb dists/stable/main`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logger.Println(err)
		return err
	}

	err = os.Chdir("dists/stable/main")
	if err != nil {
		return err
	}

	err = os.Mkdir("binary", 0755)
	if err != nil {
		logger.Println(err)
		return err
	}

	err = os.Mkdir("source", 0755)
	if err != nil {
		logger.Println(err)
		return err
	}

	// Generate source and package files
	scanPackCmd := exec.Command("bash", "-c", `dpkg-scanpackages --multiversion . | gzip -9c > binary/Packages.gz`)
	scanPackCmd.Stderr = os.Stderr
	scanPackCmd.Stdout = os.Stdout
	err = scanPackCmd.Start()
	if err != nil {
		logger.Println(err)
		return err
	}
	scanPackCmd.Wait()

	scanSourceCmd := exec.Command("bash", "-c", `dpkg-scansources . | gzip -9c > source/Sources.gz`)
	scanSourceCmd.Stderr = os.Stderr
	scanSourceCmd.Stdout = os.Stdout
	err = scanSourceCmd.Start()
	if err != nil {
		logger.Println(err)
		return err
	}

	/*
	 * Create release files in binary and source directories
	 */

	_, err = createReleaseFile("repo creation", repoName)
	if err != nil {
		return err
	}

	cmd = exec.Command("cp", "release", "binary")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logger.Println(err)
		return err
	}

	cmd = exec.Command("mv", "release", "binary")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logger.Println(err)
		return err
	}
	return nil
}

func addPackageToRepo(pkgLocation string, repoName string) error {
	cmd := exec.Command("aptly", "repo", "add", repoName, pkgLocation)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		logger.Println(err)
	}
	cmd.Wait()

	return err
}

func changeRepoDir(repoLocation string) error {
	search_str := "\"rootDir\": \"/root/.aptly\""
	replace_str :=  "\"rootDir\": \"" + repoLocation + "\""
	cmd := exec.Command("sed", "-i", "s%"+search_str+"%"+ replace_str+"%", "/root/.aptly.conf")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		logger.Println(err)
	}

	cmd.Wait()

	return err
}

func generateDebPackage() error {
	cmd := exec.Command("./debian/rules", "binary")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		logger.Println(err)
		return err
	}
	cmd.Wait()

	return nil
}

func renameDebPackageName(pkg string, toLocation string) error {
	tokens := strings.Split(pkg, "-")
	idx, pkgName := getPackageName(strings.Split(pkg, "-"))
	newVersion := strings.Join(tokens[idx:], "-")

	oldName := "../"+strings.ToLower(pkgName) + "_" + version + "_all.deb"
	var newName string
	if toLocation == "" {
		newName = "../"+ pkgName + "_" + newVersion + ".deb"
	} else {
		newName = toLocation + "/" + pkgName + "_" + newVersion + ".deb"
	}
	logger.Printf("oldName : %v\n", oldName)
	logger.Printf("newName : %v\n", newName)

	cmd := exec.Command("mv", oldName, newName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		logger.Println(err)
	}
	cmd.Wait()
	return err
}

func cleanUp(fdir string, pkgName string) error {
	err := os.Chdir(fdir)
	if err != nil {
		logger.Println("Error changing present working dir to", fdir, err)
		return err
	}

	cmd := exec.Command("rm", "-fR", pkgName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	if err != nil {
		logger.Println(err)
	}
	cmd.Wait()
	return err
}

func CreateAndExtractRPMFromURL(url string, fdir string, fname string) (string, error) {

	if len(fname) <=4 {
		logger.Println("Invalid filename provided")
		return "", errors.New("Invalid filename provided")
	}

	err := os.Chdir(fdir);
	if err != nil {
		return "", err
	}

	pkgName := fname[:len(fname)-4]

	err = os.Mkdir(pkgName, 0755)
	if err != nil {
		return "", err
	}

	err = os.Chdir(pkgName);
	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)

	pkg, err := rpm.Read(br)
	if err != nil {
		return "", err
	}

	if compression := pkg.PayloadCompression(); compression != "gzip" {
		return "", errors.New("Unsupported compression: " + compression)
	}

	gzipReader, err := gzip.NewReader(br)
	if err != nil {
		return "", err
	}

	if format := pkg.PayloadFormat(); format != "cpio" {
		return "", errors.New("Unsupported payload format: "+ format)
	}

	cpioReader := cpio.NewReader(gzipReader)
	for {
		hdr, err := cpioReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if !hdr.Mode.IsRegular() {
			continue
		}

		if dirName := filepath.Dir(hdr.Name); dirName != "" {
			if err := os.MkdirAll(dirName, 0o755); err != nil {
				return "", err
			}
		}

		outFile, err := os.Create(hdr.Name)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(outFile, cpioReader); err != nil {
			outFile.Close()
			return "", err
		}
		outFile.Close()
	}
	return pkgName, nil
}

func convertSingleRPMFromURL(url string, toLocation string) error {

	 // TODO
	var pkg debparser.Package

	tokens := strings.Split(url, "/")
	fname := tokens[len(tokens)-1]
	fdir := toLocation

	xerr := debparser.ParseRPMXmlFileFromURL(url, fdir, fname, &pkg)
	if xerr != nil {
		return xerr
	}

	logger.Printf("Fdir %v Fname %v\n", fdir, fname)

	pkgName, err := CreateAndExtractRPMFromURL(url, fdir, fname)
	if err != nil {
		logger.Printf("Error: %v Failed to Create and Extract the provided file %s\n", err, url)
		goto clean
	}

	/*
	 * Remember at this stage. present working directory is fdir/pkgName.
	 * Lets perform debian activity
	 */
	err = getDebFilesGenerated(pkgName, pkg)
	if err != nil {
		logger.Println("Error: Generating the Debian mandatory files for", fname)
		goto clean
	}

	err = generateDebPackage()
	if err != nil {
		logger.Println("Error: Generating the Debian package for", fname)
		goto clean
	}

	renameDebPackageName(pkgName, toLocation)
clean:
	cleanUp(fdir, pkgName)

	return err
}

func convertSingleRPM(fpath string, toLocation string) error {
	err := validateInputFile(fpath)
	if err != nil {
		return err
	}

	var pkg debparser.Package
	err = debparser.ParseRPMXmlFile(fpath, &pkg)
	if err != nil {
		return err
	}
	logger.Printf("Depends : %v\n", pkg.Depends)
	tokens := strings.Split(fpath, "/")
	fname := tokens[len(tokens)-1]
	fdir := strings.Join(tokens[:len(tokens)-1], "/")
	logger.Printf("Fdir %v Fname %v\n", fdir, fname)

	pkgName, err := CreateAndExtractRPM(fdir, fname)
	if err != nil {
		logger.Println("Error: Failed to Create and Extract the provided file", fpath)
		goto clean
	}

	/*
	 * Remember at this stage. present working directory is fdir/pkgName.
	 * Lets perform debian activity
	 */
	err = getDebFilesGenerated(pkgName, pkg)
	if err != nil {
		logger.Println("Error: Generating the Debian mandatory files for", fname)
		goto clean
	}

	err = generateDebPackage()
	if err != nil {
		logger.Println("Error: Generating the Debian package for", fname)
		goto clean
	}

	renameDebPackageName(pkgName, toLocation)
clean:
	cleanUp(fdir, pkgName)

	return err
}

func getLinks(fromURL string, body io.Reader) []string {
	var links []string
	token := html.NewTokenizer(body)
	for {
		tag := token.Next()
		switch tag {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			mytoken := token.Token()
			if "a" == mytoken.Data {
				for _, attr := range mytoken.Attr {
					if attr.Key == "href" && strings.Contains(attr.Val, ".rpm") {
						url := fromURL + attr.Val
						links = append(links, url)
					}
				}
			}
		}
	}
	return links
}

func createAndValidateToLocation(location string) error {
	dir, err := os.Stat(location)
	if err != nil {
		return err
	}
	if !dir.IsDir() {
		return errors.New("Not a directory")
	}
	return nil
}

func doWork(toLoc string, ch chan string, quit chan bool) {
	for {
		select {
		case url := <-ch:
			convertSingleRPMFromURL(url, toLoc)
		case <-quit:
			logger.Println("Done with the processing")
			return
		default:
			time.After(1 * time.Second)
		}
	}
}

func worker(ch chan string, lists []string) {
	for _, v := range lists {
		ch <- v
	}
}

func handleMultipleScenarioRequest(fromURL string, toLocation string) error {

	resp, err := http.Get(fromURL)
	if err != nil {
		logger.Fatal(err)
		return err
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)

	links := getLinks(fromURL, br)

	err = createAndValidateToLocation(toLocation)
	if err != nil {
		logger.Fatal(err)
		return err
	}

	logger.Println("Total number of rpms", len(links))
	ch := make(chan string)
	quit := make(chan bool)

	var wg sync.WaitGroup
	go doWork(toLocation, ch, quit)
	wg.Add(1)

	for _, v := range links {
		logger.Println(v)
		ch <- v
	}
	close(ch)
	quit <- true
	close(quit)
	wg.Wait()

	return nil
}

//TODO Check required software
func checkSystemRequiredAndAccess() {
	currUser, err := user.Current()
	if err != nil {
		logger.Fatal(err)
	}

	if currUser.Username != "root" {
		logger.Fatal("Permission Denied: Sudo priviledge required!!\n")
	}
}

var logger = log.New(ioutil.Discard, "DEBUG: ", log.LstdFlags)

func main() {
	var debug bool
	var manualRepo bool
	var filePath string
	var fromURL string
	var toLocation string
	var pkgLocation string
	var repoName string

	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("h", "?"))
	opt.BoolVar(&debug, "debug", false)
	opt.BoolVar(&manualRepo, "manual-repo", false, opt.Alias("m"),
		opt.Description("Direction if the repository is to be created manually"))
	opt.StringVar(&filePath, "file", "", opt.Alias("f"),
		opt.Description("RPM file path to be converted to DEB format"))
	opt.StringVar(&fromURL, "from-url", "", opt.Alias("u"),
		opt.Description("RPM repo URL location from where the RPM packages will be downloaded and converted to DEB packages"))
	opt.StringVar(&toLocation, "to-loc", "", opt.Alias("t"),
		opt.Description("Directory Location where to put the converted DEB packages"))
	opt.StringVar(&pkgLocation, "pkg-loc", "", opt.Alias("p"),
		opt.Description("Directory Location from where to pick up packages for repo creation"))
	opt.StringVar(&repoName, "repo-name", "", opt.Alias("r"),
		opt.Description("Name of the repository that is to be created or to which packages need to be added"))

	remaining, err := opt.Parse(os.Args[1:])

	if opt.Called("help") {
		fmt.Fprintf(os.Stderr, opt.Help())
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		fmt.Fprintf(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		os.Exit(1)
	}

	if filePath == "" &&  fromURL == ""  && pkgLocation == "" {
		fmt.Fprintf(os.Stderr, "ERROR: No input provided\n\n")
		fmt.Fprintf(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		os.Exit(1)
	} else if filePath != "" && fromURL != "" {
		fmt.Fprintf(os.Stderr, "ERROR: Either file or url can be provided but not both\n\n")
		fmt.Fprintf(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		os.Exit(1)
	}
	if fromURL != "" && toLocation == "" {
		fmt.Fprintf(os.Stderr, "ERROR: to-loc not provided\n\n")
		fmt.Fprintf(os.Stderr, opt.Help(getoptions.HelpSynopsis))
		os.Exit(1)
	}

	if debug {
		logger.SetOutput(os.Stderr)
	}
	logger.Printf("Unhandled CLI args: %v\n", remaining)

	checkSystemRequiredAndAccess()

	if filePath != "" {
		err = convertSingleRPM(filePath, toLocation)
		if err != nil {
			logger.Fatalf("ERROR: %v: Failed converting %s\n\n", err, filePath)
		}
	} else if fromURL != "" && toLocation != "" {
		err = handleMultipleScenarioRequest(fromURL, toLocation)
		if err != nil {
			logger.Fatalf("ERROR: %v: Processing Failed for url %s\n\n", err, fromURL)
		}

	} else if pkgLocation != "" && repoName!= "" {
		if manualRepo {
			err = createRepoManually(pkgLocation, repoName)
		} else {
			err = createRepo(pkgLocation, repoName)
		}
	}
}
