package imapagent

import (
	"log"

	"github.com/emersion/go-imap/v2/imapclient"
)

func Connect(address string, options *imapclient.Options) (*imapclient.Client, error) {
	log.Printf("Connecting to IMAP server...")
	return imapclient.DialTLS(address, options)

}

func Login(email string, password string, c *imapclient.Client) error {
	log.Printf("Login to IMAP server...")
	if err := c.Login(email, password).Wait(); err != nil {
		log.Printf("Failed to login: %v", err)
		return err
	}
	log.Printf("Login successful")

	return nil
}
