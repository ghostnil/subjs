package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/C0RB3N/subjs/banner"
	"github.com/PuerkitoBio/goquery"
)

func errCheck(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func validateURLSchema(domain string) string {
	if !strings.Contains(domain, "http://") && !strings.Contains(domain, "https://") {
		fmt.Println("[+] Provide right schema, all requests will be https://", domain)
		newURL := "https://" + domain
		return newURL
	}
	return ""
}

func main() {

	var (
		outf         string
		domains      []string
		singleDomain string
	)

	singleDomainOut := make(map[string][]string)
	out := make(map[string][]string)

	//menu
	fmt.Println(banner.Banner())
	flag.StringVar(&outf, "o", "", "Name of the output file")
	flag.StringVar(&singleDomain, "d", "", "Name of the uniq domain to search for")
	flag.Parse()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var subjs = &http.Client{
		Timeout: time.Second * 5,
	}

	re := regexp.MustCompile("^(?:https?:\\/\\/)?(?:www\\.)?([^\\/]+)")

	// scan single domain
	if outf != "" && singleDomain != "" {

		//check domain
		newURL := validateURLSchema(singleDomain)
		if newURL != "" {
			singleDomain = newURL
		}

		resp, err := subjs.Get(singleDomain)
		errCheck(err)

		host := re.FindStringSubmatch(singleDomain)
		if err == nil {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			errCheck(err)

			doc.Find("script").Each(func(index int, s *goquery.Selection) {
				js, _ := s.Attr("src")
				if js != "" {
					if strings.HasPrefix(js, "http://") || strings.HasPrefix(js, "https://") || strings.HasPrefix(js, "//") {
						singleDomainOut[singleDomain] = append(singleDomainOut[singleDomain], js)
					} else {
						js := strings.Join([]string{host[1], js}, "")
						singleDomainOut[singleDomain] = append(singleDomainOut[singleDomain], js)
					}
				}
			})

			if len(singleDomainOut) != 0 {
				bytes, err := json.MarshalIndent(singleDomainOut, "", "    ")
				errCheck(err)
				if err == nil {
					fmt.Println(string(bytes))
				}
				if outf != "" {
					ioutil.WriteFile(outf, bytes, 0644)
				}
			}
			fmt.Println("[+] Operation sucess ouput in: ", outf)
		}
	} else {

		// validation open file
		m, _ := os.Stdin.Stat()
		if (m.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				domains = append(domains, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "-> subjs - corben leo\n-> usage: cat urls.txt | subjs", err)
				os.Exit(3)
			}
		} else {
			fmt.Fprintf(os.Stderr, "-> subjs - corben leo \n-> usage: cat urls.txt | subjs")
		}

		// send urls from file to http handler
		for _, domain := range domains {

			newURL := validateURLSchema(domain)
			if newURL != "" {
				domain = newURL
			}

			resp, err := subjs.Get(domain)
			errCheck(err)

			host := re.FindStringSubmatch(domain)
			errCheck(err)

			if err == nil {
				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					fmt.Println("Error parsing response from: ", domain)
				}

				doc.Find("script").Each(func(index int, s *goquery.Selection) {
					js, _ := s.Attr("src")
					if js != "" {
						if strings.HasPrefix(js, "http://") || strings.HasPrefix(js, "https://") || strings.HasPrefix(js, "//") {
							out[domain] = append(out[domain], js)
						} else {
							js := strings.Join([]string{host[1], js}, "")
							out[domain] = append(out[domain], js)
						}
					}
				})
			}
			// creation file output
			if len(out) != 0 {
				bytes, err := json.MarshalIndent(out, "", "    ")
				errCheck(err)
				if err == nil {
					fmt.Println(string(bytes))
				}
				if outf != "" {
					ioutil.WriteFile(outf, bytes, 0644)
				}
			}
			fmt.Println("[+] Operation sucess ouput in: ", outf)
		}
	}

} //end main
