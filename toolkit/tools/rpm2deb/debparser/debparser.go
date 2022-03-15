package debparser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
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
	Name string
	Location string
	Changelog string
	Compat string
	Depends []string
	Conflict string
}

const version string = "0.1"

func validatePackageName(name string) bool {
	// TODO combine regex
	soMatched, _ := regexp.MatchString(`\.so`, name)
	slashMatched, _ := regexp.MatchString(`\/`, name)
	rpmlibMatched, _ := regexp.MatchString(`rpmlib`, name)
	bracketMatched, _ := regexp.MatchString(`\(`, name)
	if soMatched || slashMatched || rpmlibMatched || bracketMatched  {
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

func ParseRPMXmlFile(fpath string, rpmpkg *Package) error {
	cmd := exec.Command("rpm", "-q", "--xml", fpath)

	out, err := cmd.Output()
	if err != nil {
		// TODO rpm -q --xml give exit code 1 . Don't return
		fmt.Println(err)
	}

	byteValue, _ := ioutil.ReadAll(bytes.NewReader(out))

	var myxml RpmHeader
	requiresMap := make(map[string]bool)
	xml.Unmarshal(byteValue, &myxml)
	for i:= 0; i < len(myxml.RpmTag); i++ {
		if myxml.RpmTag[i].Name == "Requirename" {
			fmt.Println(myxml.RpmTag[i].Name)
			for j := 0; j < len(myxml.RpmTag[i].String); j++ {
				matched := validatePackageName(myxml.RpmTag[i].String[j])
				if matched {
					_, ok := requiresMap[myxml.RpmTag[i].String[j]]
					if !ok {
						requiresMap[myxml.RpmTag[i].String[j]] = true
						fmt.Println(myxml.RpmTag[i].String[j])
						rpmpkg.Depends = append(rpmpkg.Depends, myxml.RpmTag[i].String[j])
					}
				}
			}
			break
		}
		// TODO for Conflict and other attributes
	}
	/*
	cmd = exec.Command("rpm", "-q", "--changelog", fpath)

	out, err := cmd.Output()
	*/
	log.Println(rpmpkg.Depends)
	return nil
}
