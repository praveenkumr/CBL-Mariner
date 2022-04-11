package debparser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

type RpmHeader struct {
	XMLName xml.Name `xml:"rpmHeader"`
	Text    string   `xml:",chardata"`
	RpmTag  []struct {
		Text    string   `xml:",chardata"`
		Name    string   `xml:"name,attr"`
		String  []string `xml:"string"`
		Integer []string `xml:"integer"`
		Base64  string   `xml:"base64"`
	} `xml:"rpmTag"`
}

type Package struct {
	Name string		// rpmTag name="Name"
	Version string		// rpmTag name="Version"
	Release string		// rpmTag name="Release"
	Summary string		// rpmTag name="Summary"
	Description string	// rpmTag name="Description"
	Distribution string	// rpmTag name="Distribution"
	Vendor string		// rpmTag name="Vendor"
	License string		// rpmTag name="License"
	Group string		// rpmTag name="Group"
	Url string		// rpmTag name="Url"
	Os string		// rpmTag name="Os"
	Architecture string	// rpmTag name="Arch"
	SourceRpm string	// rpmTag name="Sourcerpm"
	Changelog string	// Taken from rpm -q --changelog <rpm>
	Compat string		// constant for now
	Depends []string	// rpmTag name="Requirename"
	Provides []string	// rpmTag name="Providename"
	Conflict []string
	PostIn []string		// rpmTag name="Postin"
	PostInProg string	// rpmTag name="Postinprog"
	PostUn []string		// rpmTag name="Postun"
	PostUnProg string	// rpmTag name="Postunprog"
	ObsoleteName string	// rpmTag name="Obsoletename"
	Cookie string		// rpmTag name="Cookie"
	Platform string		// rpmTag name = "Platform"
}

/*
  <rpmTag name="Postinprog">
	<string>&lt;lua&gt;</string>		// <lua>
  </rpmTag>

  <rpmTag name="Postinprog">
  	<string>/bin/sh</string>
  </rpmTag>
*/
func ParseProgram(prog string) string {
	r := strings.NewReplacer("<", "", ">", "")
	return r.Replace(prog)
}

func validatePackageName(name string) bool {
	// TODO combine regex
	soMatched, _ := regexp.MatchString(`\.so`, name)
	slashMatched, _ := regexp.MatchString(`\/`, name)
	rpmlibMatched, _ := regexp.MatchString(`rpmlib`, name)
	//bracketMatched, _ := regexp.MatchString(`\(`, name)
	//if soMatched || slashMatched || rpmlibMatched || bracketMatched  {
	if soMatched || slashMatched || rpmlibMatched {
		return false
	}
	return true
}

func ParseRPMXmlFileFromURL(url string, fdir string, fname string, rpmpkg *Package) error {
	_, err := exec.Command("wget", url, "-P", fdir).Output()
	if err != nil {
		log.Println(err)
		return err
	}
	fpath := fdir + "/" + fname
	err = ParseRPMXmlFile(fpath, rpmpkg)
	if err != nil {
		log.Printf("Error Parsing RPM file %v : Error : %v\n", fpath, err)
	}

	_, err = exec.Command("rm", "-f", fpath).Output()
	if err != nil {
		log.Println(err)
	}
	return err
}

func ProcessDebianSpecificName(name string) string {
	// debian build system requires pkg name in lower case and no special symbol "(?^:^(?^:[a-z0-9][-+\.a-z0-9]+)$)"
	name = strings.ToLower(name)
	// if we have /bin/grub lets send back grub
	tokens := strings.Split(name, "/")
	name = tokens[len(tokens)-1]

	var rname string
	for i, ch := range name {
		pos := len(rname)
		if ch == '(' || ch == ')' || ch == '_' || ch == ':' || ch == '=' {
			if (pos != 0 && rname[pos-1] != '-') && (i != len(name)-1 && name[i+1] != ' ') {
				rname += "-"
			} else {
				continue
			}
		} else {
			rname += string(name[i])
		}
	}
	// last characters cannot be any special character
	var idx int
	for idx = len(rname)-1; idx >=0; idx-- {
		if rname[idx] == '-' {
			continue
		}
		break
	}
	return rname[:idx+1]
}

