package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

//CHANGE ME
const pat = "xxxxx" //Your DigitalOcean API Key
const user = "root"

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func DropletList(client *godo.Client) ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		for _, d := range droplets {
			list = append(list, d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

func OutputList(output []godo.Droplet) string {
	if len(output) == 0 {
		fmt.Println("You have no droplets, log in and create one.")
	}

	var v4address string

	fmt.Printf("DigitalOcean Servers List:\n\n")
	fmt.Printf("    Hostname \t\t IP Address \t\t\t Region\n")

	// make sure address is not RFC1918 space if private networking is enabled
	for i := 0; i < len(output); i++ {
		if output[i].Networks.V4[0].Type == "private" {
			v4address = output[i].Networks.V4[1].IPAddress
		} else {
			v4address = output[i].Networks.V4[0].IPAddress
		}

		fmt.Printf("[%v] %v\t\t %v\t\t %v\t\t\n", i+1, output[i].Name, v4address, output[i].Region.Name)
	}

	var servernum int

	fmt.Println("")

	for servernum > len(output) || servernum <= 0 {
		fmt.Printf("Which Droplet number would you like to connect to? ")
		fmt.Scan(&servernum)
		fmt.Println("")
	}

	var ip string

	if output[servernum-1].Networks.V4[0].Type == "private" {
		ip = output[servernum-1].Networks.V4[1].IPAddress
	} else {
		ip = output[servernum-1].Networks.V4[0].IPAddress
	}

	return ip
}

func SSHDroplet(droplet string) {
	cmdstr := "ssh " + user + "@" + droplet
	cmd := exec.Command("bash", "-c", cmdstr)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	return
}

func main() {
	tokenSource := &TokenSource{
		AccessToken: pat,
	}

	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)

	output, err := DropletList(client)
	if err != nil {
		fmt.Println("There was an error: ", err)
		return
	}

	ip := OutputList(output)

	SSHDroplet(ip)

}
