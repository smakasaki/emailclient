package imapagent

import (
	"log"

	"github.com/emersion/go-imap/v2"
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

func FetchInboxMessages(c *imapclient.Client, limit uint32, offset uint32) ([]*imapclient.FetchMessageBuffer, error) {
	log.Printf("Fetching inbox messages...")
	mbox, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		return nil, err
	}

	seqSet := imap.SeqSetNum()
	seqSet.AddRange(mbox.NumMessages-limit-offset, mbox.NumMessages-offset)

	fetchOptions := &imap.FetchOptions{
		BodySection: []*imap.FetchItemBodySection{{Peek: true}},
	}

	messages, err := c.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		return nil, err
	}

	return messages, err
}