func ParseRPMXmlFile(fpath string, rpmpkg *Package) error {
	cmd := exec.Command("rpm", "-q", "--xml", fpath)

	out, err := cmd.Output()
	if err != nil {
		// TODO rpm -q --xml give exit code 1 . Don't return
		fmt.Println(err)
	}

	byteValue, _ := ioutil.ReadAll(bytes.NewReader(out))

	var myxml RpmHeader
	xml.Unmarshal(byteValue, &myxml)
	for i:= 0; i < len(myxml.RpmTag); i++ {
		switch myxml.RpmTag[i].Name {
		case "Name":
			rpmpkg.Name = ProcessDebianSpecificName(myxml.RpmTag[i].String[0])
		case "Version":
			rpmpkg.Version = myxml.RpmTag[i].String[0]
		case "Release":
			rpmpkg.Release = myxml.RpmTag[i].String[0]
		case "Summary":
			rpmpkg.Summary = myxml.RpmTag[i].String[0]
		case "Description":
			rpmpkg.Description = myxml.RpmTag[i].String[0]
		case "Distribution":
			rpmpkg.Distribution = myxml.RpmTag[i].String[0]
		case "Vendor":
			rpmpkg.Vendor = myxml.RpmTag[i].String[0]
		case "License":
			rpmpkg.License = myxml.RpmTag[i].String[0]
		case "Group":
			rpmpkg.Group = myxml.RpmTag[i].String[0]
		case "Url":
			rpmpkg.Url = myxml.RpmTag[i].String[0]
		case "Os":
			rpmpkg.Os = myxml.RpmTag[i].String[0]
		case "Arch":
			rpmpkg.Architecture = myxml.RpmTag[i].String[0]
		case "Sourcerpm":
			rpmpkg.SourceRpm = myxml.RpmTag[i].String[0]
		case "Postin":
			for j := 0; j < len(myxml.RpmTag[i].String); j++ {
				rpmpkg.PostIn = append(rpmpkg.PostIn, myxml.RpmTag[i].String[j])
			}
		case "Postinprog":
			rpmpkg.PostInProg = ParseProgram(myxml.RpmTag[i].String[0])
		case "Postun":
			for j := 0; j < len(myxml.RpmTag[i].String); j++ {
				rpmpkg.PostUn = append(rpmpkg.PostUn, myxml.RpmTag[i].String[j])
			}
		case "Postunprog":
			rpmpkg.PostUnProg = ParseProgram(myxml.RpmTag[i].String[0])
		case "Obsoletename":
			rpmpkg.ObsoleteName = myxml.RpmTag[i].String[0]
		case "Cookie":
			rpmpkg.Cookie = myxml.RpmTag[i].String[0]
		case "Platform":
			rpmpkg.Platform = myxml.RpmTag[i].String[0]

		case "Requirename":
			fmt.Println(myxml.RpmTag[i].Name)
			requiresMap := make(map[string]bool)
			for j := 0; j < len(myxml.RpmTag[i].String); j++ {
				matched := validatePackageName(myxml.RpmTag[i].String[j])
				if matched {
					_, ok := requiresMap[myxml.RpmTag[i].String[j]]
					if !ok {
						requiresMap[myxml.RpmTag[i].String[j]] = true
						fmt.Println(myxml.RpmTag[i].String[j])
						rpmpkg.Depends = append(rpmpkg.Depends, ProcessDebianSpecificName(myxml.RpmTag[i].String[j]))
					}
				}
			}
		case "Providename":
			for j := 0; j < len(myxml.RpmTag[i].String); j++ {
				matched := validatePackageName(myxml.RpmTag[i].String[j])
				if matched {
					rpmpkg.Provides = append(rpmpkg.Provides, ProcessDebianSpecificName(myxml.RpmTag[i].String[j]))
				}
			}
		}
		// TODO for Conflict
	}
	out, err = exec.Command("rpm", "-q", "--changelog", fpath).Output()
	if err != nil {
		fmt.Println(err)
	}
	rpmpkg.Changelog = string(out)
//	log.Printf("Name %v,\n", rpmpkg.Name)
//	log.Printf("Depends %v,\n", rpmpkg.Depends)
//	log.Printf("Changelog %v,\n", rpmpkg.Changelog)
	return nil
}
